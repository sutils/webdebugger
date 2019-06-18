package webdebugger

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

//Config is pojo to debuger configure
type Config struct {
	Hosts    []*ConfigHost            `json:"hosts"`
	Decorder []map[string]interface{} `json:"decorder"`
}

//ConfigHost is pojo to debuger configure
type ConfigHost struct {
	Host        string `json:"host"`
	IP          string `json:"ip"`
	Decorder    string `json:"decorder"`
	Forward     string `json:"forward"`
	DumpRequest int    `json:"dump_request"`
}

type remoteAddrConn struct {
	net.Conn
	Remote string
}

func (r *remoteAddrConn) RemoteAddr() net.Addr {
	return r
}

func (r *remoteAddrConn) Network() string {
	return "deugger"
}

func (r *remoteAddrConn) String() string {
	return r.Remote
}

//Debuger provider the web debuger suppported
type Debuger struct {
	*Config
	configLck sync.RWMutex
	closed    bool
	connQueue chan net.Conn
	server    *http.Server
	Decorder  DecorderCreator
}

//NewDebuger will return new Debuger
func NewDebuger(config *Config) (debuger *Debuger) {
	debuger = &Debuger{
		Config:    config,
		configLck: sync.RWMutex{},
		closed:    false,
		connQueue: make(chan net.Conn, 1000),
		Decorder:  DefaultDecorderCreator,
	}
	return
}

//Serve will start the http proxy server
func (d *Debuger) Serve() (err error) {
	d.server = &http.Server{Handler: d}
	err = d.server.Serve(d)
	return
}

//ProcConn will proc raw connection to uri
func (d *Debuger) ProcConn(uri string, raw net.Conn) (async bool, err error) {
	if d.closed {
		err = fmt.Errorf("Debuger is closed")
		return
	}
	var host *ConfigHost
	for _, h := range d.Hosts {
		if h.Host == uri || h.IP == uri {
			host = h
			break
		}
	}
	if host == nil { //direct
		DebugLog("Debuger start proc %v to %v by direct", raw, uri)
		var conn net.Conn
		conn, err = net.Dial("tcp", uri)
		if err == nil {
			go io.Copy(conn, raw)
			_, err = io.Copy(raw, conn)
		}
		return
	}
	InfoLog("Debuger start proc %v to %v by forwarding to %v", raw, uri, host.Forward)
	d.configLck.RLock()
	var decorderConfig map[string]interface{}
	for _, c := range d.Config.Decorder {
		if n, _ := c["name"].(string); n == host.Decorder {
			decorderConfig = c
			break
		}
	}
	d.configLck.RUnlock()
	decorder, err := d.Decorder(host.Decorder, decorderConfig)
	if err != nil {
		return
	}
	conn, err := decorder.Decord(host.Host, raw)
	if err != nil {
		return
	}
	d.connQueue <- &remoteAddrConn{Conn: conn, Remote: uri}
	async = true
	return
}

func (d *Debuger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var host *ConfigHost
	for _, h := range d.Hosts {
		if h.Host == r.RemoteAddr || h.IP == r.RemoteAddr {
			host = h
			break
		}
	}
	if host == nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, "%v is not configured", r.Host)
		return
	}
	target, err := url.Parse(host.Forward)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "parse %v fail with %v", host.Forward, err)
		return
	}
	if host.DumpRequest > 0 {
		buf := bytes.NewBuffer(nil)
		//
		fmt.Fprintln(buf, "---URL---")
		fmt.Fprintln(buf, "Host\t", r.Host)
		fmt.Fprintln(buf, "Path\t", r.URL.Path)
		fmt.Fprintln(buf, "RawPath\t", r.URL.RawPath)
		fmt.Fprintln(buf, "RawQuery\t", r.URL.RawQuery)
		fmt.Fprintln(buf, "User\t", r.URL.User)
		//
		fmt.Fprintln(buf, "\n---Header---")
		for k, v := range r.Header {
			fmt.Fprintln(buf, k, "\t", v)
		}
		//
		fmt.Fprintln(buf, "\n---Form---")
		for k, v := range r.Form {
			fmt.Fprintln(buf, k, "\t", v)
		}
		//
		fmt.Fprintln(buf, "---PostForm---")
		for k, v := range r.PostForm {
			fmt.Fprintln(buf, k, "\t", v)
		}
		fmt.Fprintf(buf, "\n\n\n")
		InfoLog("Debuger dump request:\n%v", string(buf.Bytes()))
	}
	r.Host = target.Host
	r.Header.Add("WebDebuggerProxy", "v1.0.0")
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ServeHTTP(w, r)
}

//Accept will accept on conn from queue
func (d *Debuger) Accept() (conn net.Conn, err error) {
	if !d.closed {
		conn = <-d.connQueue
	}
	if conn == nil {
		err = fmt.Errorf("Debuger is closed")
	}
	return
}

//Close will close all server
func (d *Debuger) Close() (err error) {
	if !d.closed {
		d.closed = true
		d.server.Close()
		close(d.connQueue)
	}
	return
}

// Addr returns the listener's network address.
func (d *Debuger) Addr() net.Addr {
	return d
}

//Network is impl to net.Addr
func (d *Debuger) Network() string {
	return "debuger"
}

func (d *Debuger) String() string {
	return "debuger"
}
