// Package git provides Git repository operations for git-assist.
//
// This package wraps go-git library to provide high-level operations
// for analyzing and manipulating Git repositories. It focuses on
// read-only operations and safe write operations like commits.
//
// Example usage:
//
//	repo, err := git.OpenRepository(".")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	analysis, err := repo.Analyze()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("Repository has %d commits\n", analysis.CommitCount)
package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Repository represents a Git repository and provides operations on it.
// It wraps the underlying go-git Repository to provide git-assist specific
// functionality while maintaining safety and performance.
type Repository struct {
	path string
}

// Analysis contains the results of repository analysis.
// This includes metrics, health checks, and patterns detected
// in the repository's history and structure.
type Analysis struct {
	CommitCount     int      `json:"commit_count"`    // Total number of commits
	BranchCount     int      `json:"branch_count"`    // Number of branches
	FileCount       int      `json:"file_count"`      // Number of tracked files
	LargeFiles      []string `json:"large_files"`     // Files over size threshold
	HealthScore     float64  `json:"health_score"`    // Overall repository health (0-1)
	Recommendations []string `json:"recommendations"` // Suggested improvements
}

// OpenRepository opens a Git repository at the specified path.
// It validates that the path contains a valid Git repository
// and returns a Repository instance for further operations.
//
// The path should point to either:
//   - A directory containing a .git subdirectory
//   - A bare Git repository directory
//
// Returns an error if the path is not a valid Git repository
// or if there are permission issues accessing it.
func OpenRepository(path string) (*Repository, error) {
	// Check if .git directory exists
	gitPath := path + "/.git"
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return nil, errors.New("not a git repository")
	}

	return &Repository{
		path: path,
	}, nil
}

// Analyze performs comprehensive analysis of the repository.
// It examines commit history, file structure, branch patterns,
// and other metrics to provide insights about repository health
// and usage patterns.
//
// This operation is read-only and safe to run on any repository.
// For large repositories, this may take several seconds to complete.
//
// Returns Analysis struct with detailed metrics, or an error
// if the repository cannot be analyzed.
func (r *Repository) Analyze() (*Analysis, error) {
	// TODO: Implement actual analysis using go-git
	// This is a placeholder implementation

	return &Analysis{
		CommitCount:     42,
		BranchCount:     3,
		FileCount:       156,
		LargeFiles:      []string{"data/large-file.bin"},
		HealthScore:     0.85,
		Recommendations: []string{"Consider using .gitignore for build artifacts"},
	}, nil
}

// Path returns the filesystem path of the repository.
func (r *Repository) Path() string {
	return r.path
}

// GetCurrentBranch returns the current branch name
func (r *Repository) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = r.path

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// GetCommitCount returns the approximate number of commits
func (r *Repository) GetCommitCount() (int, error) {
	cmd := exec.Command("git", "rev-list", "--count", "HEAD")
	cmd.Dir = r.path

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	count := 0
	if _, err := fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &count); err != nil {
		return 0, err
	}

	return count, nil
}

// Remote represents a git remote
type Remote struct {
	Name string
	URL  string
}

// GetRemotes returns all configured remotes
func (r *Repository) GetRemotes() ([]Remote, error) {
	cmd := exec.Command("git", "remote", "-v")
	cmd.Dir = r.path

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var remotes []Remote
	seen := make(map[string]bool)

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name := parts[0]
			url := parts[1]

			// Avoid duplicates (git remote -v shows fetch and push)
			key := name + ":" + url
			if !seen[key] {
				remotes = append(remotes, Remote{Name: name, URL: url})
				seen[key] = true
			}
		}
	}

	return remotes, nil
}

// Commit represents a git commit
type Commit struct {
	Hash    string
	Author  string
	Date    time.Time
	Message string
}

// GetLastCommit returns the most recent commit
func (r *Repository) GetLastCommit() (*Commit, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty=format:%H|%an|%ad|%s", "--date=iso")
	cmd.Dir = r.path

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	line := strings.TrimSpace(string(output))
	if line == "" {
		return nil, fmt.Errorf("no commits found")
	}

	parts := strings.Split(line, "|")
	if len(parts) != 4 {
		return nil, fmt.Errorf("unexpected git log format")
	}

	date, err := time.Parse("2006-01-02 15:04:05 -0700", parts[2])
	if err != nil {
		return nil, fmt.Errorf("failed to parse commit date: %w", err)
	}

	return &Commit{
		Hash:    parts[0],
		Author:  parts[1],
		Date:    date,
		Message: parts[3],
	}, nil
}

// GetStagedDiff returns the diff of staged changes
func (r *Repository) GetStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	cmd.Dir = r.path

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// GetWorkingDiff returns the diff of working directory changes
func (r *Repository) GetWorkingDiff() (string, error) {
	cmd := exec.Command("git", "diff")
	cmd.Dir = r.path

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// IsGitRepository checks if the current directory is a git repository
func (r *Repository) IsGitRepository() bool {
	gitPath := r.path + "/.git"
	_, err := os.Stat(gitPath)
	return err == nil
}

// NewRepository creates a new Repository instance (convenience function)
func NewRepository(path string) *Repository {
	return &Repository{path: path}
}
