//go:build !windows
// +build !windows

package ai

import (
	"os"
	"os/exec"
	"syscall"
)

// setupBackgroundProcess configures the command to run in background on Unix-like systems
func setupBackgroundProcess(cmd *exec.Cmd) {
	// On Unix-like systems, use setsid to create new session
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
}

// terminateProcess terminates a process gracefully on Unix-like systems
func terminateProcess(process *os.Process) error {
	return process.Signal(syscall.SIGTERM)
}
