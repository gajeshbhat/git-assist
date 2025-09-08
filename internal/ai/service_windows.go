//go:build windows
// +build windows

package ai

import (
	"os"
	"os/exec"
)

// setupBackgroundProcess configures the command to run in background on Windows
func setupBackgroundProcess(cmd *exec.Cmd) {
	// Windows doesn't need special process attributes for this use case
	// The process will run in the background by default
}

// terminateProcess terminates a process on Windows
func terminateProcess(process *os.Process) error {
	return process.Kill()
}
