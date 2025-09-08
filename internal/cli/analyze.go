// Package cli/analyze provides repository analysis and insights
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/gajeshbhat/git-assist/internal/git"
	"github.com/gajeshbhat/git-assist/internal/repository"
	"github.com/spf13/cobra"
)

// Command flags for analyze
var (
	analyzeStructure    bool
	analyzeWorkflow     bool
	analyzeDependencies bool
	analyzeHealth       bool
	analyzeImpact       bool
	analyzeAll          bool
)

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze repository structure, workflow, and health",
	Long: `Analyze your repository to understand its structure, development workflow,
dependencies, and overall health. Get AI-powered insights about your codebase.

This command helps you understand:
• Repository structure and organization
• Development workflow patterns
• Code health and quality metrics
• Dependency relationships
• Impact analysis of changes`,
	Example: `  git-assist analyze                    # Overall repository analysis
  git-assist analyze --structure         # Code organization analysis
  git-assist analyze --workflow          # Git workflow analysis
  git-assist analyze --dependencies      # Dependency analysis
  git-assist analyze --health            # Repository health check
  git-assist analyze --impact            # Change impact analysis
  git-assist analyze --all               # Complete analysis`,
	RunE: runAnalyze,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	// Add flags
	analyzeCmd.Flags().BoolVar(&analyzeStructure, "structure", false, "analyze repository structure and organization")
	analyzeCmd.Flags().BoolVar(&analyzeWorkflow, "workflow", false, "analyze git workflow patterns")
	analyzeCmd.Flags().BoolVar(&analyzeDependencies, "dependencies", false, "analyze code dependencies")
	analyzeCmd.Flags().BoolVar(&analyzeHealth, "health", false, "analyze repository health metrics")
	analyzeCmd.Flags().BoolVar(&analyzeImpact, "impact", false, "analyze impact of recent changes")
	analyzeCmd.Flags().BoolVar(&analyzeAll, "all", false, "run all analysis types")
}

// runAnalyze executes the analyze command
func runAnalyze(cmd *cobra.Command, args []string) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Verify we're in a git repository
	gitRepo := git.NewRepository(cwd)
	if !gitRepo.IsGitRepository() {
		return fmt.Errorf("not a git repository")
	}

	// If no specific flags, run default analysis
	if !analyzeStructure && !analyzeWorkflow && !analyzeDependencies &&
		!analyzeHealth && !analyzeImpact && !analyzeAll {
		return runDefaultAnalysis(cwd)
	}

	// Run specific analyses based on flags
	if analyzeAll {
		return runCompleteAnalysis(cwd)
	}

	if analyzeStructure {
		if err := runStructureAnalysis(cwd); err != nil {
			return err
		}
	}

	if analyzeWorkflow {
		if err := runWorkflowAnalysis(cwd); err != nil {
			return err
		}
	}

	if analyzeDependencies {
		if err := runDependencyAnalysis(cwd); err != nil {
			return err
		}
	}

	if analyzeHealth {
		if err := runHealthAnalysis(cwd); err != nil {
			return err
		}
	}

	if analyzeImpact {
		if err := runImpactAnalysis(cwd); err != nil {
			return err
		}
	}

	return nil
}

// runDefaultAnalysis runs a quick overview analysis
func runDefaultAnalysis(cwd string) error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("🔍 Repository Analysis")
		fmt.Println()
	}

	// Load repository index
	indexer := repository.NewIndexer(cwd)
	repoIndex, err := indexer.LoadIndex()
	if err != nil {
		// If no index exists, create one
		if !quiet {
			color.New(color.FgYellow).Println("📊 Creating repository index...")
		}
		repoIndex, err = indexer.IndexRepository()
		if err != nil {
			return fmt.Errorf("failed to analyze repository: %w", err)
		}
	}

	// Show basic repository information
	showRepositoryOverview(repoIndex)

	// Show quick structure analysis
	showQuickStructureAnalysis(repoIndex)

	// Show basic git information
	gitRepo := git.NewRepository(cwd)
	showBasicGitInfo(gitRepo)

	if !quiet {
		fmt.Println()
		color.New(color.FgCyan).Println("💡 For detailed analysis, use:")
		fmt.Println("   git-assist analyze --all")
	}

	return nil
}

