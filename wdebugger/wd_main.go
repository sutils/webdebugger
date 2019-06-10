package main

import (
	"flag"
	"log"
	"os"
)

var argConf string
var argRunServer bool
var argRunProxy bool
var exitf = os.Exit

func init() {
	flag.StringVar(&argConf, "f", "./wdebugger.json", "the web debugger configure file")
	flag.BoolVar(&argRunServer, "s", false, "start cert center server")
	flag.BoolVar(&argRunProxy, "p", true, "start web debuger proxy server")
	flag.Parse()
}

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)
	log.SetOutput(os.Stdout)
	if argRunServer {
		startServer(argConf)
	} else if argRunProxy {
		startProxy(argConf)
	} else {
		flag.Usage()
	}
}
