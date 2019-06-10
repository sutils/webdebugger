package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

var privoxyRunner *exec.Cmd
var privoxyPath = filepath.Join(execDir(), "privoxy.exe")

func runPrivoxyNative(conf string) (err error) {
	privoxyRunner = exec.Command(privoxyPath, conf)
	privoxyRunner.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
	privoxyRunner.Stderr = os.Stdout
	privoxyRunner.Stdout = os.Stderr
	err = privoxyRunner.Start()
	if err == nil {
		err = privoxyRunner.Wait()
	}
	privoxyRunner = nil
	return
}
