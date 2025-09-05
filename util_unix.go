//go:build linux || darwin

package main

import (
	"os/exec"
	"syscall"
)

func initCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{}
}
