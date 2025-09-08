// Package repository/context provides context analysis for AI-powered operations
package repository

import (
	"fmt"
	"path/filepath"
	"strings"
)

// CommitContext contains contextual information for generating commit messages
type CommitContext struct {
	Repository   *RepositoryIndex `json:"repository"`
	ChangedFiles []*ChangedFile   `json:"changed_files"`
	ChangeType   string           `json:"change_type"` // "feature", "fix", "refactor", etc.
	Scope        string           `json:"scope"`       // "frontend", "api", "auth", etc.
	Impact       string           `json:"impact"`      // "major", "minor", "patch"
	RelatedFiles []*FileInfo      `json:"related_files"`
	Suggestions  []string         `json:"suggestions"`
	Context      string           `json:"context"` // Human-readable context
}

// ChangedFile represents a file that was modified in the current commit
type ChangedFile struct {
	Path         string   `json:"path"`
	Status       string   `json:"status"` // "added", "modified", "deleted", "renamed"
	Language     string   `json:"language"`
	Type         string   `json:"type"`
	LinesAdded   int      `json:"lines_added"`
	LinesDeleted int      `json:"lines_deleted"`
	Functions    []string `json:"functions,omitempty"`    // Functions that were modified
	Dependencies []string `json:"dependencies,omitempty"` // Dependencies that were added/removed
}

// ContextAnalyzer analyzes repository context for AI operations
type ContextAnalyzer struct {
	indexer *Indexer
	index   *RepositoryIndex
}

// NewContextAnalyzer creates a new context analyzer
func NewContextAnalyzer(rootPath string) *ContextAnalyzer {
	return &ContextAnalyzer{
		indexer: NewIndexer(rootPath),
	}
}

// AnalyzeCommitContext analyzes the context for a commit based on the diff
func (ca *ContextAnalyzer) AnalyzeCommitContext(diff string) (*CommitContext, error) {
	// Load or create repository index
	var err error
	ca.index, err = ca.indexer.LoadIndex()
	if err != nil {
		// If index doesn't exist or is corrupted, create a new one
		ca.index, err = ca.indexer.IndexRepository()
		if err != nil {
			return nil, fmt.Errorf("failed to create repository index: %w", err)
		}
	}

	// Check if index is stale and refresh if needed
	if stale, _ := ca.indexer.IsIndexStale(); stale {
		ca.index, err = ca.indexer.IndexRepository()
		if err != nil {
			// Non-fatal error, continue with existing index
			fmt.Printf("Warning: Failed to refresh index: %v\n", err)
		}
	}

	// Parse the diff to extract changed files
	changedFiles, err := ca.parseDiff(diff)
	if err != nil {
		return nil, fmt.Errorf("failed to parse diff: %w", err)
	}

	// Analyze the changes
	context := &CommitContext{
		Repository:   ca.index,
		ChangedFiles: changedFiles,
	}

	// Determine change type and scope
	ca.analyzeChangeType(context)
	ca.analyzeScope(context)
	ca.analyzeImpact(context)

	// Find related files
	ca.findRelatedFiles(context)

	// Generate suggestions and context
	ca.generateSuggestions(context)
	ca.generateContextDescription(context)

	return context, nil
}

// parseDiff parses a git diff and extracts changed file information
func (ca *ContextAnalyzer) parseDiff(diff string) ([]*ChangedFile, error) {
	lines := strings.Split(diff, "\n")
	var changedFiles []*ChangedFile
	var currentFile *ChangedFile

	for _, line := range lines {
		// File header: diff --git a/path b/path
		if strings.HasPrefix(line, "diff --git") {
			if currentFile != nil {
				changedFiles = append(changedFiles, currentFile)
			}

			// Extract file path
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				path := strings.TrimPrefix(parts[2], "a/")
				currentFile = &ChangedFile{
					Path:     path,
					Status:   "modified", // Default, will be updated
					Language: detectLanguage(path),
					Type:     detectFileType(path),
				}
			}
		}

		// File status indicators
		if currentFile != nil {
			if strings.HasPrefix(line, "new file mode") {
				currentFile.Status = "added"
			} else if strings.HasPrefix(line, "deleted file mode") {
				currentFile.Status = "deleted"
			} else if strings.HasPrefix(line, "rename from") {
				currentFile.Status = "renamed"
			}

			// Count line changes
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				currentFile.LinesAdded++
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
				currentFile.LinesDeleted++
			}
		}
	}

	// Add the last file
	if currentFile != nil {
		changedFiles = append(changedFiles, currentFile)
	}

	return changedFiles, nil
}

// analyzeChangeType determines the type of change (feature, fix, refactor, etc.)
func (ca *ContextAnalyzer) analyzeChangeType(context *CommitContext) {
	// Analyze file types and changes to determine change type
	hasNewFiles := false
	hasDeletedFiles := false
	hasTestFiles := false
	hasConfigFiles := false
	hasDocFiles := false

	for _, file := range context.ChangedFiles {
		switch file.Status {
		case "added":
			hasNewFiles = true
		case "deleted":
			hasDeletedFiles = true
		}

		switch file.Type {
		case "test":
			hasTestFiles = true
		case "config":
			hasConfigFiles = true
		case "docs":
			hasDocFiles = true
		}
	}

	// Determine change type based on patterns
	if hasDocFiles && len(context.ChangedFiles) == 1 {
		context.ChangeType = "docs"
	} else if hasTestFiles && !hasNewFiles {
		context.ChangeType = "test"
	} else if hasConfigFiles {
		context.ChangeType = "config"
	} else if hasNewFiles {
		context.ChangeType = "feat"
	} else if hasDeletedFiles {
		context.ChangeType = "refactor"
	} else {
		// Default to fix for modifications
		context.ChangeType = "fix"
	}
}

