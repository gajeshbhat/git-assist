// Package cli/rebase provides intelligent rebase operations
package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/gajeshbhat/git-assist/internal/ai"
	"github.com/gajeshbhat/git-assist/internal/git"
	"github.com/spf13/cobra"
)

// Command flags for rebase
var (
	rebaseInteractive bool
	rebaseExplain     bool
	rebaseSafe        bool
	rebaseCleanup     bool
	rebaseConflicts   bool
	rebaseTarget      string
	rebaseAuto        bool
)

// rebaseCmd represents the rebase command
var rebaseCmd = &cobra.Command{
	Use:   "rebase [target-branch]",
	Short: "AI-powered intelligent rebasing with safety checks",
	Long: `Perform intelligent git rebase operations with AI guidance and safety checks.
This command helps you rebase safely while keeping your changes and ensuring
you have all the latest updates from the target branch.

Features:
• AI-guided interactive rebase
• Safety checks before rebasing
• Automatic conflict resolution assistance
• Clean up commit history
• Explain rebase process step-by-step`,
	Example: `  git-assist rebase main                # Rebase current branch onto main
  git-assist rebase --interactive        # AI-guided interactive rebase
  git-assist rebase --safe main          # Safe rebase with checks
  git-assist rebase --explain            # Explain rebase process
  git-assist rebase --cleanup            # Clean up commit history
  git-assist rebase --conflicts          # Help resolve rebase conflicts`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRebase,
}

func init() {
	rootCmd.AddCommand(rebaseCmd)

	// Add flags
	rebaseCmd.Flags().BoolVar(&rebaseInteractive, "interactive", false, "AI-guided interactive rebase")
	rebaseCmd.Flags().BoolVar(&rebaseExplain, "explain", false, "explain rebase process step-by-step")
	rebaseCmd.Flags().BoolVar(&rebaseSafe, "safe", false, "perform safety checks before rebasing")
	rebaseCmd.Flags().BoolVar(&rebaseCleanup, "cleanup", false, "clean up commit history")
	rebaseCmd.Flags().BoolVar(&rebaseConflicts, "conflicts", false, "help resolve rebase conflicts")
	rebaseCmd.Flags().StringVar(&rebaseTarget, "target", "", "target branch to rebase onto")
	rebaseCmd.Flags().BoolVar(&rebaseAuto, "auto", false, "automatically handle simple rebases")
}

// runRebase executes the rebase command
func runRebase(cmd *cobra.Command, args []string) error {
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

	// Handle different rebase operations
	if rebaseExplain {
		return explainRebaseProcess()
	}

	if rebaseConflicts {
		return helpResolveConflicts(gitRepo)
	}

	if rebaseCleanup {
		return cleanupCommitHistory(gitRepo)
	}

	// Determine target branch
	targetBranch := rebaseTarget
	if len(args) > 0 {
		targetBranch = args[0]
	}
	if targetBranch == "" {
		targetBranch = "main" // Default to main
	}

	// Perform the rebase operation
	if rebaseInteractive {
		return runInteractiveRebase(gitRepo, targetBranch)
	}

	if rebaseSafe || !rebaseAuto {
		return runSafeRebase(gitRepo, targetBranch)
	}

	return runAutoRebase(gitRepo, targetBranch)
}

// explainRebaseProcess explains how rebase works
func explainRebaseProcess() error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("📚 Understanding Git Rebase")
		fmt.Println()
	}

	explanations := []struct {
		title       string
		description string
	}{
		{
			"What is Rebase?",
			"Rebase moves your commits to a new base, creating a linear history. It 'replays' your commits on top of another branch.",
		},
		{
			"When to Use Rebase",
			"• Update your feature branch with latest main\n   • Clean up commit history before merging\n   • Create a linear project history",
		},
		{
			"Rebase vs Merge",
			"• Rebase: Linear history, rewrites commits (new hashes)\n   • Merge: Preserves history, creates merge commits",
		},
		{
			"Safety Rules",
			"• Never rebase commits that have been pushed and shared\n   • Always commit your changes before rebasing\n   • Use 'git rebase --abort' if things go wrong",
		},
		{
			"Interactive Rebase",
			"Allows you to:\n   • Reorder commits\n   • Squash multiple commits into one\n   • Edit commit messages\n   • Drop unwanted commits",
		},
	}

	for i, explanation := range explanations {
		color.New(color.FgGreen, color.Bold).Printf("%d. %s\n", i+1, explanation.title)
		fmt.Printf("   %s\n", explanation.description)
		fmt.Println()
	}

	color.New(color.FgCyan).Println("💡 Ready to try? Use:")
	fmt.Println("   git-assist rebase --safe main")
	fmt.Println("   git-assist rebase --interactive")

	return nil
}

