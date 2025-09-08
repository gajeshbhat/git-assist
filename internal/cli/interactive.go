// Package cli/interactive provides interactive commit message review functionality
package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

// CommitReviewOptions contains options for the interactive commit review
type CommitReviewOptions struct {
	InitialMessage  string
	Diff            string
	GenerationMode  string // "ai", "rules", "custom"
	AllowRegenerate bool
	AllowEdit       bool
}

// CommitReviewResult contains the result of the interactive review
type CommitReviewResult struct {
	Message    string
	Action     string // "commit", "cancel", "edit"
	Regenerate bool
}

// InteractiveCommitReview handles the simplified interactive commit message review
func InteractiveCommitReview(options CommitReviewOptions) (*CommitReviewResult, error) {
	reader := bufio.NewReader(os.Stdin)
	currentMessage := options.InitialMessage

	// Show the commit message
	displayCommitMessage(currentMessage, options.GenerationMode)

	// Show simple options
	fmt.Println()
	color.New(color.FgYellow, color.Bold).Println("Options:")
	color.New(color.FgGreen).Println("  [a] Accept and commit (default)")
	color.New(color.FgRed).Println("  [r] Reject and cancel")
	color.New(color.FgCyan).Println("  [e] Edit message")

	// Get user input
	fmt.Print("\nChoose an option [a/r/e]: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "a", "accept", "":
		// Accept and commit
		return &CommitReviewResult{
			Message: currentMessage,
			Action:  "commit",
		}, nil

	case "r", "reject", "cancel":
		// Cancel the commit
		return &CommitReviewResult{
			Action: "cancel",
		}, nil

	case "e", "edit":
		// Edit the message
		editedMessage, err := editCommitMessage(currentMessage)
		if err != nil {
			return nil, fmt.Errorf("failed to edit message: %w", err)
		}

		if editedMessage != "" {
			return &CommitReviewResult{
				Message: editedMessage,
				Action:  "commit",
			}, nil
		} else {
			// If editing was cancelled, go back to original message
			return &CommitReviewResult{
				Message: currentMessage,
				Action:  "commit",
			}, nil
		}

	default:
		// Default to accept for any other input
		return &CommitReviewResult{
			Message: currentMessage,
			Action:  "commit",
		}, nil
	}
}

// displayCommitMessage shows the current commit message with simple formatting
func displayCommitMessage(message, mode string) {
	color.New(color.FgCyan, color.Bold).Println("📝 Generated Commit Message:")
	fmt.Println()

	// Show generation mode
	switch mode {
	case "ai":
		color.New(color.FgGreen).Println("🤖 Generated using AI")
	case "rules":
		color.New(color.FgBlue).Println("🔧 Generated using rules")
	case "custom":
		color.New(color.FgMagenta).Println("✏️  Custom message")
	}

	fmt.Println()
	color.New(color.FgWhite, color.Bold).Println(message)
}

// editCommitMessage opens an editor to edit the commit message
func editCommitMessage(currentMessage string) (string, error) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "git-assist-commit-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write current message to file
	if _, err := tmpFile.WriteString(currentMessage); err != nil {
		return "", fmt.Errorf("failed to write to temp file: %w", err)
	}
	tmpFile.Close()

	// Get editor from environment or use default
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano" // Default to nano as it's user-friendly
	}

	// Open editor
	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor failed: %w", err)
	}

	// Read the edited content
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read edited file: %w", err)
	}

	return strings.TrimSpace(string(content)), nil
}
