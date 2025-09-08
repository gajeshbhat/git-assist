package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// createTestRepo creates a temporary git repository for testing
func createTestRepo(t *testing.T) string {
	tempDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Configure git user (required for commits)
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

// addTestFile adds a file and commits it to the test repo
func addTestFile(t *testing.T, repoPath, filename, content string) {
	filePath := filepath.Join(repoPath, filename)

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Add file
	cmd := exec.Command("git", "add", filename)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	// Commit file
	cmd = exec.Command("git", "commit", "-m", "Add "+filename)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit file: %v", err)
	}
}

func TestNewRepository(t *testing.T) {
	testPath := "/test/path"
	repo := NewRepository(testPath)

	if repo == nil {
		t.Fatal("NewRepository() returned nil")
	}

	if repo.Path() != testPath {
		t.Errorf("Expected path '%s', got '%s'", testPath, repo.Path())
	}
}

func TestIsGitRepository(t *testing.T) {
	// Test with actual git repository
	repoPath := createTestRepo(t)
	repo := NewRepository(repoPath)

	if !repo.IsGitRepository() {
		t.Error("IsGitRepository() should return true for valid git repo")
	}

	// Test with non-git directory
	nonGitPath := t.TempDir()
	nonGitRepo := NewRepository(nonGitPath)

	if nonGitRepo.IsGitRepository() {
		t.Error("IsGitRepository() should return false for non-git directory")
	}
}

func TestGetCurrentBranch(t *testing.T) {
	repoPath := createTestRepo(t)
	repo := NewRepository(repoPath)

	// Add initial commit (required for branch to exist)
	addTestFile(t, repoPath, "README.md", "# Test Repo")

	branch, err := repo.GetCurrentBranch()
	if err != nil {
		t.Fatalf("GetCurrentBranch() failed: %v", err)
	}

	// Default branch should be main or master
	if branch != "main" && branch != "master" {
		t.Errorf("Expected branch 'main' or 'master', got '%s'", branch)
	}
}

func TestGetCommitCount(t *testing.T) {
	repoPath := createTestRepo(t)
	repo := NewRepository(repoPath)

	// Initially should have 0 commits
	count, err := repo.GetCommitCount()
	if err == nil && count != 0 {
		t.Errorf("Expected 0 commits in empty repo, got %d", count)
	}

	// Add a commit
	addTestFile(t, repoPath, "file1.txt", "content1")

	count, err = repo.GetCommitCount()
	if err != nil {
		t.Fatalf("GetCommitCount() failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 commit, got %d", count)
	}

	// Add another commit
	addTestFile(t, repoPath, "file2.txt", "content2")

	count, err = repo.GetCommitCount()
	if err != nil {
		t.Fatalf("GetCommitCount() failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 commits, got %d", count)
	}
}

func TestGetRemotes(t *testing.T) {
	repoPath := createTestRepo(t)
	repo := NewRepository(repoPath)

	// Initially should have no remotes
	remotes, err := repo.GetRemotes()
	if err != nil {
		t.Fatalf("GetRemotes() failed: %v", err)
	}

	if len(remotes) != 0 {
		t.Errorf("Expected 0 remotes, got %d", len(remotes))
	}

	// Add a remote
	cmd := exec.Command("git", "remote", "add", "origin", "https://github.com/test/repo.git")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add remote: %v", err)
	}

	remotes, err = repo.GetRemotes()
	if err != nil {
		t.Fatalf("GetRemotes() failed: %v", err)
	}

	if len(remotes) != 1 {
		t.Errorf("Expected 1 remote, got %d", len(remotes))
	}

	if remotes[0].Name != "origin" {
		t.Errorf("Expected remote name 'origin', got '%s'", remotes[0].Name)
	}

	if remotes[0].URL != "https://github.com/test/repo.git" {
		t.Errorf("Expected remote URL 'https://github.com/test/repo.git', got '%s'", remotes[0].URL)
	}
}

func TestGetLastCommit(t *testing.T) {
	repoPath := createTestRepo(t)
	repo := NewRepository(repoPath)

	// Should fail with no commits
	_, err := repo.GetLastCommit()
	if err == nil {
		t.Error("GetLastCommit() should fail with no commits")
	}

	// Add a commit
	addTestFile(t, repoPath, "test.txt", "test content")

	commit, err := repo.GetLastCommit()
	if err != nil {
		t.Fatalf("GetLastCommit() failed: %v", err)
	}

	if commit == nil {
		t.Fatal("GetLastCommit() returned nil commit")
	}

	if commit.Hash == "" {
		t.Error("Commit hash should not be empty")
	}

	if commit.Author != "Test User" {
		t.Errorf("Expected author 'Test User', got '%s'", commit.Author)
	}

	if commit.Message != "Add test.txt" {
		t.Errorf("Expected message 'Add test.txt', got '%s'", commit.Message)
	}

	if commit.Date.IsZero() {
		t.Error("Commit date should not be zero")
	}
}

func TestGetStagedDiff(t *testing.T) {
	repoPath := createTestRepo(t)
	repo := NewRepository(repoPath)

	// Initially should have no staged changes
	diff, err := repo.GetStagedDiff()
	if err != nil {
		t.Fatalf("GetStagedDiff() failed: %v", err)
	}

	if strings.TrimSpace(diff) != "" {
		t.Error("Expected empty staged diff")
	}

	// Add a file but don't commit
	filePath := filepath.Join(repoPath, "staged.txt")
	err = os.WriteFile(filePath, []byte("staged content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Stage the file
	cmd := exec.Command("git", "add", "staged.txt")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to stage file: %v", err)
	}

	// Now should have staged changes
	diff, err = repo.GetStagedDiff()
	if err != nil {
		t.Fatalf("GetStagedDiff() failed: %v", err)
	}

	if strings.TrimSpace(diff) == "" {
		t.Error("Expected non-empty staged diff")
	}

	if !strings.Contains(diff, "staged content") {
		t.Error("Staged diff should contain file content")
	}
}

func TestGetWorkingDiff(t *testing.T) {
	repoPath := createTestRepo(t)
	repo := NewRepository(repoPath)

	// Add initial commit
	addTestFile(t, repoPath, "existing.txt", "original content")

	// Initially should have no working changes
	diff, err := repo.GetWorkingDiff()
	if err != nil {
		t.Fatalf("GetWorkingDiff() failed: %v", err)
	}

	if strings.TrimSpace(diff) != "" {
		t.Error("Expected empty working diff")
	}

	// Modify the file
	filePath := filepath.Join(repoPath, "existing.txt")
	err = os.WriteFile(filePath, []byte("modified content"), 0644)
	if err != nil {
		t.Fatalf("Failed to modify file: %v", err)
	}

	// Now should have working changes
	diff, err = repo.GetWorkingDiff()
	if err != nil {
		t.Fatalf("GetWorkingDiff() failed: %v", err)
	}

	if strings.TrimSpace(diff) == "" {
		t.Error("Expected non-empty working diff")
	}

	if !strings.Contains(diff, "modified content") {
		t.Error("Working diff should contain modified content")
	}
}

func TestOpenRepository(t *testing.T) {
	repoPath := createTestRepo(t)

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository() failed: %v", err)
	}

	if repo == nil {
		t.Fatal("OpenRepository() returned nil")
	}

	if repo.Path() != repoPath {
		t.Errorf("Expected path '%s', got '%s'", repoPath, repo.Path())
	}

	// Test with non-git directory
	nonGitPath := t.TempDir()
	_, err = OpenRepository(nonGitPath)
	if err == nil {
		t.Error("OpenRepository() should fail for non-git directory")
	}
}
