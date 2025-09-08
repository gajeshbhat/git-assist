// Package cli provides the command-line interface for git-assist.
//
// This package uses the Cobra library to create a structured CLI with
// subcommands, flags, and help text. It follows CLI best practices
// and provides both interactive and flag-driven interfaces.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// These variables are set at build time using -ldflags
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// Global flags
var (
	cfgFile    string
	outputJSON bool
	quiet      bool
	verbose    bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "git-assist",
	Short: "AI-powered Git assistant for enhanced development workflow",
	Long: `git-assist - AI-powered Git assistant

An intelligent Git companion that enhances your development workflow with
AI-generated commit messages, repository analysis, smart branch management,
and guided git operations.

Features:
  - AI-generated commit messages based on staged changes
  - Comprehensive repository analysis and health insights
  - Intelligent branch management with safe cleanup
  - Smart history navigation and search
  - AI-guided rebasing with safety checks
  - Automatic shell completion setup

Examples:
  git-assist commit                  # Generate AI commit message
  git-assist analyze                 # Analyze repository structure
  git-assist branch cleanup          # Clean up merged branches
  git-assist history --search auth   # Search commit history
  git-assist rebase --safe main      # Safe rebase with checks
  git-assist config --setup-completion  # Setup shell completion

Get started:
  git-assist config --install-ollama    # Install AI backend
  git-assist config --setup-completion  # Setup tab completion
  git-assist init                        # Initialize in repository`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, show help
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.git-assist/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&outputJSON, "json", false, "output in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet mode (minimal output)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Version flag
	rootCmd.Flags().BoolP("version", "V", false, "show version information")

	// Custom version handling
	rootCmd.SetVersionTemplate(getVersionInfo())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".git-assist" (without extension).
		viper.AddConfigPath(home + "/.git-assist")
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && !quiet {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// getVersionInfo returns formatted version information
func getVersionInfo() string {
	return fmt.Sprintf(`git-assist version %s
Commit: %s
Built: %s
Go version: %s
Platform: %s/%s
`, version, commit, date,
		"go1.25.1",        // We'll make this dynamic later
		"darwin", "arm64") // We'll make this dynamic later
}

// SetVersion sets the version information (called from main)
func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
	rootCmd.Version = v
}