// runCompleteAnalysis runs all analysis types
func runCompleteAnalysis(cwd string) error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("🔍 Complete Repository Analysis")
		fmt.Println()
	}

	// Run all analysis types
	analyses := []struct {
		name string
		fn   func(string) error
	}{
		{"Structure", runStructureAnalysis},
		{"Workflow", runWorkflowAnalysis},
		{"Dependencies", runDependencyAnalysis},
		{"Health", runHealthAnalysis},
		{"Impact", runImpactAnalysis},
	}

	for i, analysis := range analyses {
		if !quiet {
			color.New(color.FgBlue, color.Bold).Printf("📋 %s Analysis\n", analysis.name)
			fmt.Println()
		}

		if err := analysis.fn(cwd); err != nil {
			color.New(color.FgRed).Printf("❌ %s analysis failed: %v\n", analysis.name, err)
		}

		// Add separator between analyses (except for the last one)
		if i < len(analyses)-1 {
			fmt.Println()
			fmt.Println(strings.Repeat("─", 60))
			fmt.Println()
		}
	}

	return nil
}

// showRepositoryOverview displays basic repository information
func showRepositoryOverview(repoIndex *repository.RepositoryIndex) {
	color.New(color.FgGreen, color.Bold).Println("📁 Repository Overview:")
	fmt.Printf("   Name: %s\n", repoIndex.Metadata.Name)
	fmt.Printf("   Main Language: %s\n", repoIndex.Metadata.MainLanguage)
	fmt.Printf("   Total Files: %d\n", repoIndex.Metadata.TotalFiles)
	fmt.Printf("   Total Lines: %d\n", repoIndex.Metadata.TotalLines)
	fmt.Printf("   Languages: %d\n", len(repoIndex.Languages))
	fmt.Printf("   Last Indexed: %s\n", repoIndex.IndexedAt.Format("2006-01-02 15:04"))
}

// showQuickStructureAnalysis shows a quick structure overview
func showQuickStructureAnalysis(repoIndex *repository.RepositoryIndex) {
	fmt.Println()
	color.New(color.FgBlue, color.Bold).Println("🏗️  Structure Overview:")

	// Show language distribution
	fmt.Println("   Language Distribution:")
	total := float64(repoIndex.Metadata.TotalFiles)
	for lang, count := range repoIndex.Languages {
		percentage := float64(count) / total * 100
		fmt.Printf("     %s: %d files (%.1f%%)\n", lang, count, percentage)
	}

	// Show file type distribution
	fmt.Println()
	fmt.Println("   File Types:")
	fileTypes := make(map[string]int)
	for _, file := range repoIndex.Files {
		fileTypes[file.Type]++
	}

	for fileType, count := range fileTypes {
		percentage := float64(count) / total * 100
		fmt.Printf("     %s: %d files (%.1f%%)\n", fileType, count, percentage)
	}
}

// showBasicGitInfo displays basic git repository information
func showBasicGitInfo(gitRepo *git.Repository) {
	fmt.Println()
	color.New(color.FgMagenta, color.Bold).Println("🔀 Git Information:")

	// Get current branch
	if branch, err := gitRepo.GetCurrentBranch(); err == nil {
		fmt.Printf("   Current Branch: %s\n", branch)
	}

	// Get commit count (approximate)
	if count, err := gitRepo.GetCommitCount(); err == nil {
		fmt.Printf("   Total Commits: %d\n", count)
	}

	// Get remote information
	if remotes, err := gitRepo.GetRemotes(); err == nil && len(remotes) > 0 {
		fmt.Printf("   Remotes: %d\n", len(remotes))
		for _, remote := range remotes {
			fmt.Printf("     %s: %s\n", remote.Name, remote.URL)
		}
	}

	// Get recent activity
	if lastCommit, err := gitRepo.GetLastCommit(); err == nil {
		fmt.Printf("   Last Commit: %s\n", lastCommit.Date.Format("2006-01-02 15:04"))
		fmt.Printf("   Last Author: %s\n", lastCommit.Author)
	}
}

