// Package repository provides repository analysis and indexing functionality
//
// This package handles repository structure analysis, file indexing, and
// context extraction for AI-powered Git operations.
package repository

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RepositoryIndex represents the indexed state of a repository
type RepositoryIndex struct {
	Version   string               `json:"version"`
	IndexedAt time.Time            `json:"indexed_at"`
	RootPath  string               `json:"root_path"`
	Files     map[string]*FileInfo `json:"files"`
	Structure *DirectoryStructure  `json:"structure"`
	Languages map[string]int       `json:"languages"`
	Patterns  *RepositoryPatterns  `json:"patterns"`
	Metadata  *RepositoryMetadata  `json:"metadata"`
}

// FileInfo contains information about a single file
type FileInfo struct {
	Path         string    `json:"path"`
	RelativePath string    `json:"relative_path"`
	Size         int64     `json:"size"`
	ModTime      time.Time `json:"mod_time"`
	Language     string    `json:"language"`
	Type         string    `json:"type"` // "source", "config", "docs", "test", etc.
	Hash         string    `json:"hash"`
	LineCount    int       `json:"line_count"`
	Functions    []string  `json:"functions,omitempty"`
	Imports      []string  `json:"imports,omitempty"`
	Exports      []string  `json:"exports,omitempty"`
	Dependencies []string  `json:"dependencies,omitempty"`
	Tags         []string  `json:"tags,omitempty"`
}

// DirectoryStructure represents the repository's directory structure
type DirectoryStructure struct {
	Name      string                         `json:"name"`
	Path      string                         `json:"path"`
	Type      string                         `json:"type"` // "root", "src", "test", "docs", etc.
	Children  map[string]*DirectoryStructure `json:"children,omitempty"`
	FileCount int                            `json:"file_count"`
	Languages map[string]int                 `json:"languages"`
}

// RepositoryPatterns contains detected patterns in the repository
type RepositoryPatterns struct {
	Framework     string   `json:"framework,omitempty"`      // "react", "vue", "django", etc.
	BuildSystem   string   `json:"build_system,omitempty"`   // "webpack", "vite", "maven", etc.
	TestFramework string   `json:"test_framework,omitempty"` // "jest", "pytest", "go test", etc.
	Architecture  string   `json:"architecture,omitempty"`   // "mvc", "microservices", etc.
	Conventions   []string `json:"conventions,omitempty"`    // Detected naming conventions
}

