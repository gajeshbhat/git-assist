// Package ai/installer handles automatic installation of AI dependencies
// across different operating systems (macOS, Linux, Windows)
package ai

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// SystemInfo contains information about the current system
type SystemInfo struct {
	OS             string // "darwin", "linux", "windows"
	Arch           string // "amd64", "arm64"
	PackageManager string // "brew", "apt", "yum", "choco", "manual"
	HasAdmin       bool   // Whether user has admin privileges
}

// InstallationMethod represents how to install Ollama on this system
type InstallationMethod struct {
	Method        string   // "package_manager", "script", "manual"
	Commands      []string // Commands to run
	Description   string   // Human-readable description
	RequiresAdmin bool     // Whether admin privileges are needed
}

// SystemDetector detects the current system and available installation methods
type SystemDetector struct{}

// NewSystemDetector creates a new system detector
func NewSystemDetector() *SystemDetector {
	return &SystemDetector{}
}

// DetectSystem analyzes the current system and returns system information
func (sd *SystemDetector) DetectSystem() (*SystemInfo, error) {
	info := &SystemInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	// Detect package manager based on OS
	switch info.OS {
	case "darwin": // macOS
		if sd.commandExists("brew") {
			info.PackageManager = "brew"
		} else {
			info.PackageManager = "manual"
		}
	case "linux":
		if sd.commandExists("apt") {
			info.PackageManager = "apt"
		} else if sd.commandExists("yum") {
			info.PackageManager = "yum"
		} else if sd.commandExists("dnf") {
			info.PackageManager = "dnf"
		} else if sd.commandExists("pacman") {
			info.PackageManager = "pacman"
		} else {
			info.PackageManager = "manual"
		}
	case "windows":
		if sd.commandExists("choco") {
			info.PackageManager = "choco"
		} else if sd.commandExists("winget") {
			info.PackageManager = "winget"
		} else {
			info.PackageManager = "manual"
		}
	default:
		info.PackageManager = "manual"
	}

	// Check admin privileges (simplified check)
	info.HasAdmin = sd.hasAdminPrivileges()

	return info, nil
}

// commandExists checks if a command is available in PATH
func (sd *SystemDetector) commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// hasAdminPrivileges checks if the user has admin privileges
func (sd *SystemDetector) hasAdminPrivileges() bool {
	switch runtime.GOOS {
	case "windows":
		// On Windows, try to run a command that requires admin
		cmd := exec.Command("net", "session")
		err := cmd.Run()
		return err == nil
	case "darwin", "linux":
		// On Unix-like systems, check if we can sudo
		cmd := exec.Command("sudo", "-n", "true")
		err := cmd.Run()
		return err == nil
	default:
		return false
	}
}

// GetInstallationMethods returns available installation methods for Ollama
func (sd *SystemDetector) GetInstallationMethods(info *SystemInfo) []InstallationMethod {
	var methods []InstallationMethod

	switch info.OS {
	case "darwin": // macOS
		if info.PackageManager == "brew" {
			methods = append(methods, InstallationMethod{
				Method:        "package_manager",
				Commands:      []string{"brew install ollama"},
				Description:   "Install using Homebrew (recommended)",
				RequiresAdmin: false,
			})
		}

		// Always offer manual installation
		methods = append(methods, InstallationMethod{
			Method:        "manual",
			Commands:      []string{"curl -fsSL https://ollama.ai/install.sh | sh"},
			Description:   "Install using official script",
			RequiresAdmin: false,
		})

	case "linux":
		// Package manager installation
		switch info.PackageManager {
		case "apt":
			methods = append(methods, InstallationMethod{
				Method: "package_manager",
				Commands: []string{
					"curl -fsSL https://ollama.ai/install.sh | sh",
				},
				Description:   "Install using official script (recommended)",
				RequiresAdmin: true,
			})
		case "yum", "dnf":
			methods = append(methods, InstallationMethod{
				Method: "package_manager",
				Commands: []string{
					"curl -fsSL https://ollama.ai/install.sh | sh",
				},
				Description:   "Install using official script (recommended)",
				RequiresAdmin: true,
			})
		}

		// Manual installation
		methods = append(methods, InstallationMethod{
			Method:        "manual",
			Commands:      []string{"curl -fsSL https://ollama.ai/install.sh | sh"},
			Description:   "Install using official script",
			RequiresAdmin: true,
		})

	case "windows":
		if info.PackageManager == "choco" {
			methods = append(methods, InstallationMethod{
				Method:        "package_manager",
				Commands:      []string{"choco install ollama"},
				Description:   "Install using Chocolatey",
				RequiresAdmin: true,
			})
		}

		if info.PackageManager == "winget" {
			methods = append(methods, InstallationMethod{
				Method:        "package_manager",
				Commands:      []string{"winget install Ollama.Ollama"},
				Description:   "Install using Windows Package Manager",
				RequiresAdmin: false,
			})
		}

		// Manual installation
		methods = append(methods, InstallationMethod{
			Method:        "manual",
			Commands:      []string{}, // Will be handled specially
			Description:   "Download and install from https://ollama.ai/download",
			RequiresAdmin: false,
		})
	}

	return methods
}

// OllamaInstaller handles Ollama installation
type OllamaInstaller struct {
	detector *SystemDetector
}

// NewOllamaInstaller creates a new Ollama installer
func NewOllamaInstaller() *OllamaInstaller {
	return &OllamaInstaller{
		detector: NewSystemDetector(),
	}
}

// IsOllamaInstalled checks if Ollama is already installed
func (oi *OllamaInstaller) IsOllamaInstalled() bool {
	return oi.detector.commandExists("ollama")
}

// InstallOllama attempts to install Ollama automatically
func (oi *OllamaInstaller) InstallOllama(method InstallationMethod) error {
	if oi.IsOllamaInstalled() {
		return fmt.Errorf("ollama is already installed")
	}

	switch method.Method {
	case "package_manager":
		return oi.runCommands(method.Commands, method.RequiresAdmin)
	case "manual":
		if runtime.GOOS == "windows" {
			return fmt.Errorf("manual installation on Windows requires downloading from https://ollama.ai/download")
		}
		return oi.runCommands(method.Commands, method.RequiresAdmin)
	default:
		return fmt.Errorf("unsupported installation method: %s", method.Method)
	}
}

// runCommands executes a series of commands
func (oi *OllamaInstaller) runCommands(commands []string, requiresAdmin bool) error {
	for _, cmdStr := range commands {
		parts := strings.Fields(cmdStr)
		if len(parts) == 0 {
			continue
		}

		var cmd *exec.Cmd
		if requiresAdmin && runtime.GOOS != "windows" {
			// Use sudo on Unix-like systems
			args := append([]string{"-S"}, parts...)
			cmd = exec.Command("sudo", args...)
		} else {
			cmd = exec.Command(parts[0], parts[1:]...)
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command failed: %s (%w)", cmdStr, err)
		}
	}

	return nil
}

// GetInstallationInstructions returns human-readable installation instructions
func (oi *OllamaInstaller) GetInstallationInstructions() ([]InstallationMethod, error) {
	info, err := oi.detector.DetectSystem()
	if err != nil {
		return nil, fmt.Errorf("failed to detect system: %w", err)
	}

	return oi.detector.GetInstallationMethods(info), nil
}