// runSafeRebase performs a rebase with safety checks
func runSafeRebase(gitRepo *git.Repository, targetBranch string) error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Printf("🛡️  Safe Rebase onto '%s'\n", targetBranch)
		fmt.Println()
	}

	// Perform safety checks
	if err := performSafetyChecks(gitRepo, targetBranch); err != nil {
		return err
	}

	// Get current branch
	currentBranch, err := gitRepo.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Show what will happen
	color.New(color.FgBlue).Printf("📋 Rebase Plan:\n")
	fmt.Printf("   Current branch: %s\n", currentBranch)
	fmt.Printf("   Target branch: %s\n", targetBranch)
	fmt.Printf("   Operation: Move %s commits onto %s\n", currentBranch, targetBranch)

	// Ask for confirmation
	fmt.Println()
	color.New(color.FgYellow).Print("⚠️  Proceed with rebase? [y/N]: ")

	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		color.New(color.FgBlue).Println("🚫 Rebase cancelled")
		return nil
	}

	// Perform the rebase
	return executeRebase(gitRepo, targetBranch, false)
}

// runAutoRebase performs an automatic rebase
func runAutoRebase(gitRepo *git.Repository, targetBranch string) error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Printf("🤖 Auto Rebase onto '%s'\n", targetBranch)
		fmt.Println()
	}

	// Quick safety check
	if err := checkBasicSafety(gitRepo); err != nil {
		return err
	}

	// Perform the rebase
	return executeRebase(gitRepo, targetBranch, false)
}

// runInteractiveRebase performs an AI-guided interactive rebase
func runInteractiveRebase(gitRepo *git.Repository, targetBranch string) error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Printf("🧙 AI-Guided Interactive Rebase\n")
		fmt.Println()
	}

	// Perform safety checks
	if err := performSafetyChecks(gitRepo, targetBranch); err != nil {
		return err
	}

	// Get commit history for analysis
	commits, err := getCommitHistory(gitRepo, targetBranch)
	if err != nil {
		return fmt.Errorf("failed to get commit history: %w", err)
	}

	if len(commits) == 0 {
		color.New(color.FgYellow).Println("⚠️  No commits to rebase")
		return nil
	}

	// Analyze commits and provide suggestions
	suggestions := analyzeCommitsForRebase(commits)

	color.New(color.FgGreen).Println("📊 Commit Analysis:")
	for i, commit := range commits {
		fmt.Printf("   %d. %s - %s\n", i+1, commit.Hash[:8], commit.Message)
	}

	fmt.Println()
	color.New(color.FgMagenta).Println("💡 AI Suggestions:")
	for _, suggestion := range suggestions {
		fmt.Printf("   • %s\n", suggestion)
	}

	// Ask if user wants to proceed with interactive rebase
	fmt.Println()
	color.New(color.FgYellow).Print("🔄 Start interactive rebase? [y/N]: ")

	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		color.New(color.FgBlue).Println("🚫 Interactive rebase cancelled")
		return nil
	}

	// Perform interactive rebase
	return executeRebase(gitRepo, targetBranch, true)
}

// performSafetyChecks performs comprehensive safety checks
func performSafetyChecks(gitRepo *git.Repository, targetBranch string) error {
	color.New(color.FgBlue).Println("🔍 Performing Safety Checks...")

	// Check if working directory is clean
	if !isWorkingDirectoryClean(gitRepo) {
		return fmt.Errorf("working directory is not clean - commit or stash your changes first")
	}
	color.New(color.FgGreen).Println("   ✅ Working directory is clean")

	// Check if target branch exists
	if !branchExists(gitRepo, targetBranch) {
		return fmt.Errorf("target branch '%s' does not exist", targetBranch)
	}
	color.New(color.FgGreen).Printf("   ✅ Target branch '%s' exists\n", targetBranch)

	// Check if we're not already on target branch
	currentBranch, err := gitRepo.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	if currentBranch == targetBranch {
		return fmt.Errorf("already on target branch '%s'", targetBranch)
	}
	color.New(color.FgGreen).Printf("   ✅ Current branch '%s' is different from target\n", currentBranch)

	// Check for unpushed commits (warning, not error)
	if hasUnpushedCommits(gitRepo) {
		color.New(color.FgYellow).Println("   ⚠️  You have unpushed commits - they will be rewritten")
	}

	fmt.Println()
	return nil
}

// checkBasicSafety performs basic safety checks
func checkBasicSafety(gitRepo *git.Repository) error {
	if !isWorkingDirectoryClean(gitRepo) {
		return fmt.Errorf("working directory is not clean - commit or stash your changes first")
	}
	return nil
}

