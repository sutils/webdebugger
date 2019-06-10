package webdebugger

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
	"time"
)

func init() {
	SetLogLevel(LogLevelDebug)
}

func TestDebuger(t *testing.T) {
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
		config := map[string]interface{}{}
		ReadJSON("debuger_s_test.json", &config)
		certs := []map[string]interface{}{}
		data, _ := json.Marshal(config["certs"])
		json.Unmarshal(data, &certs)
		center := NewTLSCertCenter(certs...)
		server := &http.Server{Addr: config["listen"].(string)}
		server.Handler = center
		server.ListenAndServe()
	}()
	config := &Config{}
	ReadJSON("debuger_c_test.json", &config)
	debugger := NewDebuger(config)
	go debugger.Serve()
	time.Sleep(100 * time.Millisecond)
	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				a, b, _ := CreatePipeConn()
				_, err := debugger.ProcConn(addr, a)
				if err != nil {
					panic(err)
				}
				return b, nil
			},
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
