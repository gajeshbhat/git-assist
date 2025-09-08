package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/gajeshbhat/git-assist/internal/ai"
	"github.com/gajeshbhat/git-assist/internal/repository"
	"github.com/spf13/cobra"
)

// commitCmd represents the commit command
var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Generate AI-powered commit messages",
	Long: `Generate AI-powered commit messages based on staged changes.

This command analyzes your staged changes and generates appropriate
commit messages following your configured practices and rules.

The command will:
• Analyze staged changes (git diff --cached)
• Generate commit message using AI
• Validate against configured rules
• Commit with the generated message

Examples:
  git-assist commit                  # Interactive review → commit
  git-assist commit --auto-commit    # Generate and commit immediately
  git-assist commit --dry-run        # Generate message only (no commit)
  git-assist commit --with-context   # Use repository context for better messages
  git-assist commit --no-ai          # Use rule-based generation
  git-assist commit --force-ai       # Require AI (fail if unavailable)
  git-assist commit -m "fix: bug"    # Use custom message`,
	RunE: runCommit,
}

// Command flags
var (
	dryRun            bool
	commitInteractive bool
	commitMessage     string
	regenerate        bool
	noAI              bool
	forceAI           bool
	autoCommit        bool
	skipInteractive   bool
	withContext       bool
)

func init() {
	rootCmd.AddCommand(commitCmd)

	// Add flags
	commitCmd.Flags().BoolVar(&dryRun, "dry-run", false, "generate message without committing")
	commitCmd.Flags().BoolVarP(&commitInteractive, "interactive", "i", false, "interactive commit builder")
	commitCmd.Flags().StringVarP(&commitMessage, "message", "m", "", "use provided message instead of generating")
	commitCmd.Flags().BoolVar(&regenerate, "regenerate", false, "regenerate message if one exists")

	// AI control flags
	commitCmd.Flags().BoolVar(&noAI, "no-ai", false, "use rule-based generation instead of AI")
	commitCmd.Flags().BoolVar(&forceAI, "force-ai", false, "require AI generation (fail if AI unavailable)")

	// Interactive control flags
	commitCmd.Flags().BoolVar(&autoCommit, "auto-commit", false, "automatically commit without confirmation")
	commitCmd.Flags().BoolVar(&skipInteractive, "skip-interactive", false, "skip interactive review (same as --auto-commit)")

	// Context flags
	commitCmd.Flags().BoolVar(&withContext, "with-context", false, "use repository context for enhanced commit messages")
}

func runCommit(cmd *cobra.Command, args []string) error {
	// Check if we're in a Git repository
	if !isGitRepository() {
		return fmt.Errorf("not in a Git repository")
	}

	// Check if git-assist is initialized
	if !isGitAssistInitialized() {
		return fmt.Errorf("git-assist not initialized. Run 'git-assist init' first")
	}

	// Check for staged changes
	hasStagedChanges, err := checkStagedChanges()
	if err != nil {
		return fmt.Errorf("failed to check staged changes: %w", err)
	}

	if !hasStagedChanges {
		return fmt.Errorf("no staged changes found. Use 'git add' to stage changes first")
	}

	// Get staged changes
	diff, err := getStagedDiff()
	if err != nil {
		return fmt.Errorf("failed to get staged changes: %w", err)
	}

	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("🤖 Analyzing staged changes...")
		fmt.Println()
	}

	// Generate commit message
	var generatedMessage string
	if commitMessage != "" {
		generatedMessage = commitMessage
	} else {
		if withContext {
			generatedMessage, err = generateContextAwareCommitMessage(diff)
		} else {
			generatedMessage, err = generateCommitMessage(diff)
		}
		if err != nil {
			return fmt.Errorf("failed to generate commit message: %w", err)
		}
	}

	// If dry-run, just show the message
	if dryRun {
		if !quiet {
			color.New(color.FgGreen, color.Bold).Println("📝 Generated commit message:")
			fmt.Println()
			color.New(color.FgWhite, color.Bold).Println(generatedMessage)
			fmt.Println()
			color.New(color.FgYellow).Println("🔍 Dry run mode - no commit created")
		}
		return nil
	}

	// Determine if we should use interactive mode
	useInteractive := !autoCommit && !skipInteractive && !quiet

	if useInteractive {
		// Interactive review mode
		return runInteractiveCommitReview(generatedMessage, diff)
	} else {
		// Auto-commit mode
		if !quiet {
			color.New(color.FgGreen, color.Bold).Println("📝 Generated commit message:")
			fmt.Println()
			color.New(color.FgWhite, color.Bold).Println(generatedMessage)
			fmt.Println()
		}

		// Commit with the generated message
		if err := performCommit(generatedMessage); err != nil {
			return fmt.Errorf("failed to commit: %w", err)
		}

		if !quiet {
			color.New(color.FgGreen, color.Bold).Println("✅ Commit created successfully!")
		}

		return nil
	}
}

