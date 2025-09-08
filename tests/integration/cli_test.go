package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const binaryName = "git-assist"

// buildBinary builds the git-assist binary for testing
func buildBinary(t *testing.T) string {
	// Get the project root (two levels up from tests/integration)
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	// Build the binary
	binaryPath := filepath.Join(projectRoot, binaryName+"-test")
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/git-assist")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}

	return binaryPath
}

// createTestRepo creates a temporary git repository for testing
func createTestRepo(t *testing.T) string {
	tempDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Configure git user
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user.name: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user.email: %v", err)
	}

	return tempDir
}

// runCommand runs git-assist with the given arguments
func runCommand(t *testing.T, binaryPath, workDir string, args ...string) (string, string, error) {
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = workDir

	stdout, err := cmd.Output()
	stderr := ""
	if exitError, ok := err.(*exec.ExitError); ok {
		stderr = string(exitError.Stderr)
	}

	return string(stdout), stderr, err
}

func TestCLIHelp(t *testing.T) {
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	stdout, stderr, err := runCommand(t, binaryPath, ".", "--help")
	if err != nil {
		t.Fatalf("Help command failed: %v\nStderr: %s", err, stderr)
	}

	// Check that help output contains expected content
	if !strings.Contains(stdout, "git-assist - AI-powered Git assistant") {
		t.Error("Help output should contain main description")
	}

	if !strings.Contains(stdout, "Available Commands:") {
		t.Error("Help output should contain available commands")
	}

	// Check for main commands
	expectedCommands := []string{"analyze", "branch", "commit", "config", "history", "rebase"}
	for _, cmd := range expectedCommands {
		if !strings.Contains(stdout, cmd) {
			t.Errorf("Help output should contain command: %s", cmd)
		}
	}
}

func TestCLIVersion(t *testing.T) {
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	stdout, stderr, err := runCommand(t, binaryPath, ".", "--version")
	if err != nil {
		t.Fatalf("Version command failed: %v\nStderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "git-assist version") {
		t.Error("Version output should contain version information")
	}
}

func TestAnalyzeCommand(t *testing.T) {
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	repoPath := createTestRepo(t)

	// Add some files to analyze
	testFile := filepath.Join(repoPath, "main.go")
	content := `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Commit the file
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add files: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Run analyze command
	stdout, stderr, err := runCommand(t, binaryPath, repoPath, "analyze")
	if err != nil {
		t.Fatalf("Analyze command failed: %v\nStderr: %s", err, stderr)
	}

	// Check output contains expected analysis
	if !strings.Contains(stdout, "Repository Analysis") {
		t.Error("Analyze output should contain 'Repository Analysis'")
	}

	if !strings.Contains(stdout, "Go") {
		t.Error("Analyze output should detect Go language")
	}
}

func TestBranchCommand(t *testing.T) {
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	repoPath := createTestRepo(t)

	// Add initial commit
	testFile := filepath.Join(repoPath, "README.md")
	err := os.WriteFile(testFile, []byte("# Test Repo"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cmd := exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add files: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Run branch analyze command
	stdout, stderr, err := runCommand(t, binaryPath, repoPath, "branch", "analyze")
	if err != nil {
		t.Fatalf("Branch analyze command failed: %v\nStderr: %s", err, stderr)
	}

	// Check output contains expected branch analysis
	if !strings.Contains(stdout, "Branch Analysis") {
		t.Error("Branch output should contain 'Branch Analysis'")
	}

	if !strings.Contains(stdout, "Current Branch:") {
		t.Error("Branch output should show current branch")
	}
}

func TestHistoryCommand(t *testing.T) {
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	repoPath := createTestRepo(t)

	// Add multiple commits
	for i := 1; i <= 3; i++ {
		testFile := filepath.Join(repoPath, fmt.Sprintf("file%d.txt", i))
		content := fmt.Sprintf("Content of file %d", i)
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file %d: %v", i, err)
		}

		cmd := exec.Command("git", "add", ".")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add files: %v", err)
		}

		cmd = exec.Command("git", "commit", "-m", fmt.Sprintf("Add file%d.txt", i))
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to commit: %v", err)
		}
	}

	// Run history command
	stdout, stderr, err := runCommand(t, binaryPath, repoPath, "history", "--count", "2")
	if err != nil {
		t.Fatalf("History command failed: %v\nStderr: %s", err, stderr)
	}

	// Check output contains expected history
	if !strings.Contains(stdout, "Enhanced Git History") {
		t.Error("History output should contain 'Enhanced Git History'")
	}

	if !strings.Contains(stdout, "Add file") {
		t.Error("History output should contain commit messages")
	}
}

func TestConfigCommand(t *testing.T) {
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	repoPath := createTestRepo(t)

	// Initialize git-assist first
	_, _, err := runCommand(t, binaryPath, repoPath, "init")
	if err != nil {
		t.Fatalf("Init command failed: %v", err)
	}

	// Test config show (should work after init)
	stdout, stderr, err := runCommand(t, binaryPath, repoPath, "config")
	if err != nil {
		t.Fatalf("Config command failed: %v\nStderr: %s", err, stderr)
	}

	// Should show current configuration
	if !strings.Contains(stdout, "Configuration") {
		t.Error("Config output should contain 'Configuration'")
	}
}

func TestRebaseExplain(t *testing.T) {
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	repoPath := createTestRepo(t)

	// Run rebase explain command
	stdout, stderr, err := runCommand(t, binaryPath, repoPath, "rebase", "--explain")
	if err != nil {
		t.Fatalf("Rebase explain command failed: %v\nStderr: %s", err, stderr)
	}

	// Check output contains rebase explanation
	if !strings.Contains(stdout, "Understanding Git Rebase") {
		t.Error("Rebase explain output should contain explanation")
	}

	if !strings.Contains(stdout, "What is Rebase?") {
		t.Error("Rebase explain output should contain 'What is Rebase?'")
	}
}

func TestNonGitRepository(t *testing.T) {
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	nonGitDir := t.TempDir()

	// Commands that require git repo should fail gracefully
	_, stderr, err := runCommand(t, binaryPath, nonGitDir, "analyze")
	if err == nil {
		t.Error("Analyze command should fail in non-git directory")
	}

	if !strings.Contains(stderr, "not a git repository") {
		t.Error("Error message should indicate not a git repository")
	}
}

func TestQuietMode(t *testing.T) {
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	repoPath := createTestRepo(t)

	// Add a file and commit
	testFile := filepath.Join(repoPath, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cmd := exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add files: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Test commit")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Run command with quiet flag
	stdout, stderr, err := runCommand(t, binaryPath, repoPath, "analyze", "--quiet")
	if err != nil {
		t.Fatalf("Analyze command with quiet flag failed: %v\nStderr: %s", err, stderr)
	}

	// Quiet mode should produce less output
	normalStdout, _, _ := runCommand(t, binaryPath, repoPath, "analyze")

	// Check that quiet mode produces different (usually less) output
	if stdout == normalStdout {
		t.Error("Quiet mode should produce different output than normal mode")
	}
}