// RepositoryMetadata contains high-level repository information
type RepositoryMetadata struct {
	Name         string   `json:"name"`
	Description  string   `json:"description,omitempty"`
	MainLanguage string   `json:"main_language"`
	TotalFiles   int      `json:"total_files"`
	TotalLines   int      `json:"total_lines"`
	LastCommit   string   `json:"last_commit,omitempty"`
	Branch       string   `json:"branch,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Contributors []string `json:"contributors,omitempty"`
	License      string   `json:"license,omitempty"`
}

// Indexer handles repository indexing operations
type Indexer struct {
	rootPath    string
	indexPath   string
	gitIgnore   []string
	maxFileSize int64
}

// NewIndexer creates a new repository indexer
func NewIndexer(rootPath string) *Indexer {
	indexPath := filepath.Join(rootPath, ".git", "git-assist", "index.json")

	return &Indexer{
		rootPath:    rootPath,
		indexPath:   indexPath,
		gitIgnore:   []string{},
		maxFileSize: 1024 * 1024, // 1MB max file size for indexing
	}
}

// IndexRepository performs a full repository index
func (idx *Indexer) IndexRepository() (*RepositoryIndex, error) {
	fmt.Println("🔍 Indexing repository...")

	// Load .gitignore patterns
	if err := idx.loadGitIgnore(); err != nil {
		// Non-fatal error, continue without gitignore
		fmt.Printf("Warning: Could not load .gitignore: %v\n", err)
	}

	// Create repository index
	repoIndex := &RepositoryIndex{
		Version:   "1.0",
		IndexedAt: time.Now(),
		RootPath:  idx.rootPath,
		Files:     make(map[string]*FileInfo),
		Languages: make(map[string]int),
		Structure: &DirectoryStructure{
			Name:      filepath.Base(idx.rootPath),
			Path:      idx.rootPath,
			Type:      "root",
			Children:  make(map[string]*DirectoryStructure),
			Languages: make(map[string]int),
		},
		Patterns: &RepositoryPatterns{},
		Metadata: &RepositoryMetadata{
			Name: filepath.Base(idx.rootPath),
		},
	}

	// Walk the repository
	err := filepath.WalkDir(idx.rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory and other ignored paths
		if idx.shouldIgnore(path) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Process files
		if !d.IsDir() {
			fileInfo, err := idx.analyzeFile(path)
			if err != nil {
				// Non-fatal error, skip this file
				return nil
			}

			if fileInfo != nil {
				repoIndex.Files[fileInfo.RelativePath] = fileInfo
				repoIndex.Languages[fileInfo.Language]++
				repoIndex.Metadata.TotalFiles++
				repoIndex.Metadata.TotalLines += fileInfo.LineCount
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk repository: %w", err)
	}

	// Analyze patterns and metadata
	idx.analyzePatterns(repoIndex)
	idx.analyzeMetadata(repoIndex)

	// Save index to disk
	if err := idx.saveIndex(repoIndex); err != nil {
		return nil, fmt.Errorf("failed to save index: %w", err)
	}

	fmt.Printf("✅ Indexed %d files in %d languages\n",
		repoIndex.Metadata.TotalFiles, len(repoIndex.Languages))

	return repoIndex, nil
}

// LoadIndex loads an existing repository index
func (idx *Indexer) LoadIndex() (*RepositoryIndex, error) {
	data, err := os.ReadFile(idx.indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read index file: %w", err)
	}

	var repoIndex RepositoryIndex
	if err := json.Unmarshal(data, &repoIndex); err != nil {
		return nil, fmt.Errorf("failed to parse index file: %w", err)
	}

	return &repoIndex, nil
}

// IsIndexStale checks if the index needs to be refreshed
func (idx *Indexer) IsIndexStale() (bool, error) {
	repoIndex, err := idx.LoadIndex()
	if err != nil {
		return true, nil // If we can't load, assume stale
	}

	// Check if index is older than 1 hour
	if time.Since(repoIndex.IndexedAt) > time.Hour {
		return true, nil
	}

	// TODO: Check if any tracked files have been modified
	// This would require comparing file modification times

	return false, nil
}

// loadGitIgnore loads .gitignore patterns
func (idx *Indexer) loadGitIgnore() error {
	gitIgnorePath := filepath.Join(idx.rootPath, ".gitignore")
	file, err := os.Open(gitIgnorePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			idx.gitIgnore = append(idx.gitIgnore, line)
		}
	}

	return scanner.Err()
}

// shouldIgnore checks if a path should be ignored
func (idx *Indexer) shouldIgnore(path string) bool {
	relPath, err := filepath.Rel(idx.rootPath, path)
	if err != nil {
		return true
	}

	// Always ignore .git directory
	if strings.Contains(relPath, ".git") {
		return true
	}

	// Common ignore patterns
	ignorePaths := []string{
		"node_modules", "vendor", ".venv", "venv", "__pycache__",
		".pytest_cache", ".coverage", "dist", "build", "target",
		".DS_Store", "Thumbs.db", "*.log", "*.tmp",
	}

	for _, pattern := range ignorePaths {
		if strings.Contains(relPath, pattern) {
			return true
		}
	}

	// TODO: Implement proper gitignore pattern matching

	return false
}

// analyzeFile analyzes a single file and extracts information
func (idx *Indexer) analyzeFile(path string) (*FileInfo, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// Skip large files
	if stat.Size() > idx.maxFileSize {
		return nil, nil
	}

	relPath, err := filepath.Rel(idx.rootPath, path)
	if err != nil {
		return nil, err
	}

	// Calculate file hash
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(content))

	// Count lines
	lineCount := strings.Count(string(content), "\n") + 1

	fileInfo := &FileInfo{
		Path:         path,
		RelativePath: relPath,
		Size:         stat.Size(),
		ModTime:      stat.ModTime(),
		Language:     detectLanguage(path),
		Type:         detectFileType(path),
		Hash:         hash,
		LineCount:    lineCount,
	}

	// TODO: Extract functions, imports, exports for supported languages

	return fileInfo, nil
}

// detectLanguage detects the programming language of a file
func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	languageMap := map[string]string{
		".go":    "Go",
		".js":    "JavaScript",
		".ts":    "TypeScript",
		".py":    "Python",
		".java":  "Java",
		".cpp":   "C++",
		".c":     "C",
		".rs":    "Rust",
		".rb":    "Ruby",
		".php":   "PHP",
		".cs":    "C#",
		".swift": "Swift",
		".kt":    "Kotlin",
		".scala": "Scala",
		".sh":    "Shell",
		".bash":  "Shell",
		".zsh":   "Shell",
		".fish":  "Shell",
		".ps1":   "PowerShell",
		".sql":   "SQL",
		".html":  "HTML",
		".css":   "CSS",
		".scss":  "SCSS",
		".sass":  "Sass",
		".less":  "Less",
		".vue":   "Vue",
		".jsx":   "JSX",
		".tsx":   "TSX",
		".json":  "JSON",
		".yaml":  "YAML",
		".yml":   "YAML",
		".xml":   "XML",
		".toml":  "TOML",
		".ini":   "INI",
		".cfg":   "Config",
		".conf":  "Config",
		".md":    "Markdown",
		".txt":   "Text",
		".rst":   "reStructuredText",
		".tex":   "LaTeX",
	}

	if lang, exists := languageMap[ext]; exists {
		return lang
	}

	return "Unknown"
}

// detectFileType detects the type/purpose of a file
func detectFileType(path string) string {
	filename := strings.ToLower(filepath.Base(path))
	dir := strings.ToLower(filepath.Dir(path))

	// Test files
	if strings.Contains(filename, "test") || strings.Contains(dir, "test") ||
		strings.Contains(filename, "spec") || strings.Contains(dir, "spec") {
		return "test"
	}

	// Documentation
	if strings.Contains(dir, "doc") || strings.Contains(filename, "readme") ||
		strings.HasSuffix(filename, ".md") {
		return "docs"
	}

	// Configuration
	configFiles := []string{"config", "settings", "env", "dockerfile", "makefile",
		"package.json", "go.mod", "requirements.txt", "pom.xml", "build.gradle"}
	for _, configFile := range configFiles {
		if strings.Contains(filename, configFile) {
			return "config"
		}
	}

	// Source code (default)
	return "source"
}

// analyzePatterns detects patterns in the repository
func (idx *Indexer) analyzePatterns(repoIndex *RepositoryIndex) {
	// TODO: Implement pattern detection
	// - Framework detection (React, Vue, Django, etc.)
	// - Build system detection (webpack, maven, etc.)
	// - Architecture patterns (MVC, microservices, etc.)
}

// analyzeMetadata extracts repository metadata
func (idx *Indexer) analyzeMetadata(repoIndex *RepositoryIndex) {
	// Find main language
	maxCount := 0
	for lang, count := range repoIndex.Languages {
		if count > maxCount {
			maxCount = count
			repoIndex.Metadata.MainLanguage = lang
		}
	}

	// TODO: Extract more metadata
	// - Git information (last commit, branch, contributors)
	// - License detection
	// - Project description from README
}

// saveIndex saves the repository index to disk
func (idx *Indexer) saveIndex(repoIndex *RepositoryIndex) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(idx.indexPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(repoIndex, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(idx.indexPath, data, 0644)
}
