package main

import (
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"

	"github.com/sutils/webdebugger"
)

var privoxyRunner *exec.Cmd
var privoxyPath = filepath.Join(execDir(), "privoxy")

func runPrivoxyNative(conf string) (err error) {
	privoxyRunner = exec.Command(privoxyPath, "--no-daemon", conf)
	privoxyRunner.Stderr = os.Stdout
	privoxyRunner.Stdout = os.Stderr
	err = privoxyRunner.Start()
	if err == nil {
		err = privoxyRunner.Wait()
	}
	privoxyRunner = nil
	return
}

var clientKillSignal chan os.Signal

func handlerClientKill() {
	clientKillSignal = make(chan os.Signal, 1000)
	// signal.Notify(clientKillSignal, os.Kill, os.Interrupt)
	signal.Notify(clientKillSignal)
	v := <-clientKillSignal
	webdebugger.WarnLog("Clien receive kill signal:%v", v)
	stopClient()
}
