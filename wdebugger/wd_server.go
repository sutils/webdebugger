package main

import (
	"net/http"
	"path/filepath"

	"github.com/sutils/webdebugger"
)

//ServerConf is pojo for server configure
type ServerConf struct {
	Listen   string                   `json:"listen"`
	Certs    []map[string]interface{} `json:"certs"`
	LogLevel int                      `json:"log"`
}

var serverConf string
var serverConfDir string
var certCenter *webdebugger.TLSCertCenter

func startServer(c string) (err error) {
	conf := &ServerConf{}
	err = webdebugger.ReadJSON(c, &conf)
	if err != nil {
		webdebugger.ErrorLog("Server read configure from %v fail with %v", c, err)
		exitf(1)
		return
	}
	serverConf = c
	serverConfDir = filepath.Dir(serverConf)
	webdebugger.SetLogLevel(conf.LogLevel)
	certCenter = webdebugger.NewTLSCertCenter(conf.Certs...)
	http.ListenAndServe(conf.Listen, certCenter)
	return
}