// executeRebase performs the actual rebase operation
func executeRebase(gitRepo *git.Repository, targetBranch string, interactive bool) error {
	color.New(color.FgBlue).Printf("🔄 Rebasing onto '%s'...\n", targetBranch)

	var cmd *exec.Cmd
	if interactive {
		cmd = exec.Command("git", "rebase", "-i", targetBranch)
	} else {
		cmd = exec.Command("git", "rebase", targetBranch)
	}

	cmd.Dir = gitRepo.Path()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		color.New(color.FgRed).Println("❌ Rebase failed")
		fmt.Println()
		color.New(color.FgCyan).Println("💡 To resolve conflicts:")
		fmt.Println("   1. Edit conflicted files")
		fmt.Println("   2. git add <resolved-files>")
		fmt.Println("   3. git rebase --continue")
		fmt.Println()
		fmt.Println("💡 To abort rebase:")
		fmt.Println("   git rebase --abort")
		fmt.Println()
		fmt.Println("💡 Get help with conflicts:")
		fmt.Println("   git-assist rebase --conflicts")

		return fmt.Errorf("rebase failed - see above for resolution steps")
	}

	color.New(color.FgGreen, color.Bold).Println("✅ Rebase completed successfully!")
	return nil
}

// helpResolveConflicts helps resolve rebase conflicts
func helpResolveConflicts(gitRepo *git.Repository) error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("🔧 Rebase Conflict Resolution Help")
		fmt.Println()
	}

	// Check if we're in a rebase state
	if !isInRebaseState(gitRepo) {
		color.New(color.FgYellow).Println("⚠️  Not currently in a rebase state")
		return nil
	}

	// Get conflicted files
	conflictedFiles, err := getConflictedFiles(gitRepo)
	if err != nil {
		return fmt.Errorf("failed to get conflicted files: %w", err)
	}

	if len(conflictedFiles) == 0 {
		color.New(color.FgGreen).Println("✅ No conflicts found - you can continue the rebase")
		fmt.Println()
		color.New(color.FgCyan).Println("💡 Continue rebase with:")
		fmt.Println("   git rebase --continue")
		return nil
	}

	color.New(color.FgRed).Printf("⚠️  Found %d conflicted files:\n", len(conflictedFiles))
	for _, file := range conflictedFiles {
		fmt.Printf("   • %s\n", file)
	}

	fmt.Println()
	color.New(color.FgCyan).Println("🔧 Resolution Steps:")
	fmt.Println("   1. Open each conflicted file in your editor")
	fmt.Println("   2. Look for conflict markers: <<<<<<<, =======, >>>>>>>")
	fmt.Println("   3. Choose which changes to keep")
	fmt.Println("   4. Remove conflict markers")
	fmt.Println("   5. Save the file")
	fmt.Println("   6. Stage resolved files: git add <file>")
	fmt.Println("   7. Continue rebase: git rebase --continue")

	fmt.Println()
	color.New(color.FgMagenta).Println("💡 Pro Tips:")
	fmt.Println("   • Use a merge tool: git mergetool")
	fmt.Println("   • See the conflict: git diff")
	fmt.Println("   • Abort if needed: git rebase --abort")

	// Offer AI assistance if available
	if err := offerAIConflictHelp(conflictedFiles); err == nil {
		fmt.Println()
		color.New(color.FgGreen).Println("🤖 AI conflict analysis provided above")
	}

	return nil
}

// cleanupCommitHistory helps clean up commit history
func cleanupCommitHistory(gitRepo *git.Repository) error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("🧹 Commit History Cleanup")
		fmt.Println()
	}

	// Get recent commits for analysis
	commits, err := getRecentCommits(gitRepo, 10)
	if err != nil {
		return fmt.Errorf("failed to get recent commits: %w", err)
	}

	if len(commits) <= 1 {
		color.New(color.FgGreen).Println("✅ Commit history looks clean")
		return nil
	}

	// Analyze commits for cleanup opportunities
	suggestions := analyzeCommitsForCleanup(commits)

	color.New(color.FgBlue).Println("📊 Recent Commits:")
	for i, commit := range commits {
		fmt.Printf("   %d. %s - %s\n", i+1, commit.Hash[:8], commit.Message)
	}

	fmt.Println()
	color.New(color.FgMagenta).Println("💡 Cleanup Suggestions:")
	for _, suggestion := range suggestions {
		fmt.Printf("   • %s\n", suggestion)
	}

	fmt.Println()
	color.New(color.FgCyan).Println("🔧 Cleanup Options:")
	fmt.Println("   • Interactive rebase: git-assist rebase --interactive")
	fmt.Println("   • Squash commits: git rebase -i HEAD~N")
	fmt.Println("   • Amend last commit: git commit --amend")

	return nil
}

// Helper functions

// isWorkingDirectoryClean checks if working directory is clean
func isWorkingDirectoryClean(gitRepo *git.Repository) bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = gitRepo.Path()

	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(output)) == ""
}

