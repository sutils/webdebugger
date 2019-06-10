package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

func init() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)
	log.SetOutput(os.Stdout)
	privoxyPath = "./privoxy"
	switch runtime.GOOS {
	case "darwin":
		exec.Command("cp", "-f", "../privoxy-Darwin", "privoxy").CombinedOutput()
	case "linux":
		exec.Command("cp", "-f", "../privoxy-Linux", "privoxy").CombinedOutput()
	default:
		exec.Command("powershell", "cp", "..\\privoxy-Win.exe", "privoxy.exe").CombinedOutput()
	}
}

func TestMain(t *testing.T) {
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("test server receive from %v\n", r.RemoteAddr)
			fmt.Fprintf(w, "ok")
		})
		server := &http.Server{Addr: ":10021", Handler: mux}
		server.ListenAndServe()
	}()
	go func() {
		startServer("wdebugger-s.json")
	}()
	go func() {
		startProxy("wdebugger-c.json")
	}()
	time.Sleep(time.Second)
	proxyURL, _ := url.Parse("http://127.0.0.1:10201")
	client := &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := doGet(client, "https://wdebugger.snows.io")
	if err != nil || resp != "ok" {
		t.Errorf("err:%v,resp:%v", err, resp)
		return
	}
}

func doGet(client *http.Client, url string) (data string, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err == nil {
		var resp *http.Response
		resp, err = client.Do(req)
		if err == nil {
			var d []byte
			d, err = ioutil.ReadAll(resp.Body)
			data = string(d)
		}
	}
	return
}
