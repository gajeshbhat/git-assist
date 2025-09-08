// Package ai/service handles background service management for Ollama
package ai

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ServiceManager manages the Ollama background service
type ServiceManager struct {
	endpoint string
	client   *http.Client
}

// NewServiceManager creates a new service manager
func NewServiceManager(endpoint string) *ServiceManager {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}

	return &ServiceManager{
		endpoint: endpoint,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// ServiceStatus represents the status of the Ollama service
type ServiceStatus struct {
	Running  bool   `json:"running"`
	PID      int    `json:"pid,omitempty"`
	Endpoint string `json:"endpoint"`
	Uptime   string `json:"uptime,omitempty"`
	Error    string `json:"error,omitempty"`
}

// IsRunning checks if Ollama service is running
func (sm *ServiceManager) IsRunning() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", sm.endpoint+"/api/tags", nil)
	if err != nil {
		return false
	}

	resp, err := sm.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// GetStatus returns detailed status information about the Ollama service
func (sm *ServiceManager) GetStatus() ServiceStatus {
	status := ServiceStatus{
		Endpoint: sm.endpoint,
		Running:  sm.IsRunning(),
	}

	if !status.Running {
		status.Error = "Service not responding"
		return status
	}

	// Try to get process information
	if pid := sm.findOllamaPID(); pid > 0 {
		status.PID = pid
		status.Uptime = sm.getProcessUptime(pid)
	}

	return status
}

// StartService starts the Ollama service in the background
func (sm *ServiceManager) StartService() error {
	if sm.IsRunning() {
		return fmt.Errorf("ollama service is already running")
	}

	// Check if ollama command exists
	if _, err := exec.LookPath("ollama"); err != nil {
		return fmt.Errorf("ollama is not installed. Run 'git-assist config --install-ollama' first")
	}

	// Start ollama serve in the background
	cmd := exec.Command("ollama", "serve")

	// Set up the command to run in background
	setupBackgroundProcess(cmd)

	// Redirect output to prevent hanging
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ollama service: %w", err)
	}

	// Wait a moment for the service to start
	time.Sleep(2 * time.Second)

	// Verify it's running
	if !sm.IsRunning() {
		return fmt.Errorf("ollama service failed to start properly")
	}

	return nil
}

// StopService stops the Ollama service
func (sm *ServiceManager) StopService() error {
	if !sm.IsRunning() {
		return fmt.Errorf("ollama service is not running")
	}

	pid := sm.findOllamaPID()
	if pid <= 0 {
		return fmt.Errorf("could not find ollama process")
	}

	// Find the process and terminate it
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("could not find process %d: %w", pid, err)
	}

	// Send termination signal
	err = terminateProcess(process)

	if err != nil {
		return fmt.Errorf("failed to stop ollama service: %w", err)
	}

	// Wait for it to stop
	for i := 0; i < 10; i++ {
		if !sm.IsRunning() {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("ollama service did not stop within timeout")
}

// RestartService restarts the Ollama service
func (sm *ServiceManager) RestartService() error {
	if sm.IsRunning() {
		if err := sm.StopService(); err != nil {
			return fmt.Errorf("failed to stop service: %w", err)
		}
	}

	return sm.StartService()
}

// findOllamaPID finds the process ID of the running Ollama service
func (sm *ServiceManager) findOllamaPID() int {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Use tasklist on Windows
		cmd = exec.Command("tasklist", "/FI", "IMAGENAME eq ollama.exe", "/FO", "CSV", "/NH")
	case "darwin", "linux":
		// Use pgrep on Unix-like systems
		cmd = exec.Command("pgrep", "-f", "ollama serve")
	default:
		return 0
	}

	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	if runtime.GOOS == "windows" {
		// Parse Windows tasklist output
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "ollama.exe") {
				parts := strings.Split(line, ",")
				if len(parts) >= 2 {
					pidStr := strings.Trim(parts[1], `"`)
					if pid, err := strconv.Atoi(pidStr); err == nil {
						return pid
					}
				}
			}
		}
	} else {
		// Parse pgrep output (just the PID)
		pidStr := strings.TrimSpace(string(output))
		if pid, err := strconv.Atoi(pidStr); err == nil {
			return pid
		}
	}

	return 0
}

// getProcessUptime gets the uptime of a process (simplified)
func (sm *ServiceManager) getProcessUptime(pid int) string {
	// This is a simplified implementation
	// In a real implementation, you'd query the process start time
	// and calculate the difference from now
	return "unknown"
}

// AutoStartService starts the service automatically if it's not running
func (sm *ServiceManager) AutoStartService() error {
	if sm.IsRunning() {
		return nil // Already running
	}

	return sm.StartService()
}

// WaitForService waits for the service to become available
func (sm *ServiceManager) WaitForService(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for ollama service to start")
		case <-ticker.C:
			if sm.IsRunning() {
				return nil
			}
		}
	}
}
