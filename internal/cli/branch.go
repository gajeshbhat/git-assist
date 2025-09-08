// Package cli/branch provides intelligent branch management
package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/gajeshbhat/git-assist/internal/git"
	"github.com/spf13/cobra"
)

// Command flags for branch
var (
	branchAnalyze  bool
	branchCleanup  bool
	branchSuggest  bool
	branchStrategy bool
	branchCreate   string
	branchSafe     bool
)

// branchCmd represents the branch command
var branchCmd = &cobra.Command{
	Use:   "branch",
	Short: "Intelligent branch management and analysis",
	Long: `Manage git branches with AI-powered insights and automation.
Get suggestions for branch names, safely clean up old branches,
and understand branch relationships.

This command helps you:
• Analyze branch relationships and status
• Safely clean up merged branches
• Suggest meaningful branch names
• Choose the right merge strategy
• Create branches with best practices`,
	Example: `  git-assist branch analyze             # Analyze branch relationships
  git-assist branch cleanup             # Safe branch cleanup
  git-assist branch suggest             # Suggest branch names
  git-assist branch create feature-auth # Create branch with validation
  git-assist branch --strategy          # Suggest merge strategy`,
	RunE: runBranch,
}

func init() {
	rootCmd.AddCommand(branchCmd)

	// Add flags
	branchCmd.Flags().BoolVar(&branchAnalyze, "analyze", false, "analyze branch relationships and status")
	branchCmd.Flags().BoolVar(&branchCleanup, "cleanup", false, "safely clean up merged branches")
	branchCmd.Flags().BoolVar(&branchSuggest, "suggest", false, "suggest branch names based on current changes")
	branchCmd.Flags().BoolVar(&branchStrategy, "strategy", false, "suggest merge strategy for current branch")
	branchCmd.Flags().StringVar(&branchCreate, "create", "", "create a new branch with validation")
	branchCmd.Flags().BoolVar(&branchSafe, "safe", false, "use safe mode (ask before destructive operations)")
}

// runBranch executes the branch command
func runBranch(cmd *cobra.Command, args []string) error {
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

	// Handle different branch operations
	if branchAnalyze {
		return analyzeBranches(gitRepo)
	}

	if branchCleanup {
		return cleanupBranches(gitRepo)
	}

	if branchSuggest {
		return suggestBranchNames(gitRepo)
	}

	if branchStrategy {
		return suggestMergeStrategy(gitRepo)
	}

	if branchCreate != "" {
		return createBranchWithValidation(gitRepo, branchCreate)
	}

	// Default: show branch analysis
	return analyzeBranches(gitRepo)
}

// analyzeBranches analyzes branch relationships and status
func analyzeBranches(gitRepo *git.Repository) error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("🌿 Branch Analysis")
		fmt.Println()
	}

	// Get current branch
	currentBranch, err := gitRepo.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	color.New(color.FgGreen, color.Bold).Printf("📍 Current Branch: %s\n", currentBranch)
	fmt.Println()

	// Get all branches
	branches, err := getAllBranches(gitRepo)
	if err != nil {
		return fmt.Errorf("failed to get branches: %w", err)
	}

	// Analyze local branches
	analyzeBranchList(branches.Local, "Local Branches", currentBranch)

	// Analyze remote branches if any
	if len(branches.Remote) > 0 {
		fmt.Println()
		analyzeBranchList(branches.Remote, "Remote Branches", "")
	}

	// Provide insights and suggestions
	fmt.Println()
	provideBranchInsights(branches, currentBranch)

	return nil
}

// cleanupBranches safely removes merged branches
func cleanupBranches(gitRepo *git.Repository) error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("🧹 Branch Cleanup")
		fmt.Println()
	}

	// Get merged branches
	mergedBranches, err := getMergedBranches(gitRepo)
	if err != nil {
		return fmt.Errorf("failed to get merged branches: %w", err)
	}

	if len(mergedBranches) == 0 {
		color.New(color.FgGreen).Println("✅ No branches need cleanup")
		return nil
	}

	color.New(color.FgYellow).Printf("🔍 Found %d merged branches:\n", len(mergedBranches))
	for _, branch := range mergedBranches {
		fmt.Printf("   • %s\n", branch)
	}

	// Ask for confirmation if in safe mode or interactive
	if branchSafe || !quiet {
		fmt.Println()
		color.New(color.FgYellow).Print("⚠️  Delete these branches? [y/N]: ")

		var response string
		fmt.Scanln(&response)

		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			color.New(color.FgBlue).Println("🚫 Cleanup cancelled")
			return nil
		}
	}

	// Delete merged branches
	deleted := 0
	for _, branch := range mergedBranches {
		if err := deleteBranch(gitRepo, branch); err != nil {
			color.New(color.FgRed).Printf("❌ Failed to delete %s: %v\n", branch, err)
		} else {
			color.New(color.FgGreen).Printf("✅ Deleted %s\n", branch)
			deleted++
		}
	}

	fmt.Println()
	color.New(color.FgGreen, color.Bold).Printf("🎉 Cleanup complete! Deleted %d branches\n", deleted)

	return nil
}