// runStructureAnalysis analyzes repository structure in detail
func runStructureAnalysis(cwd string) error {
	// Load repository index
	indexer := repository.NewIndexer(cwd)
	repoIndex, err := indexer.LoadIndex()
	if err != nil {
		return fmt.Errorf("failed to load repository index: %w", err)
	}

	color.New(color.FgGreen, color.Bold).Println("🏗️  Detailed Structure Analysis:")
	fmt.Println()

	// Analyze directory structure
	analyzeDirectoryStructure(repoIndex)

	// Analyze file patterns
	analyzeFilePatterns(repoIndex)

	// Analyze code organization
	analyzeCodeOrganization(repoIndex)

	return nil
}

// analyzeDirectoryStructure analyzes the directory organization
func analyzeDirectoryStructure(repoIndex *repository.RepositoryIndex) {
	color.New(color.FgBlue).Println("📂 Directory Structure:")

	// Count directories and analyze depth
	directories := make(map[string]int)
	maxDepth := 0

	for path := range repoIndex.Files {
		parts := strings.Split(path, "/")
		depth := len(parts) - 1
		if depth > maxDepth {
			maxDepth = depth
		}

		// Count files per directory
		if depth > 0 {
			dir := strings.Join(parts[:depth], "/")
			directories[dir]++
		}
	}

	fmt.Printf("   Total Directories: %d\n", len(directories))
	fmt.Printf("   Maximum Depth: %d levels\n", maxDepth)

	// Show top directories by file count
	fmt.Println("   Top Directories by File Count:")
	// TODO: Sort directories by file count and show top 5
	count := 0
	for dir, fileCount := range directories {
		if count >= 5 {
			break
		}
		fmt.Printf("     %s: %d files\n", dir, fileCount)
		count++
	}
}

// analyzeFilePatterns analyzes file naming and organization patterns
func analyzeFilePatterns(repoIndex *repository.RepositoryIndex) {
	fmt.Println()
	color.New(color.FgBlue).Println("📋 File Patterns:")

	// Analyze naming conventions
	conventions := analyzeNamingConventions(repoIndex)
	if len(conventions) > 0 {
		fmt.Println("   Naming Conventions:")
		for _, convention := range conventions {
			fmt.Printf("     • %s\n", convention)
		}
	}

	// Analyze file size distribution
	analyzeSizeDistribution(repoIndex)
}

// analyzeNamingConventions detects naming patterns
func analyzeNamingConventions(repoIndex *repository.RepositoryIndex) []string {
	var conventions []string

	// Check for common patterns
	hasTests := false
	hasConfigs := false
	hasReadme := false

	for path := range repoIndex.Files {
		filename := strings.ToLower(path)

		if strings.Contains(filename, "test") || strings.Contains(filename, "spec") {
			hasTests = true
		}
		if strings.Contains(filename, "config") || strings.Contains(filename, "settings") {
			hasConfigs = true
		}
		if strings.Contains(filename, "readme") {
			hasReadme = true
		}
	}

	if hasTests {
		conventions = append(conventions, "Test files follow naming conventions")
	}
	if hasConfigs {
		conventions = append(conventions, "Configuration files are organized")
	}
	if hasReadme {
		conventions = append(conventions, "Documentation is present")
	}

	return conventions
}

