// Package cli/completion provides enhanced autocompletion for git-assist commands
package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gajeshbhat/git-assist/internal/ai"
	"github.com/gajeshbhat/git-assist/internal/repository"
	"github.com/spf13/cobra"
)

// setupCompletionFunctions adds intelligent autocompletion to commands
func setupCompletionFunctions() {
	// Add completion for config command flags
	configCmd.RegisterFlagCompletionFunc("set-model", completeModelNames)
	configCmd.RegisterFlagCompletionFunc("pull-model", completeAvailableModels)
	configCmd.RegisterFlagCompletionFunc("set-endpoint", completeEndpoints)

	// Add completion for commit command
	commitCmd.RegisterFlagCompletionFunc("message", completeCommitMessages)
}

// completeModelNames provides completion for installed model names
func completeModelNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Try to get installed models
	modelManager := ai.NewModelManager("http://localhost:11434")
	models, err := modelManager.ListInstalledModels()
	if err != nil {
		// If we can't get models, return common ones
		return []string{
			"codellama:7b",
			"codellama:13b",
			"mistral:7b",
			"llama3:8b",
		}, cobra.ShellCompDirectiveNoFileComp
	}

	// Extract model names
	var modelNames []string
	for _, model := range models {
		if strings.HasPrefix(model.Name, toComplete) {
			modelNames = append(modelNames, model.Name)
		}
	}

	return modelNames, cobra.ShellCompDirectiveNoFileComp
}

// completeAvailableModels provides completion for models available for download
func completeAvailableModels(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	modelManager := ai.NewModelManager("http://localhost:11434")
	recommendations := modelManager.GetRecommendedModels()

	var modelNames []string

	// Add all recommended models
	for _, category := range recommendations {
		for _, model := range category {
			if strings.HasPrefix(model.Name, toComplete) {
				modelNames = append(modelNames, model.Name)
			}
		}
	}

	return modelNames, cobra.ShellCompDirectiveNoFileComp
}

// completeEndpoints provides completion for common AI endpoints
func completeEndpoints(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	endpoints := []string{
		"http://localhost:11434",
		"http://localhost:8080",
		"http://localhost:3000",
		"https://api.openai.com",
		"https://api.anthropic.com",
	}

	var matches []string
	for _, endpoint := range endpoints {
		if strings.HasPrefix(endpoint, toComplete) {
			matches = append(matches, endpoint)
		}
	}

	return matches, cobra.ShellCompDirectiveNoFileComp
}

// completeCommitMessages provides intelligent completion for commit messages
func completeCommitMessages(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Try to analyze current changes for suggestions
	diff, err := getGitDiff()
	if err != nil {
		// Return common commit message patterns
		return []string{
			"feat: ",
			"fix: ",
			"docs: ",
			"style: ",
			"refactor: ",
			"test: ",
			"chore: ",
		}, cobra.ShellCompDirectiveNoSpace
	}

	// Try to get context-aware suggestions
	contextAnalyzer := repository.NewContextAnalyzer(cwd)
	commitContext, err := contextAnalyzer.AnalyzeCommitContext(diff)
	if err != nil {
		// Fallback to basic patterns
		return []string{
			"feat: ",
			"fix: ",
			"docs: ",
		}, cobra.ShellCompDirectiveNoSpace
	}

	// Generate contextual suggestions
	var suggestions []string

	if commitContext.ChangeType != "" && commitContext.Scope != "" {
		suggestions = append(suggestions, fmt.Sprintf("%s(%s): ", commitContext.ChangeType, commitContext.Scope))
	} else if commitContext.ChangeType != "" {
		suggestions = append(suggestions, fmt.Sprintf("%s: ", commitContext.ChangeType))
	}

	// Add file-specific suggestions
	if len(commitContext.ChangedFiles) == 1 {
		file := commitContext.ChangedFiles[0]
		switch file.Status {
		case "added":
			suggestions = append(suggestions, fmt.Sprintf("feat: add %s", file.Path))
		case "modified":
			suggestions = append(suggestions, fmt.Sprintf("fix: update %s", file.Path))
		case "deleted":
			suggestions = append(suggestions, fmt.Sprintf("refactor: remove %s", file.Path))
		}
	}

	// Filter suggestions based on what user has typed
	var matches []string
	for _, suggestion := range suggestions {
		if strings.HasPrefix(suggestion, toComplete) {
			matches = append(matches, suggestion)
		}
	}

	if len(matches) == 0 {
		// Return all suggestions if no matches
		return suggestions, cobra.ShellCompDirectiveNoSpace
	}

	return matches, cobra.ShellCompDirectiveNoSpace
}

// getGitDiff gets the current git diff for staged changes
func getGitDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// addCompletionExamples documents completion examples
// Note: Completion command is automatically generated by Cobra
func addCompletionExamples() {
	// Examples for setting up completion:
	//
	// Zsh (macOS):
	//   git-assist completion zsh > $(brew --prefix)/share/zsh/site-functions/_git-assist
	//
	// Bash (Linux):
	//   git-assist completion bash > /etc/bash_completion.d/git-assist
	//
	// Test in current session:
	//   source <(git-assist completion zsh)
}

// Enhanced completion features that could be added:

// 1. File path completion for specific file types
func completeSourceFiles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Complete with .go, .js, .py files etc.
	return nil, cobra.ShellCompDirectiveFilterFileExt
}

// 2. Branch name completion
func completeBranchNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Get git branches and complete
	return nil, cobra.ShellCompDirectiveNoFileComp
}

// 3. Commit hash completion
func completeCommitHashes(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Get recent commit hashes
	return nil, cobra.ShellCompDirectiveNoFileComp
}

// 4. Configuration key completion
func completeConfigKeys(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	keys := []string{
		"ai.provider",
		"ai.model",
		"ai.endpoint",
		"practices.style",
		"practices.max_length",
		"preferences.output_format",
		"preferences.color_output",
	}

	var matches []string
	for _, key := range keys {
		if strings.HasPrefix(key, toComplete) {
			matches = append(matches, key)
		}
	}

	return matches, cobra.ShellCompDirectiveNoFileComp
}