// hasUnpushedCommits checks if there are unpushed commits
func hasUnpushedCommits(gitRepo *git.Repository) bool {
	cmd := exec.Command("git", "log", "@{u}..", "--oneline")
	cmd.Dir = gitRepo.Path()

	output, err := cmd.Output()
	if err != nil {
		return false // Assume no upstream or no unpushed commits
	}

	return strings.TrimSpace(string(output)) != ""
}

// isInRebaseState checks if repository is in rebase state
func isInRebaseState(gitRepo *git.Repository) bool {
	rebaseDirs := []string{
		gitRepo.Path() + "/.git/rebase-merge",
		gitRepo.Path() + "/.git/rebase-apply",
	}

	for _, dir := range rebaseDirs {
		if _, err := os.Stat(dir); err == nil {
			return true
		}
	}

	return false
}

// getConflictedFiles returns files with merge conflicts
func getConflictedFiles(gitRepo *git.Repository) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	cmd.Dir = gitRepo.Path()

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var files []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}

	return files, nil
}

// getCommitHistory gets commit history relative to target branch
func getCommitHistory(gitRepo *git.Repository, targetBranch string) ([]CommitInfo, error) {
	cmd := exec.Command("git", "log", "--oneline", "--format=%H|%s|%an", targetBranch+"..")
	cmd.Dir = gitRepo.Path()

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var commits []CommitInfo
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			commits = append(commits, CommitInfo{
				Hash:    parts[0],
				Message: parts[1],
				Author:  parts[2],
			})
		}
	}

	return commits, nil
}

// getRecentCommits gets recent commits
func getRecentCommits(gitRepo *git.Repository, count int) ([]CommitInfo, error) {
	cmd := exec.Command("git", "log", "--oneline", "--format=%H|%s|%an", fmt.Sprintf("-%d", count))
	cmd.Dir = gitRepo.Path()

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var commits []CommitInfo
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			commits = append(commits, CommitInfo{
				Hash:    parts[0],
				Message: parts[1],
				Author:  parts[2],
			})
		}
	}

	return commits, nil
}

// analyzeCommitsForRebase analyzes commits and provides rebase suggestions
func analyzeCommitsForRebase(commits []CommitInfo) []string {
	var suggestions []string

	if len(commits) > 5 {
		suggestions = append(suggestions, "Consider squashing some commits for cleaner history")
	}

	// Check for fixup commits
	for _, commit := range commits {
		if strings.HasPrefix(strings.ToLower(commit.Message), "fix") ||
			strings.HasPrefix(strings.ToLower(commit.Message), "fixup") {
			suggestions = append(suggestions, "Found fixup commits - consider squashing with original commits")
			break
		}
	}

	// Check for WIP commits
	for _, commit := range commits {
		if strings.Contains(strings.ToLower(commit.Message), "wip") ||
			strings.Contains(strings.ToLower(commit.Message), "work in progress") {
			suggestions = append(suggestions, "Found WIP commits - consider cleaning up commit messages")
			break
		}
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "Commit history looks good for rebasing")
	}

	return suggestions
}

// analyzeCommitsForCleanup analyzes commits for cleanup opportunities
func analyzeCommitsForCleanup(commits []CommitInfo) []string {
	var suggestions []string

	// Similar analysis as rebase but focused on cleanup
	if len(commits) > 3 {
		suggestions = append(suggestions, "Consider squashing related commits")
	}

	// Check for typo fixes
	for _, commit := range commits {
		if strings.Contains(strings.ToLower(commit.Message), "typo") ||
			strings.Contains(strings.ToLower(commit.Message), "fix") {
			suggestions = append(suggestions, "Found typo/fix commits - consider squashing with original")
			break
		}
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "Commit history is already clean")
	}

	return suggestions
}

// offerAIConflictHelp uses AI to help with conflict resolution
func offerAIConflictHelp(conflictedFiles []string) error {
	// Create AI manager
	aiManager := ai.NewManager()

	// Try to setup AI
	err := aiManager.SetupFromConfig()
	if err != nil {
		return err
	}

	// Create conflict analysis prompt
	prompt := fmt.Sprintf(`I have merge conflicts in these files during a git rebase:
%s

Please provide general guidance on:
1. How to approach resolving these conflicts
2. What to look for in the conflict markers
3. Best practices for conflict resolution
4. How to verify the resolution is correct

Keep the advice practical and beginner-friendly.`, strings.Join(conflictedFiles, "\n"))

	// Generate advice
	advice, err := aiManager.GenerateCommitMessage(prompt)
	if err != nil {
		return err
	}

	// Display AI advice
	fmt.Println()
	color.New(color.FgGreen).Println("🤖 AI Conflict Resolution Advice:")
	fmt.Println(advice)

	return nil
}