// suggestBranchNames suggests branch names based on current changes
func suggestBranchNames(gitRepo *git.Repository) error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("💡 Branch Name Suggestions")
		fmt.Println()
	}

	// Get current changes to suggest names
	diff, err := gitRepo.GetStagedDiff()
	if err != nil {
		diff, _ = gitRepo.GetWorkingDiff() // Try working directory changes
	}

	suggestions := generateBranchNameSuggestions(diff)

	color.New(color.FgGreen).Println("🌿 Suggested branch names:")
	for i, suggestion := range suggestions {
		fmt.Printf("   %d. %s\n", i+1, suggestion)
	}

	fmt.Println()
	color.New(color.FgCyan).Println("💡 Branch naming best practices:")
	fmt.Println("   • Use descriptive names (feature/user-authentication)")
	fmt.Println("   • Include type prefix (feature/, bugfix/, hotfix/)")
	fmt.Println("   • Use kebab-case (feature/user-auth, not feature/userAuth)")
	fmt.Println("   • Keep names concise but clear")

	return nil
}

// suggestMergeStrategy suggests the best merge strategy
func suggestMergeStrategy(gitRepo *git.Repository) error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("🔀 Merge Strategy Suggestion")
		fmt.Println()
	}

	currentBranch, err := gitRepo.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Analyze branch for merge strategy
	strategy := analyzeMergeStrategy(gitRepo, currentBranch)

	color.New(color.FgGreen, color.Bold).Printf("📋 Recommended strategy for '%s':\n", currentBranch)
	fmt.Printf("   Strategy: %s\n", strategy.Name)
	fmt.Printf("   Reason: %s\n", strategy.Reason)
	fmt.Printf("   Command: %s\n", strategy.Command)

	fmt.Println()
	color.New(color.FgBlue).Println("📚 Strategy explanations:")
	fmt.Println("   • Merge: Preserves branch history, creates merge commit")
	fmt.Println("   • Rebase: Linear history, cleaner but rewrites commits")
	fmt.Println("   • Squash: Combines all commits into one")

	return nil
}

// createBranchWithValidation creates a branch with name validation
func createBranchWithValidation(gitRepo *git.Repository, branchName string) error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Printf("🌱 Creating Branch: %s\n", branchName)
		fmt.Println()
	}

	// Validate branch name
	if err := validateBranchName(branchName); err != nil {
		return fmt.Errorf("invalid branch name: %w", err)
	}

	// Check if branch already exists
	if branchExists(gitRepo, branchName) {
		return fmt.Errorf("branch '%s' already exists", branchName)
	}

	// Create the branch
	if err := createBranch(gitRepo, branchName); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	color.New(color.FgGreen).Printf("✅ Created branch '%s'\n", branchName)

	// Ask if user wants to switch to the new branch
	if !quiet {
		fmt.Println()
		color.New(color.FgYellow).Print("🔄 Switch to new branch? [Y/n]: ")

		var response string
		fmt.Scanln(&response)

		if strings.ToLower(response) != "n" && strings.ToLower(response) != "no" {
			if err := switchToBranch(gitRepo, branchName); err != nil {
				color.New(color.FgRed).Printf("❌ Failed to switch to branch: %v\n", err)
			} else {
				color.New(color.FgGreen).Printf("🔄 Switched to branch '%s'\n", branchName)
			}
		}
	}

	return nil
}

// BranchList represents local and remote branches
type BranchList struct {
	Local  []string
	Remote []string
}

// getAllBranches gets all local and remote branches
func getAllBranches(gitRepo *git.Repository) (*BranchList, error) {
	// Get local branches
	localCmd := exec.Command("git", "branch")
	localCmd.Dir = gitRepo.Path()
	localOutput, err := localCmd.Output()
	if err != nil {
		return nil, err
	}

	// Get remote branches
	remoteCmd := exec.Command("git", "branch", "-r")
	remoteCmd.Dir = gitRepo.Path()
	remoteOutput, _ := remoteCmd.Output() // Don't fail if no remotes

	branches := &BranchList{
		Local:  parseBranchOutput(string(localOutput), true),
		Remote: parseBranchOutput(string(remoteOutput), false),
	}

	return branches, nil
}

// parseBranchOutput parses git branch command output
func parseBranchOutput(output string, isLocal bool) []string {
	var branches []string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove current branch indicator (*)
		if strings.HasPrefix(line, "*") {
			line = strings.TrimSpace(line[1:])
		}

		// Skip HEAD references for remote branches
		if !isLocal && strings.Contains(line, "HEAD") {
			continue
		}

		branches = append(branches, line)
	}

	return branches
}