// isGitAssistInitialized checks if git-assist is initialized in the repository
func isGitAssistInitialized() bool {
	_, err := os.Stat(".git/git-assist/config.json")
	return err == nil
}

// checkStagedChanges checks if there are any staged changes
func checkStagedChanges() (bool, error) {
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	err := cmd.Run()

	// git diff --quiet returns 0 if no differences, 1 if differences exist
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode() == 1, nil
		}
		return false, err
	}

	return false, nil // No staged changes
}

// getStagedDiff gets the diff of staged changes
func getStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// generateCommitMessage generates a commit message using AI or rules based on flags
func generateCommitMessage(diff string) (string, error) {
	// Check for conflicting flags
	if noAI && forceAI {
		return "", fmt.Errorf("cannot use both --no-ai and --force-ai flags")
	}

	// If user explicitly wants rule-based generation
	if noAI {
		if !quiet {
			color.New(color.FgBlue).Println("🔧 Using rule-based generation (--no-ai)")
		}
		return generateRuleBasedCommitMessage(diff)
	}

	// Try AI generation
	aiManager := ai.NewManager()
	err := aiManager.SetupFromConfig()

	if err != nil {
		// AI setup failed
		if forceAI {
			return "", fmt.Errorf("AI generation required (--force-ai) but AI setup failed: %w", err)
		}

		// Fall back to rule-based
		if !quiet {
			color.New(color.FgYellow).Printf("🤖➡️🔧 AI not available (%v), using rule-based generation\n", err)
		}
		return generateRuleBasedCommitMessage(diff)
	}

	// Generate commit message using AI
	message, err := aiManager.GenerateCommitMessage(diff)
	if err != nil {
		// AI generation failed
		if forceAI {
			return "", fmt.Errorf("AI generation required (--force-ai) but failed: %w", err)
		}

		// Fall back to rule-based
		if !quiet {
			color.New(color.FgYellow).Printf("🤖➡️🔧 AI generation failed (%v), using rule-based generation\n", err)
		}
		return generateRuleBasedCommitMessage(diff)
	}

	// AI generation successful
	if !quiet {
		color.New(color.FgGreen).Println("🤖 Generated using AI")
	}
	return message, nil
}