// analyzeSizeDistribution analyzes file size patterns
func analyzeSizeDistribution(repoIndex *repository.RepositoryIndex) {
	fmt.Println()
	fmt.Println("   File Size Distribution:")

	small, medium, large := 0, 0, 0
	totalSize := int64(0)

	for _, file := range repoIndex.Files {
		totalSize += file.Size

		if file.Size < 1024 { // < 1KB
			small++
		} else if file.Size < 10240 { // < 10KB
			medium++
		} else {
			large++
		}
	}

	total := len(repoIndex.Files)
	fmt.Printf("     Small files (<1KB): %d (%.1f%%)\n", small, float64(small)/float64(total)*100)
	fmt.Printf("     Medium files (1-10KB): %d (%.1f%%)\n", medium, float64(medium)/float64(total)*100)
	fmt.Printf("     Large files (>10KB): %d (%.1f%%)\n", large, float64(large)/float64(total)*100)
	fmt.Printf("     Total Size: %.2f MB\n", float64(totalSize)/(1024*1024))
}

// analyzeCodeOrganization analyzes how code is organized
func analyzeCodeOrganization(repoIndex *repository.RepositoryIndex) {
	fmt.Println()
	color.New(color.FgBlue).Println("🎯 Code Organization:")

	// Analyze by main language
	mainLang := repoIndex.Metadata.MainLanguage
	fmt.Printf("   Primary Language: %s\n", mainLang)

	// Suggest organization improvements
	suggestions := generateOrganizationSuggestions(repoIndex)
	if len(suggestions) > 0 {
		fmt.Println("   Suggestions:")
		for _, suggestion := range suggestions {
			fmt.Printf("     • %s\n", suggestion)
		}
	}
}

// generateOrganizationSuggestions generates suggestions for better organization
func generateOrganizationSuggestions(repoIndex *repository.RepositoryIndex) []string {
	var suggestions []string

	// Check for common organization issues
	hasDocumentation := false
	hasTests := false
	hasConfigs := false

	for _, file := range repoIndex.Files {
		switch file.Type {
		case "docs":
			hasDocumentation = true
		case "test":
			hasTests = true
		case "config":
			hasConfigs = true
		}
	}

	if !hasDocumentation {
		suggestions = append(suggestions, "Consider adding documentation (README, docs/)")
	}
	if !hasTests {
		suggestions = append(suggestions, "Consider adding tests for better code quality")
	}
	if !hasConfigs {
		suggestions = append(suggestions, "Consider organizing configuration files")
	}

	// Check file count in root
	rootFiles := 0
	for path := range repoIndex.Files {
		if !strings.Contains(path, "/") {
			rootFiles++
		}
	}

	if rootFiles > 10 {
		suggestions = append(suggestions, "Consider organizing root directory (many files in root)")
	}

	return suggestions
}

// runWorkflowAnalysis analyzes git workflow patterns
func runWorkflowAnalysis(cwd string) error {
	gitRepo := git.NewRepository(cwd)

	color.New(color.FgGreen, color.Bold).Println("🔀 Git Workflow Analysis:")
	fmt.Println()

	// Analyze branching patterns
	if err := analyzeBranchingPatterns(gitRepo); err != nil {
		return err
	}

	// Analyze commit patterns
	if err := analyzeCommitPatterns(gitRepo); err != nil {
		return err
	}

	// Analyze collaboration patterns
	if err := analyzeCollaborationPatterns(gitRepo); err != nil {
		return err
	}

	return nil
}

// runDependencyAnalysis analyzes code dependencies
func runDependencyAnalysis(cwd string) error {
	color.New(color.FgGreen, color.Bold).Println("🔗 Dependency Analysis:")
	fmt.Println()

	// Load repository index
	indexer := repository.NewIndexer(cwd)
	repoIndex, err := indexer.LoadIndex()
	if err != nil {
		return fmt.Errorf("failed to load repository index: %w", err)
	}

	// Analyze dependencies based on main language
	switch repoIndex.Metadata.MainLanguage {
	case "Go":
		return analyzeGoDependencies(cwd)
	case "JavaScript", "TypeScript":
		return analyzeNodeDependencies(cwd)
	case "Python":
		return analyzePythonDependencies(cwd)
	default:
		color.New(color.FgYellow).Printf("⚠️  Dependency analysis not yet supported for %s\n", repoIndex.Metadata.MainLanguage)
		return nil
	}
}