// analyzeBranchList analyzes and displays branch information
func analyzeBranchList(branches []string, title, currentBranch string) {
	color.New(color.FgBlue, color.Bold).Printf("🌿 %s (%d):\n", title, len(branches))

	for _, branch := range branches {
		if branch == currentBranch {
			color.New(color.FgGreen).Printf("   • %s (current)\n", branch)
		} else {
			fmt.Printf("   • %s\n", branch)
		}
	}
}

// provideBranchInsights provides insights about branch structure
func provideBranchInsights(branches *BranchList, currentBranch string) {
	color.New(color.FgMagenta, color.Bold).Println("💡 Insights:")

	totalBranches := len(branches.Local) + len(branches.Remote)
	fmt.Printf("   Total branches: %d (%d local, %d remote)\n",
		totalBranches, len(branches.Local), len(branches.Remote))

	// Provide suggestions based on branch count
	if len(branches.Local) > 10 {
		fmt.Println("   💡 Consider cleaning up old branches with: git-assist branch cleanup")
	}

	if len(branches.Remote) == 0 {
		fmt.Println("   💡 No remote branches found - consider setting up a remote repository")
	}

	if currentBranch != "main" && currentBranch != "master" {
		fmt.Printf("   💡 Working on feature branch '%s' - remember to merge when ready\n", currentBranch)
	}
}

// getMergedBranches returns branches that have been merged
func getMergedBranches(gitRepo *git.Repository) ([]string, error) {
	cmd := exec.Command("git", "branch", "--merged")
	cmd.Dir = gitRepo.Path()

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var mergedBranches []string
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove current branch indicator
		if strings.HasPrefix(line, "*") {
			line = strings.TrimSpace(line[1:])
		}

		// Don't include main/master branches
		if line == "main" || line == "master" {
			continue
		}

		mergedBranches = append(mergedBranches, line)
	}

	return mergedBranches, nil
}

// deleteBranch deletes a local branch
func deleteBranch(gitRepo *git.Repository, branchName string) error {
	cmd := exec.Command("git", "branch", "-d", branchName)
	cmd.Dir = gitRepo.Path()
	return cmd.Run()
}

// generateBranchNameSuggestions generates branch name suggestions
func generateBranchNameSuggestions(diff string) []string {
	suggestions := []string{
		"feature/new-feature",
		"bugfix/fix-issue",
		"hotfix/urgent-fix",
		"refactor/improve-code",
		"docs/update-documentation",
	}

	// Analyze diff to provide better suggestions
	if strings.Contains(diff, "test") {
		suggestions = append(suggestions, "test/add-tests")
	}

	if strings.Contains(diff, "README") {
		suggestions = append(suggestions, "docs/update-readme")
	}

	if strings.Contains(diff, "config") {
		suggestions = append(suggestions, "config/update-settings")
	}

	return suggestions
}

// MergeStrategy represents a merge strategy recommendation
type MergeStrategy struct {
	Name    string
	Reason  string
	Command string
}

// analyzeMergeStrategy analyzes the best merge strategy for a branch
func analyzeMergeStrategy(gitRepo *git.Repository, branchName string) MergeStrategy {
	// Simple heuristics for merge strategy
	if branchName == "main" || branchName == "master" {
		return MergeStrategy{
			Name:    "No merge needed",
			Reason:  "Already on main branch",
			Command: "git status",
		}
	}

	// Default to merge for safety
	return MergeStrategy{
		Name:    "Merge",
		Reason:  "Preserves branch history and is safest option",
		Command: "git checkout main && git merge " + branchName,
	}
}

// validateBranchName validates branch name according to git rules
func validateBranchName(name string) error {
	if name == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	// Check for invalid characters
	invalidChars := []string{" ", "~", "^", ":", "?", "*", "[", "\\"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("branch name contains invalid character: %s", char)
		}
	}

	// Check for reserved names
	reserved := []string{"HEAD", "-", "refs/"}
	for _, res := range reserved {
		if strings.Contains(name, res) {
			return fmt.Errorf("branch name contains reserved word: %s", res)
		}
	}

	return nil
}

// branchExists checks if a branch already exists
func branchExists(gitRepo *git.Repository, branchName string) bool {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branchName)
	cmd.Dir = gitRepo.Path()
	return cmd.Run() == nil
}

// createBranch creates a new branch
func createBranch(gitRepo *git.Repository, branchName string) error {
	cmd := exec.Command("git", "branch", branchName)
	cmd.Dir = gitRepo.Path()
	return cmd.Run()
}

// switchToBranch switches to a branch
func switchToBranch(gitRepo *git.Repository, branchName string) error {
	cmd := exec.Command("git", "checkout", branchName)
	cmd.Dir = gitRepo.Path()
	return cmd.Run()
}