// generateContextAwareCommitMessage generates a commit message using repository context
func generateContextAwareCommitMessage(diff string) (string, error) {
	if !quiet {
		color.New(color.FgCyan).Println("🔍 Analyzing repository context...")
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Create context analyzer
	contextAnalyzer := repository.NewContextAnalyzer(cwd)

	// Analyze commit context
	commitContext, err := contextAnalyzer.AnalyzeCommitContext(diff)
	if err != nil {
		// If context analysis fails, fall back to regular generation
		if !quiet {
			color.New(color.FgYellow).Printf("⚠️  Context analysis failed (%v), using standard generation\n", err)
		}
		return generateCommitMessage(diff)
	}

	// Show context information
	if !quiet && verbose {
		color.New(color.FgBlue).Printf("📊 Context: %s\n", commitContext.Context)
		if len(commitContext.Suggestions) > 0 {
			color.New(color.FgMagenta).Println("💡 Suggestions:")
			for _, suggestion := range commitContext.Suggestions {
				fmt.Printf("   • %s\n", suggestion)
			}
		}
	}

	// Check for conflicting flags
	if noAI && forceAI {
		return "", fmt.Errorf("cannot use both --no-ai and --force-ai flags")
	}

	// If user explicitly wants rule-based generation
	if noAI {
		if !quiet {
			color.New(color.FgBlue).Println("🔧 Using rule-based generation with context (--no-ai)")
		}
		return generateContextualRuleBasedMessage(commitContext)
	}

	// Try AI generation with context
	aiManager := ai.NewManager()
	err = aiManager.SetupFromConfig()

	if err != nil {
		// AI setup failed
		if forceAI {
			return "", fmt.Errorf("AI generation required (--force-ai) but AI setup failed: %w", err)
		}

		// Fall back to contextual rule-based
		if !quiet {
			color.New(color.FgYellow).Printf("🤖➡️🔧 AI not available (%v), using contextual rule-based generation\n", err)
		}
		return generateContextualRuleBasedMessage(commitContext)
	}

	// Generate commit message using AI with context
	message, err := generateAIMessageWithContext(aiManager, commitContext)
	if err != nil {
		// AI generation failed
		if forceAI {
			return "", fmt.Errorf("AI generation required (--force-ai) but failed: %w", err)
		}

		// Fall back to contextual rule-based
		if !quiet {
			color.New(color.FgYellow).Printf("🤖➡️🔧 AI generation failed (%v), using contextual rule-based generation\n", err)
		}
		return generateContextualRuleBasedMessage(commitContext)
	}

	// AI generation successful
	if !quiet {
		color.New(color.FgGreen).Println("🤖 Generated using AI with repository context")
	}
	return message, nil
}

// generateContextualRuleBasedMessage generates a commit message using rules and context
func generateContextualRuleBasedMessage(context *repository.CommitContext) (string, error) {
	// Use context information to create a better rule-based message
	var parts []string

	// Add type and scope from context
	if context.ChangeType != "" {
		if context.Scope != "" {
			parts = append(parts, fmt.Sprintf("%s(%s):", context.ChangeType, context.Scope))
		} else {
			parts = append(parts, fmt.Sprintf("%s:", context.ChangeType))
		}
	}

	// Generate description based on changed files
	if len(context.ChangedFiles) == 1 {
		file := context.ChangedFiles[0]
		switch file.Status {
		case "added":
			parts = append(parts, fmt.Sprintf("add %s", file.Path))
		case "deleted":
			parts = append(parts, fmt.Sprintf("remove %s", file.Path))
		case "modified":
			parts = append(parts, fmt.Sprintf("update %s", file.Path))
		case "renamed":
			parts = append(parts, fmt.Sprintf("rename %s", file.Path))
		}
	} else {
		// Multiple files
		fileTypes := make(map[string]int)
		for _, file := range context.ChangedFiles {
			fileTypes[file.Type]++
		}

		var descriptions []string
		for fileType, count := range fileTypes {
			if count == 1 {
				descriptions = append(descriptions, fmt.Sprintf("1 %s file", fileType))
			} else {
				descriptions = append(descriptions, fmt.Sprintf("%d %s files", count, fileType))
			}
		}

		parts = append(parts, fmt.Sprintf("update %s", strings.Join(descriptions, " and ")))
	}

	return strings.Join(parts, " "), nil
}

// generateAIMessageWithContext generates a commit message using AI with repository context
func generateAIMessageWithContext(aiManager *ai.Manager, context *repository.CommitContext) (string, error) {
	// Create enhanced prompt with context
	prompt := createContextualPrompt(context)

	// Generate message using AI
	message, err := aiManager.GenerateCommitMessage(prompt)
	if err != nil {
		return "", err
	}

	return message, nil
}

// createContextualPrompt creates an enhanced prompt with repository context
func createContextualPrompt(context *repository.CommitContext) string {
	var prompt strings.Builder

	// Add repository context
	if context.Repository != nil {
		prompt.WriteString(fmt.Sprintf("Repository: %s (%s project)\n",
			context.Repository.Metadata.Name, context.Repository.Metadata.MainLanguage))
	}

	// Add change context
	prompt.WriteString(fmt.Sprintf("Change Type: %s\n", context.ChangeType))
	if context.Scope != "" {
		prompt.WriteString(fmt.Sprintf("Scope: %s\n", context.Scope))
	}
	prompt.WriteString(fmt.Sprintf("Impact: %s\n", context.Impact))

	// Add file information
	prompt.WriteString(fmt.Sprintf("Files changed: %d\n", len(context.ChangedFiles)))
	for _, file := range context.ChangedFiles {
		prompt.WriteString(fmt.Sprintf("- %s (%s, %s): +%d -%d lines\n",
			file.Path, file.Status, file.Language, file.LinesAdded, file.LinesDeleted))
	}

	// Add suggestions
	if len(context.Suggestions) > 0 {
		prompt.WriteString("\nSuggestions:\n")
		for _, suggestion := range context.Suggestions {
			prompt.WriteString(fmt.Sprintf("- %s\n", suggestion))
		}
	}

	prompt.WriteString("\nGenerate a concise, conventional commit message based on this context.")

	return prompt.String()
}

// generateRuleBasedCommitMessage generates a commit message using simple rules (fallback)
func generateRuleBasedCommitMessage(diff string) (string, error) {
	// Simple analysis of the diff
	lines := strings.Split(diff, "\n")
	addedLines := 0
	deletedLines := 0
	modifiedFiles := make(map[string]bool)

	for _, line := range lines {
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			if strings.Contains(line, "/") {
				parts := strings.Split(line, "/")
				if len(parts) > 0 {
					filename := parts[len(parts)-1]
					modifiedFiles[filename] = true
				}
			}
		} else if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			addedLines++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			deletedLines++
		}
	}

	// Generate a simple commit message based on analysis
	var commitType string
	var description string

	if len(modifiedFiles) == 1 {
		for filename := range modifiedFiles {
			if strings.Contains(filename, "test") {
				commitType = "test"
				description = fmt.Sprintf("update tests in %s", filename)
			} else if strings.Contains(filename, "doc") || strings.Contains(filename, "README") {
				commitType = "docs"
				description = fmt.Sprintf("update documentation in %s", filename)
			} else {
				commitType = "feat"
				description = fmt.Sprintf("update %s", filename)
			}
		}
	} else {
		if addedLines > deletedLines*2 {
			commitType = "feat"
			description = fmt.Sprintf("add new functionality (%d files changed)", len(modifiedFiles))
		} else if deletedLines > addedLines*2 {
			commitType = "refactor"
			description = fmt.Sprintf("remove unused code (%d files changed)", len(modifiedFiles))
		} else {
			commitType = "chore"
			description = fmt.Sprintf("update multiple files (%d files changed)", len(modifiedFiles))
		}
	}

	return fmt.Sprintf("%s: %s", commitType, description), nil
}