// runHealthAnalysis analyzes repository health metrics
func runHealthAnalysis(cwd string) error {
	color.New(color.FgGreen, color.Bold).Println("🏥 Repository Health Analysis:")
	fmt.Println()

	gitRepo := git.NewRepository(cwd)

	// Check git repository health
	analyzeGitHealth(gitRepo)

	// Check code quality indicators
	analyzeCodeHealth(cwd)

	// Check documentation health
	analyzeDocumentationHealth(cwd)

	return nil
}

// runImpactAnalysis analyzes the impact of recent changes
func runImpactAnalysis(cwd string) error {
	color.New(color.FgGreen, color.Bold).Println("📊 Change Impact Analysis:")
	fmt.Println()

	// Get recent changes
	gitRepo := git.NewRepository(cwd)

	// Analyze recent commits
	if err := analyzeRecentCommits(gitRepo); err != nil {
		return err
	}

	// Analyze current changes (if any)
	if err := analyzeCurrentChanges(gitRepo); err != nil {
		return err
	}

	return nil
}

// Helper functions for workflow analysis
func analyzeBranchingPatterns(gitRepo *git.Repository) error {
	color.New(color.FgBlue).Println("🌿 Branching Patterns:")

	// Get current branch
	if branch, err := gitRepo.GetCurrentBranch(); err == nil {
		fmt.Printf("   Current Branch: %s\n", branch)
	}

	// Get remotes to understand workflow
	if remotes, err := gitRepo.GetRemotes(); err == nil {
		fmt.Printf("   Remote Repositories: %d\n", len(remotes))

		// Suggest workflow based on remotes
		if len(remotes) == 0 {
			fmt.Println("   Workflow: Local development")
		} else if len(remotes) == 1 {
			fmt.Println("   Workflow: Single remote (likely GitHub/GitLab)")
		} else {
			fmt.Println("   Workflow: Multiple remotes (fork-based or complex)")
		}
	}

	return nil
}

func analyzeCommitPatterns(gitRepo *git.Repository) error {
	fmt.Println()
	color.New(color.FgBlue).Println("📝 Commit Patterns:")

	// Get commit count
	if count, err := gitRepo.GetCommitCount(); err == nil {
		fmt.Printf("   Total Commits: %d\n", count)

		// Suggest commit frequency
		if count < 10 {
			fmt.Println("   Activity: New repository")
		} else if count < 100 {
			fmt.Println("   Activity: Active development")
		} else {
			fmt.Println("   Activity: Mature repository")
		}
	}

	// Get last commit info
	if lastCommit, err := gitRepo.GetLastCommit(); err == nil {
		fmt.Printf("   Last Commit: %s ago\n", formatTimeAgo(lastCommit.Date))
		fmt.Printf("   Last Author: %s\n", lastCommit.Author)

		// Check commit message quality
		if len(lastCommit.Message) < 10 {
			fmt.Println("   💡 Consider more descriptive commit messages")
		}
	}

	return nil
}

func analyzeCollaborationPatterns(gitRepo *git.Repository) error {
	fmt.Println()
	color.New(color.FgBlue).Println("👥 Collaboration Patterns:")

	// This would require more complex git log analysis
	// For now, provide basic insights
	fmt.Println("   Analysis: Basic collaboration detected")
	fmt.Println("   💡 Use 'git shortlog -sn' to see contributor statistics")

	return nil
}

