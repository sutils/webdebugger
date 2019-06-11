package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/sutils/webdebugger"
	"golang.org/x/sys/windows"
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

func handlerClientKill() {
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	setConsoleCtrlHandler := kernel32.NewProc("SetConsoleCtrlHandler")
	setConsoleCtrlHandler.Call(
		syscall.NewCallback(func(controlType uint) uint {
			webdebugger.WarnLog("Clien receive kill signal:%v", controlType)
			stopClient()
			return 0
		}), 1)
}