// runInteractiveCommitReview runs the interactive commit review process
func runInteractiveCommitReview(message, diff string) error {
	// Determine generation mode for display
	var generationMode string
	if noAI {
		generationMode = "rules"
	} else if commitMessage != "" {
		generationMode = "custom"
	} else {
		generationMode = "ai"
	}

	// Set up review options
	options := CommitReviewOptions{
		InitialMessage:  message,
		Diff:            diff,
		GenerationMode:  generationMode,
		AllowRegenerate: false, // Simplified: no regeneration
		AllowEdit:       true,
	}

	// Run interactive review
	result, err := InteractiveCommitReview(options)
	if err != nil {
		return fmt.Errorf("interactive review failed: %w", err)
	}

	// Handle the result
	switch result.Action {
	case "commit":
		// Commit with the final message
		if err := performCommit(result.Message); err != nil {
			return fmt.Errorf("failed to commit: %w", err)
		}

		if !quiet {
			color.New(color.FgGreen, color.Bold).Println("✅ Commit created successfully!")
		}

	case "cancel":
		if !quiet {
			color.New(color.FgYellow).Println("❌ Commit cancelled by user")
		}

	default:
		return fmt.Errorf("unknown action: %s", result.Action)
	}

	return nil
}

// performCommit performs the actual git commit
func performCommit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("git commit failed: %s", string(output))
	}

	if verbose {
		fmt.Printf("Git output: %s\n", string(output))
	}

	return nil
}
