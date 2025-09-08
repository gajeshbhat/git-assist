package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize git-assist in the current repository",
	Long: `Initialize git-assist in the current Git repository.

This command sets up git-assist configuration and creates necessary
directories and files for optimal operation. It will:

• Verify you're in a Git repository
• Create .git/git-assist/ directory structure
• Set up default configuration
• Optionally run interactive setup wizard

Examples:
  git-assist init                    # Basic initialization
  git-assist init --interactive      # Run setup wizard
  git-assist init --model ollama     # Set specific model
  git-assist init --practices conventional  # Set commit style`,
	RunE: runInit,
}

// Command flags
var (
	interactive bool
	modelType   string
	practices   string
	force       bool
)

func init() {
	rootCmd.AddCommand(initCmd)

	// Add flags
	initCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "run interactive setup wizard")
	initCmd.Flags().StringVar(&modelType, "model", "", "AI model type (ollama, bundled, remote)")
	initCmd.Flags().StringVar(&practices, "practices", "", "commit practices (conventional, angular, gitmoji)")
	initCmd.Flags().BoolVar(&force, "force", false, "force initialization even if already initialized")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Check if we're in a Git repository
	if !isGitRepository() {
		return fmt.Errorf("not in a Git repository. Please run this command from within a Git repository")
	}

	// Check if already initialized
	gitAssistDir := ".git/git-assist"
	if _, err := os.Stat(gitAssistDir); err == nil && !force {
		return fmt.Errorf("git-assist is already initialized in this repository. Use --force to reinitialize")
	}

	// Print initialization header
	if !quiet {
		color.New(color.FgGreen, color.Bold).Println("🚀 Initializing git-assist...")
		fmt.Println()
	}

	// Create directory structure
	if err := createDirectoryStructure(gitAssistDir); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Create default configuration
	if err := createDefaultConfig(gitAssistDir); err != nil {
		return fmt.Errorf("failed to create default configuration: %w", err)
	}

	// Run interactive setup if requested
	if interactive {
		if err := runInteractiveSetup(gitAssistDir); err != nil {
			return fmt.Errorf("interactive setup failed: %w", err)
		}
	} else {
		// Apply command-line options
		if err := applyCommandLineOptions(gitAssistDir); err != nil {
			return fmt.Errorf("failed to apply options: %w", err)
		}
	}

	// Success message
	if !quiet {
		fmt.Println()
		color.New(color.FgGreen, color.Bold).Println("✅ git-assist initialized successfully!")
		fmt.Println()
		color.New(color.FgCyan).Println("Next steps:")
		fmt.Println("  • Run 'git-assist config' to customize settings")
		fmt.Println("  • Try 'git-assist commit' to generate your first AI commit message")
		fmt.Println("  • Use 'git-assist --help' to see all available commands")
	}

	return nil
}

// isGitRepository checks if the current directory is a Git repository
func isGitRepository() bool {
	_, err := os.Stat(".git")
	return err == nil
}

// createDirectoryStructure creates the necessary directory structure
func createDirectoryStructure(baseDir string) error {
	dirs := []string{
		baseDir,
		filepath.Join(baseDir, "cache"),
		filepath.Join(baseDir, "cache", "embeddings"),
		filepath.Join(baseDir, "cache", "analysis"),
		filepath.Join(baseDir, "rules"),
		filepath.Join(baseDir, "rules", "custom"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		if verbose {
			fmt.Printf("Created directory: %s\n", dir)
		}
	}

	return nil
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig(baseDir string) error {
	configPath := filepath.Join(baseDir, "config.json")

	defaultConfig := `{
  "version": "1.0",
  "models": [],
  "practices": {
    "industry": "conventional",
    "custom_file": "",
    "rules": ["require-type", "max-length-50"]
  },
  "preferences": {
    "auto_stage": false,
    "explain_commands": true,
    "output_format": "text",
    "color": true
  },
  "created_at": "` + fmt.Sprintf("%d", os.Getpid()) + `"
}`

	if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	if verbose {
		fmt.Printf("Created configuration: %s\n", configPath)
	}

	return nil
}

// runInteractiveSetup runs the interactive setup wizard
func runInteractiveSetup(baseDir string) error {
	if !quiet {
		color.New(color.FgYellow).Println("🧙 Interactive setup wizard")
		fmt.Println("(This will be implemented in a future version)")
	}
	return nil
}

// applyCommandLineOptions applies options provided via command line flags
func applyCommandLineOptions(baseDir string) error {
	if modelType != "" {
		if verbose {
			fmt.Printf("Setting model type: %s\n", modelType)
		}
		// TODO: Update configuration with model type
	}

	if practices != "" {
		if verbose {
			fmt.Printf("Setting practices: %s\n", practices)
		}
		// TODO: Update configuration with practices
	}

	return nil
}
