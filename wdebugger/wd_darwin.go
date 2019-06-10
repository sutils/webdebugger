package main

import (
	"os"
	"os/exec"
	"path/filepath"
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
