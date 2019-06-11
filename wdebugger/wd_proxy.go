package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sutils/webdebugger"
)

var proxyConf string
var proxyConfDir string
var proxyServer *webdebugger.SocksProxy
var debugger *webdebugger.Debuger

type proxyConfig struct {
	Socks5 string `json:"socks5"`
	HTTP   string `json:"http"`
}
type clientConfig struct {
	webdebugger.Config
	Proxy    proxyConfig `json:"proxy"`
	LogLevel int         `json:"log"`
}

func startProxy(c string) (err error) {
	conf := &clientConfig{}
	err = webdebugger.ReadJSON(c, &conf)
	if err != nil {
		webdebugger.ErrorLog("Client read configure fail with %v", err)
		exitf(1)
		return
	}
	if len(conf.Proxy.Socks5) < 1 {
		webdebugger.ErrorLog("Client proxy.socks5 is required")
		exitf(1)
		return
	}
	proxyConf = c
	proxyConfDir = filepath.Dir(proxyConf)
	webdebugger.SetLogLevel(conf.LogLevel)
	webdebugger.InfoLog("Client using config from %v", c)
	debugger = webdebugger.NewDebuger(&conf.Config)
	proxyServer = webdebugger.NewSocksProxy()
	proxyServer.ProcConn = debugger.ProcConn
	err = proxyServer.Listen(conf.Proxy.Socks5)
	if err != nil {
		webdebugger.ErrorLog("Client start proxy server fail with %v", err)
		exitf(1)
		return
	}
	// writeRuntimeVar()
	wait := sync.WaitGroup{}
	if len(conf.Proxy.HTTP) > 0 {
		wait.Add(1)
		proxyServer.HTTPUpstream = conf.Proxy.HTTP
		go func() {
			xerr := runPrivoxy(conf.Proxy.HTTP)
			webdebugger.WarnLog("Client the privoxy on %v is stopped by %v", conf.Proxy.HTTP, xerr)
			wait.Done()
		}()
	}
	wait.Add(1)
	go func() {
		debugger.Serve()
		wait.Done()
	}()
	go handlerClientKill()
	proxyServer.Run()
	webdebugger.InfoLog("Client all listener is stopped")
	wait.Wait()
	return
}

func stopClient() {
	webdebugger.InfoLog("Client stopping client listener")
	if proxyServer != nil {
		proxyServer.Close()
	}
	if privoxyRunner != nil && privoxyRunner.Process != nil {
		privoxyRunner.Process.Kill()
	}
	if debugger != nil {
		debugger.Close()
	}
}

const (
	//PrivoxyTmpl is privoxy template
	PrivoxyTmpl = `
listen-address {http}
toggle  1
enable-remote-toggle 1
enable-remote-http-toggle 1
enable-edit-actions 0
enforce-blocks 0
buffer-limit 4096
forwarded-connect-retries  0
accept-intercepted-requests 0
allow-cgi-request-crunching 0
split-large-forms 0
keep-alive-timeout 5
socket-timeout 60

forward-socks5 / {socks5} .
forward         192.168.*.*/     .
forward         10.*.*.*/        .
forward         127.*.*.*/       .

	`
)

func writePrivoxyConf(confFile, httpAddr, socksAddr string) (err error) {
	data := PrivoxyTmpl
	data = strings.Replace(data, "{http}", httpAddr, 1)
	data = strings.Replace(data, "{socks5}", socksAddr, 1)
	err = ioutil.WriteFile(confFile, []byte(data), os.ModePerm)
	return
}

func runPrivoxy(httpAddr string) (err error) {
	proxyServerParts := strings.SplitN(proxyServer.Addr().String(), ":", -1)
	socksAddr := fmt.Sprintf("127.0.0.1:%v", proxyServerParts[len(proxyServerParts)-1])
	webdebugger.InfoLog("Client start privoxy by listening http proxy on %v and forwarding to %v", httpAddr, socksAddr)
	confFile := filepath.Join(workDir, "privoxy.conf")
	err = writePrivoxyConf(confFile, httpAddr, socksAddr)
	if err != nil {
		webdebugger.WarnLog("Client save privoxy config to %v fail with %v", confFile, err)
		return
	}
	err = runPrivoxyNative(confFile)
	return
}