// analyzeScope determines the scope of changes (frontend, backend, api, etc.)
func (ca *ContextAnalyzer) analyzeScope(context *CommitContext) {
	scopeMap := make(map[string]int)

	for _, file := range context.ChangedFiles {
		// Analyze file path to determine scope
		pathParts := strings.Split(file.Path, "/")

		// Common scope patterns
		for _, part := range pathParts {
			part = strings.ToLower(part)

			switch {
			case strings.Contains(part, "frontend") || strings.Contains(part, "ui") ||
				strings.Contains(part, "client") || file.Language == "JavaScript" ||
				file.Language == "TypeScript" || file.Language == "Vue":
				scopeMap["frontend"]++

			case strings.Contains(part, "backend") || strings.Contains(part, "server") ||
				strings.Contains(part, "api") || file.Language == "Go" ||
				file.Language == "Python" || file.Language == "Java":
				scopeMap["backend"]++

			case strings.Contains(part, "auth") || strings.Contains(part, "login") ||
				strings.Contains(part, "security"):
				scopeMap["auth"]++

			case strings.Contains(part, "db") || strings.Contains(part, "database") ||
				strings.Contains(part, "migration") || file.Language == "SQL":
				scopeMap["database"]++

			case strings.Contains(part, "test") || file.Type == "test":
				scopeMap["test"]++

			case strings.Contains(part, "doc") || file.Type == "docs":
				scopeMap["docs"]++

			case strings.Contains(part, "config") || file.Type == "config":
				scopeMap["config"]++
			}
		}
	}

	// Find the most common scope
	maxCount := 0
	for scope, count := range scopeMap {
		if count > maxCount {
			maxCount = count
			context.Scope = scope
		}
	}

	// If no specific scope found, use the main language or directory
	if context.Scope == "" && len(context.ChangedFiles) > 0 {
		firstFile := context.ChangedFiles[0]
		if dir := filepath.Dir(firstFile.Path); dir != "." {
			context.Scope = strings.ToLower(filepath.Base(dir))
		}
	}
}

// analyzeImpact determines the impact level of changes
func (ca *ContextAnalyzer) analyzeImpact(context *CommitContext) {
	totalFiles := len(context.ChangedFiles)
	totalLines := 0

	for _, file := range context.ChangedFiles {
		totalLines += file.LinesAdded + file.LinesDeleted
	}

	// Determine impact based on scope of changes
	if totalFiles >= 10 || totalLines >= 500 {
		context.Impact = "major"
	} else if totalFiles >= 3 || totalLines >= 100 {
		context.Impact = "minor"
	} else {
		context.Impact = "patch"
	}

	// Adjust based on file types
	for _, file := range context.ChangedFiles {
		if file.Type == "config" && file.Status != "docs" {
			// Config changes can have major impact
			if context.Impact == "patch" {
				context.Impact = "minor"
			}
		}
	}
}

// findRelatedFiles finds files that might be related to the changes
func (ca *ContextAnalyzer) findRelatedFiles(context *CommitContext) {
	// TODO: Implement sophisticated relationship detection
	// - Files that import/depend on changed files
	// - Files in the same directory
	// - Files with similar names or purposes
	// - Test files for changed source files

	context.RelatedFiles = []*FileInfo{}
}

// generateSuggestions generates suggestions for the commit message
func (ca *ContextAnalyzer) generateSuggestions(context *CommitContext) {
	var suggestions []string

	// Suggest conventional commit format
	if context.ChangeType != "" && context.Scope != "" {
		suggestions = append(suggestions,
			fmt.Sprintf("Use format: %s(%s): <description>", context.ChangeType, context.Scope))
	} else if context.ChangeType != "" {
		suggestions = append(suggestions,
			fmt.Sprintf("Use format: %s: <description>", context.ChangeType))
	}

	// Suggest specific improvements based on changes
	if len(context.ChangedFiles) == 1 {
		file := context.ChangedFiles[0]
		suggestions = append(suggestions,
			fmt.Sprintf("Focus on changes to %s", filepath.Base(file.Path)))
	}

	// Suggest breaking change notation if major impact
	if context.Impact == "major" {
		suggestions = append(suggestions, "Consider adding BREAKING CHANGE: if this breaks compatibility")
	}

	context.Suggestions = suggestions
}

// generateContextDescription generates a human-readable context description
func (ca *ContextAnalyzer) generateContextDescription(context *CommitContext) {
	var parts []string

	// Repository info
	if context.Repository != nil {
		parts = append(parts, fmt.Sprintf("Repository: %s (%s project)",
			context.Repository.Metadata.Name, context.Repository.Metadata.MainLanguage))
	}

	// Change summary
	parts = append(parts, fmt.Sprintf("Changes: %d files modified", len(context.ChangedFiles)))

	// File types
	typeCount := make(map[string]int)
	for _, file := range context.ChangedFiles {
		typeCount[file.Type]++
	}

	var types []string
	for fileType, count := range typeCount {
		if count == 1 {
			types = append(types, fmt.Sprintf("1 %s file", fileType))
		} else {
			types = append(types, fmt.Sprintf("%d %s files", count, fileType))
		}
	}

	if len(types) > 0 {
		parts = append(parts, fmt.Sprintf("Types: %s", strings.Join(types, ", ")))
	}

	// Scope and impact
	if context.Scope != "" {
		parts = append(parts, fmt.Sprintf("Scope: %s", context.Scope))
	}

	parts = append(parts, fmt.Sprintf("Impact: %s", context.Impact))

	context.Context = strings.Join(parts, " | ")
}