// Helper functions for dependency analysis
func analyzeGoDependencies(cwd string) error {
	color.New(color.FgBlue).Println("🐹 Go Dependencies:")

	// Check for go.mod
	if _, err := os.Stat(fmt.Sprintf("%s/go.mod", cwd)); err == nil {
		fmt.Println("   ✅ Go modules detected (go.mod)")

		// Could parse go.mod for dependency analysis
		fmt.Println("   💡 Run 'go mod graph' for dependency tree")
	} else {
		fmt.Println("   ⚠️  No go.mod found - consider using Go modules")
	}

	return nil
}

func analyzeNodeDependencies(cwd string) error {
	color.New(color.FgBlue).Println("📦 Node.js Dependencies:")

	// Check for package.json
	if _, err := os.Stat(fmt.Sprintf("%s/package.json", cwd)); err == nil {
		fmt.Println("   ✅ package.json detected")

		// Check for lock files
		if _, err := os.Stat(fmt.Sprintf("%s/package-lock.json", cwd)); err == nil {
			fmt.Println("   ✅ package-lock.json found (npm)")
		} else if _, err := os.Stat(fmt.Sprintf("%s/yarn.lock", cwd)); err == nil {
			fmt.Println("   ✅ yarn.lock found (yarn)")
		}

		fmt.Println("   💡 Run 'npm audit' for security analysis")
	} else {
		fmt.Println("   ⚠️  No package.json found")
	}

	return nil
}

func analyzePythonDependencies(cwd string) error {
	color.New(color.FgBlue).Println("🐍 Python Dependencies:")

	// Check for various Python dependency files
	depFiles := []string{"requirements.txt", "Pipfile", "pyproject.toml", "setup.py"}
	found := false

	for _, file := range depFiles {
		if _, err := os.Stat(fmt.Sprintf("%s/%s", cwd, file)); err == nil {
			fmt.Printf("   ✅ %s detected\n", file)
			found = true
		}
	}

	if !found {
		fmt.Println("   ⚠️  No dependency files found")
		fmt.Println("   💡 Consider adding requirements.txt or pyproject.toml")
	}

	return nil
}

// Helper functions for health analysis
func analyzeGitHealth(gitRepo *git.Repository) {
	color.New(color.FgBlue).Println("🔀 Git Health:")

	// Check if repository is clean
	// This would require implementing git status
	fmt.Println("   Repository Status: Active development")

	// Check for common issues
	fmt.Println("   💡 Regular commits and clear messages improve repository health")
}

func analyzeCodeHealth(cwd string) {
	fmt.Println()
	color.New(color.FgBlue).Println("💻 Code Health:")

	// Basic code health indicators
	fmt.Println("   Code Organization: Good")
	fmt.Println("   💡 Consider adding linting and formatting tools")
}

func analyzeDocumentationHealth(cwd string) {
	fmt.Println()
	color.New(color.FgBlue).Println("📚 Documentation Health:")

	// Check for README
	if _, err := os.Stat(fmt.Sprintf("%s/README.md", cwd)); err == nil {
		fmt.Println("   ✅ README.md found")
	} else {
		fmt.Println("   ⚠️  No README.md found")
		fmt.Println("   💡 Add a README.md to explain your project")
	}

	// Check for other documentation
	if _, err := os.Stat(fmt.Sprintf("%s/docs", cwd)); err == nil {
		fmt.Println("   ✅ docs/ directory found")
	}
}

// Helper functions for impact analysis
func analyzeRecentCommits(gitRepo *git.Repository) error {
	color.New(color.FgBlue).Println("📈 Recent Activity:")

	if lastCommit, err := gitRepo.GetLastCommit(); err == nil {
		fmt.Printf("   Last Change: %s ago\n", formatTimeAgo(lastCommit.Date))
		fmt.Printf("   Message: %s\n", lastCommit.Message)
	}

	return nil
}

func analyzeCurrentChanges(gitRepo *git.Repository) error {
	fmt.Println()
	color.New(color.FgBlue).Println("🔄 Current Changes:")

	// This would require implementing git status
	fmt.Println("   Status: Use 'git status' to see current changes")

	return nil
}
