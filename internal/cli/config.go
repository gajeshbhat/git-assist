package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/gajeshbhat/git-assist/internal/ai"
	"github.com/gajeshbhat/git-assist/internal/config"
	"github.com/gajeshbhat/git-assist/internal/repository"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure git-assist settings",
	Long: `Configure git-assist settings including AI models, commit practices, and preferences.

This command helps you set up and manage your git-assist configuration.
You can configure AI models, set commit message practices, and customize
various preferences.

Examples:
  git-assist config                      # Show current configuration
  git-assist config --test-ai           # Test AI connection
  git-assist config --list-models       # List recommended models

AI Setup (run these commands in order):
  git-assist config --install-ollama    # 1. Install Ollama
  git-assist config --start-service     # 2. Start background service
  git-assist config --pull-model codellama:7b  # 3. Download a model
  git-assist config --set-model codellama:7b   # 4. Set as default
  git-assist config --test-ai           # 5. Verify setup

Shell Completion Setup:
  git-assist config --setup-completion  # Automatically setup tab completion`,
	RunE: runConfig,
}

// Command flags
var (
	testAI          bool
	listModels      bool
	setModel        string
	setEndpoint     string
	installOllama   bool
	startService    bool
	stopService     bool
	pullModel       string
	listInstalled   bool
	indexRepo       bool
	showIndex       bool
	setupCompletion bool
)

func init() {
	rootCmd.AddCommand(configCmd)

	// Add flags
	configCmd.Flags().BoolVar(&testAI, "test-ai", false, "test AI connection")
	configCmd.Flags().BoolVar(&listModels, "list-models", false, "list recommended AI models")
	configCmd.Flags().StringVar(&setModel, "set-model", "", "set AI model (e.g., codellama:7b)")
	configCmd.Flags().StringVar(&setEndpoint, "set-endpoint", "", "set AI endpoint (e.g., http://localhost:11434)")

	// Automation flags
	configCmd.Flags().BoolVar(&installOllama, "install-ollama", false, "automatically install Ollama")
	configCmd.Flags().BoolVar(&startService, "start-service", false, "start Ollama service in background")
	configCmd.Flags().BoolVar(&stopService, "stop-service", false, "stop Ollama service")
	configCmd.Flags().StringVar(&pullModel, "pull-model", "", "download and install a model (e.g., codellama:7b)")
	configCmd.Flags().BoolVar(&listInstalled, "list-installed", false, "list installed models")

	// Repository indexing flags
	configCmd.Flags().BoolVar(&indexRepo, "index-repo", false, "index repository for context-aware commits")
	configCmd.Flags().BoolVar(&showIndex, "show-index", false, "show repository index information")

	// Shell completion setup
	configCmd.Flags().BoolVar(&setupCompletion, "setup-completion", false, "automatically setup shell autocompletion")

	// Setup intelligent autocompletion
	setupCompletionFunctions()
}

func runConfig(cmd *cobra.Command, args []string) error {
	// Check if we're in a Git repository
	if !isGitRepository() {
		return fmt.Errorf("not in a Git repository. Run 'git-assist init' first")
	}

	// Check if git-assist is initialized
	if !isGitAssistInitialized() {
		return fmt.Errorf("git-assist not initialized. Run 'git-assist init' first")
	}

	// Handle specific flags
	if installOllama {
		return runOllamaInstallation()
	}

	if startService {
		return startOllamaService()
	}

	if stopService {
		return stopOllamaService()
	}

	if pullModel != "" {
		return pullModelCommand(pullModel)
	}

	if listInstalled {
		return listInstalledModels()
	}

	if indexRepo {
		return indexRepository()
	}

	if showIndex {
		return showRepositoryIndex()
	}

	if setupCompletion {
		return setupShellCompletion()
	}

	if listModels {
		return showRecommendedModels()
	}

	if testAI {
		return testAIConnection()
	}

	if setModel != "" || setEndpoint != "" {
		return updateAIConfig(setModel, setEndpoint)
	}

	// Default: show current configuration
	return showCurrentConfig()
}

// showRecommendedModels displays recommended AI models
func showRecommendedModels() error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("🤖 Recommended AI Models")
		fmt.Println()
	}

	models := ai.RecommendedModels()

	// Code-focused models
	color.New(color.FgGreen, color.Bold).Println("📝 Code-focused models (recommended):")
	for _, model := range models["code"] {
		fmt.Printf("  • %s\n", model)
	}
	fmt.Println()

	// General models
	color.New(color.FgBlue, color.Bold).Println("🧠 General-purpose models:")
	for _, model := range models["general"] {
		fmt.Printf("  • %s\n", model)
	}
	fmt.Println()

	// Large models
	color.New(color.FgYellow, color.Bold).Println("🚀 High-quality models (require more resources):")
	for _, model := range models["large"] {
		fmt.Printf("  • %s\n", model)
	}
	fmt.Println()

	if !quiet {
		color.New(color.FgCyan).Println("💡 To use a model:")
		fmt.Println("  1. Install Ollama: git-assist config --install-ollama")
		fmt.Printf("  2. Pull a model: git-assist config --pull-model %s\n", ai.SuggestModel())
		fmt.Printf("  3. Configure git-assist: git-assist config --set-model %s\n", ai.SuggestModel())
	}

	return nil
}

// testAIConnection tests the connection to the configured AI provider
func testAIConnection() error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("🔍 Testing AI connection...")
		fmt.Println()
	}

	// Create AI manager and test connection
	aiManager := ai.NewManager()

	// Try to setup with default configuration
	err := aiManager.SetupOllama(ai.SuggestModel(), "http://localhost:11434")
	if err != nil {
		color.New(color.FgRed).Printf("❌ AI connection failed: %v\n", err)
		fmt.Println()
		color.New(color.FgYellow).Println("💡 To fix this:")
		fmt.Println("  1. Install Ollama: git-assist config --install-ollama")
		fmt.Println("  2. Start service: git-assist config --start-service")
		fmt.Printf("  3. Pull a model: git-assist config --pull-model %s\n", ai.SuggestModel())
		return nil
	}

	// Test if AI is available
	if !aiManager.IsAvailable() {
		color.New(color.FgRed).Println("❌ No AI providers available")
		return nil
	}

	// Get model info
	modelInfo := aiManager.GetPrimaryModel()
	if modelInfo != nil {
		color.New(color.FgGreen).Printf("✅ AI connection successful!\n")
		fmt.Printf("   Model: %s (%s)\n", modelInfo.Name, modelInfo.Size)
		fmt.Printf("   Provider: %s\n", modelInfo.Provider)
	}

	// Test commit message generation
	if !quiet {
		fmt.Println()
		color.New(color.FgBlue).Println("🧪 Testing commit message generation...")

		testDiff := `diff --git a/test.txt b/test.txt
new file mode 100644
index 0000000..ce01362
--- /dev/null
+++ b/test.txt
@@ -0,0 +1 @@
+hello world`

		message, err := aiManager.GenerateCommitMessage(testDiff)
		if err != nil {
			color.New(color.FgYellow).Printf("⚠️  Commit generation test failed: %v\n", err)
		} else {
			color.New(color.FgGreen).Printf("✅ Test commit message: %s\n", message)
		}
	}

	return nil
}

// updateAIConfig updates the AI configuration
func updateAIConfig(model, endpoint string) error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("⚙️  Updating AI configuration...")
		fmt.Println()
	}

	// Create config manager
	configManager, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	// Validate model name if provided
	if model != "" {
		modelManager := ai.NewModelManager("http://localhost:11434")
		if err := modelManager.ValidateModelName(model); err != nil {
			return fmt.Errorf("invalid model name: %w", err)
		}

		// Check if model is installed
		installed, err := modelManager.IsModelInstalled(model)
		if err != nil {
			color.New(color.FgYellow).Printf("⚠️  Could not verify if model is installed: %v\n", err)
		} else if !installed {
			color.New(color.FgYellow).Printf("⚠️  Model %s is not installed\n", model)
			fmt.Printf("   Install it with: git-assist config --pull-model %s\n", model)
			fmt.Println()
		}
	}

	// Update configuration
	err = configManager.UpdateAI(model, endpoint, false) // Save to repository config
	if err != nil {
		return fmt.Errorf("failed to update configuration: %w", err)
	}

	// Show what was updated
	if model != "" {
		color.New(color.FgGreen).Printf("✅ Model set to: %s\n", model)
	}

	if endpoint != "" {
		color.New(color.FgGreen).Printf("✅ Endpoint set to: %s\n", endpoint)
	}

	// Test the new configuration
	if !quiet {
		fmt.Println()
		color.New(color.FgCyan).Println("💡 Test the new configuration:")
		fmt.Println("   git-assist config --test-ai")
	}

	return nil
}

// showCurrentConfig displays the current configuration
func showCurrentConfig() error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("⚙️  Current git-assist configuration")
		fmt.Println()
	}

	// Load configuration
	configManager, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Show AI configuration
	color.New(color.FgGreen, color.Bold).Println("🤖 AI Configuration:")
	fmt.Printf("   Provider: %s\n", cfg.AI.Provider)
	fmt.Printf("   Model: %s\n", cfg.AI.Model)
	fmt.Printf("   Endpoint: %s\n", cfg.AI.Endpoint)

	// Test current AI setup
	aiManager := ai.NewManager()
	err = aiManager.SetupFromConfig()

	if err != nil {
		color.New(color.FgRed).Printf("   Status: ❌ Not working (%v)\n", err)
	} else {
		modelInfo := aiManager.GetPrimaryModel()
		if modelInfo != nil {
			color.New(color.FgGreen).Printf("   Status: ✅ Connected\n")
			fmt.Printf("   Model Size: %s\n", modelInfo.Size)
		}
	}

	fmt.Println()

	// Show commit practices
	color.New(color.FgBlue, color.Bold).Println("📝 Commit Practices:")
	fmt.Printf("   Style: %s\n", cfg.Practices.Style)
	fmt.Printf("   Max length: %d characters\n", cfg.Practices.MaxLength)
	fmt.Printf("   Require type: %t\n", cfg.Practices.RequireType)
	if cfg.Practices.CustomFile != "" {
		fmt.Printf("   Custom rules file: %s\n", cfg.Practices.CustomFile)
	}

	fmt.Println()

	// Show preferences
	color.New(color.FgMagenta, color.Bold).Println("🎛️  Preferences:")
	fmt.Printf("   Output format: %s\n", cfg.Preferences.OutputFormat)
	fmt.Printf("   Color output: %t\n", cfg.Preferences.ColorOutput)
	fmt.Printf("   Auto-stage: %t\n", cfg.Preferences.AutoStage)

	if !quiet {
		fmt.Println()
		color.New(color.FgCyan).Println("💡 Use 'git-assist config --help' to see configuration options")
	}

	return nil
}

// runOllamaInstallation handles automatic Ollama installation
func runOllamaInstallation() error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("🔧 Installing Ollama...")
		fmt.Println()
	}

	installer := ai.NewOllamaInstaller()

	// Check if already installed
	if installer.IsOllamaInstalled() {
		color.New(color.FgGreen).Println("✅ Ollama is already installed")
		return nil
	}

	// Get installation methods
	methods, err := installer.GetInstallationInstructions()
	if err != nil {
		return fmt.Errorf("failed to get installation methods: %w", err)
	}

	if len(methods) == 0 {
		return fmt.Errorf("no installation methods available for your system")
	}

	// Use the first (recommended) method
	method := methods[0]

	if !quiet {
		color.New(color.FgYellow).Printf("Using method: %s\n", method.Description)
		if method.RequiresAdmin {
			color.New(color.FgRed).Println("⚠️  This installation requires administrator privileges")
		}
		fmt.Println()
	}

	// Attempt installation
	if err := installer.InstallOllama(method); err != nil {
		color.New(color.FgRed).Printf("❌ Installation failed: %v\n", err)
		fmt.Println()

		// Show alternative methods
		if len(methods) > 1 {
			color.New(color.FgYellow).Println("Alternative installation methods:")
			for i, alt := range methods[1:] {
				fmt.Printf("  %d. %s\n", i+1, alt.Description)
				for _, cmd := range alt.Commands {
					fmt.Printf("     %s\n", cmd)
				}
			}
		}

		return err
	}

	color.New(color.FgGreen).Println("✅ Ollama installed successfully!")

	// Suggest next steps
	if !quiet {
		fmt.Println()
		color.New(color.FgCyan).Println("Next steps:")
		fmt.Println("  1. Start the service: git-assist config --start-service")
		fmt.Println("  2. Pull a model: git-assist config --pull-model codellama:7b")
		fmt.Println("  3. Test: git-assist config --test-ai")
	}

	return nil
}

// startOllamaService starts the Ollama service in the background
func startOllamaService() error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("🚀 Starting Ollama service...")
		fmt.Println()
	}

	serviceManager := ai.NewServiceManager("http://localhost:11434")

	// Check if already running
	if serviceManager.IsRunning() {
		color.New(color.FgGreen).Println("✅ Ollama service is already running")
		return nil
	}

	// Start the service
	if err := serviceManager.StartService(); err != nil {
		color.New(color.FgRed).Printf("❌ Failed to start service: %v\n", err)
		return err
	}

	color.New(color.FgGreen).Println("✅ Ollama service started successfully!")

	// Show status
	status := serviceManager.GetStatus()
	if status.PID > 0 {
		fmt.Printf("   Process ID: %d\n", status.PID)
	}
	fmt.Printf("   Endpoint: %s\n", status.Endpoint)

	return nil
}

// stopOllamaService stops the Ollama service
func stopOllamaService() error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("🛑 Stopping Ollama service...")
		fmt.Println()
	}

	serviceManager := ai.NewServiceManager("http://localhost:11434")

	// Check if running
	if !serviceManager.IsRunning() {
		color.New(color.FgYellow).Println("⚠️  Ollama service is not running")
		return nil
	}

	// Stop the service
	if err := serviceManager.StopService(); err != nil {
		color.New(color.FgRed).Printf("❌ Failed to stop service: %v\n", err)
		return err
	}

	color.New(color.FgGreen).Println("✅ Ollama service stopped successfully!")
	return nil
}

// pullModelCommand downloads and installs a model
func pullModelCommand(modelName string) error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Printf("📥 Downloading model: %s\n", modelName)
		fmt.Println()
	}

	modelManager := ai.NewModelManager("http://localhost:11434")

	// Validate model name
	if err := modelManager.ValidateModelName(modelName); err != nil {
		return fmt.Errorf("invalid model name: %w", err)
	}

	// Check if already installed
	installed, err := modelManager.IsModelInstalled(modelName)
	if err != nil {
		return fmt.Errorf("failed to check if model is installed: %w", err)
	}

	if installed {
		color.New(color.FgGreen).Printf("✅ Model %s is already installed\n", modelName)
		return nil
	}

	// Show download estimate
	recommendations := modelManager.GetRecommendedModels()
	for _, category := range recommendations {
		for _, rec := range category {
			if rec.Name == modelName {
				fmt.Printf("   Size: %s\n", rec.Size)
				fmt.Printf("   Estimated download time: %s\n", modelManager.EstimateDownloadTime(rec.Size))
				fmt.Println()
				break
			}
		}
	}

	// Pull the model with progress
	err = modelManager.PullModel(modelName, func(progress ai.ModelPullProgress) {
		if !quiet {
			fmt.Printf("\r   Status: %s", progress.Status)
		}
	})

	if err != nil {
		color.New(color.FgRed).Printf("\n❌ Failed to download model: %v\n", err)
		return err
	}

	fmt.Println() // New line after progress
	color.New(color.FgGreen).Printf("✅ Model %s downloaded successfully!\n", modelName)

	// Suggest testing
	if !quiet {
		fmt.Println()
		color.New(color.FgCyan).Println("💡 Test the model:")
		fmt.Printf("   git-assist config --set-model %s\n", modelName)
		fmt.Println("   git-assist config --test-ai")
	}

	return nil
}

// listInstalledModels shows all installed models
func listInstalledModels() error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("📦 Installed Models")
		fmt.Println()
	}

	modelManager := ai.NewModelManager("http://localhost:11434")

	models, err := modelManager.ListInstalledModels()
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	if len(models) == 0 {
		color.New(color.FgYellow).Println("No models installed")
		fmt.Println()
		color.New(color.FgCyan).Println("💡 Install a model:")
		fmt.Println("   git-assist config --pull-model codellama:7b")
		return nil
	}

	for _, model := range models {
		color.New(color.FgGreen).Printf("✅ %s\n", model.Name)
		fmt.Printf("   Size: %.1f GB\n", float64(model.Size)/(1024*1024*1024))
		fmt.Printf("   Modified: %s\n", model.ModifiedAt.Format("2006-01-02 15:04"))
		fmt.Println()
	}

	if !quiet {
		color.New(color.FgCyan).Println("💡 Use a model:")
		fmt.Printf("   git-assist config --set-model %s\n", models[0].Name)
		fmt.Println("   git-assist config --test-ai")
	}

	return nil
}

// indexRepository indexes the current repository for context-aware commits
func indexRepository() error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("🔍 Indexing Repository")
		fmt.Println()
	}

	// Get current working directory (should be repository root)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Create indexer
	indexer := repository.NewIndexer(cwd)

	// Index the repository
	repoIndex, err := indexer.IndexRepository()
	if err != nil {
		return fmt.Errorf("failed to index repository: %w", err)
	}

	// Show summary
	if !quiet {
		color.New(color.FgGreen).Printf("✅ Repository indexed successfully!\n")
		fmt.Printf("   Files indexed: %d\n", repoIndex.Metadata.TotalFiles)
		fmt.Printf("   Total lines: %d\n", repoIndex.Metadata.TotalLines)
		fmt.Printf("   Main language: %s\n", repoIndex.Metadata.MainLanguage)
		fmt.Printf("   Languages found: %d\n", len(repoIndex.Languages))

		fmt.Println()
		color.New(color.FgCyan).Println("💡 Now you can use context-aware commits:")
		fmt.Println("   git-assist commit --with-context")
	}

	return nil
}

// showRepositoryIndex displays information about the current repository index
func showRepositoryIndex() error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("📊 Repository Index Information")
		fmt.Println()
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Create indexer and load index
	indexer := repository.NewIndexer(cwd)
	repoIndex, err := indexer.LoadIndex()
	if err != nil {
		color.New(color.FgRed).Println("❌ No repository index found")
		fmt.Println()
		color.New(color.FgYellow).Println("💡 Create an index first:")
		fmt.Println("   git-assist config --index-repo")
		return nil
	}

	// Display index information
	color.New(color.FgGreen, color.Bold).Println("📁 Repository Information:")
	fmt.Printf("   Name: %s\n", repoIndex.Metadata.Name)
	fmt.Printf("   Main Language: %s\n", repoIndex.Metadata.MainLanguage)
	fmt.Printf("   Total Files: %d\n", repoIndex.Metadata.TotalFiles)
	fmt.Printf("   Total Lines: %d\n", repoIndex.Metadata.TotalLines)
	fmt.Printf("   Indexed At: %s\n", repoIndex.IndexedAt.Format("2006-01-02 15:04:05"))

	fmt.Println()
	color.New(color.FgBlue, color.Bold).Println("🗣️  Languages:")
	for lang, count := range repoIndex.Languages {
		fmt.Printf("   %s: %d files\n", lang, count)
	}

	// Check if index is stale
	if stale, _ := indexer.IsIndexStale(); stale {
		fmt.Println()
		color.New(color.FgYellow).Println("⚠️  Index may be outdated")
		color.New(color.FgCyan).Println("💡 Refresh the index:")
		fmt.Println("   git-assist config --index-repo")
	}

	return nil
}

// setupShellCompletion automatically sets up shell autocompletion
func setupShellCompletion() error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("🔧 Setting up shell autocompletion...")
		fmt.Println()
	}

	// Detect current shell
	shell := detectShell()
	if shell == "" {
		color.New(color.FgRed).Println("❌ Could not detect shell")
		fmt.Println()
		color.New(color.FgYellow).Println("💡 Manual setup:")
		showManualCompletionInstructions()
		return nil
	}

	if !quiet {
		color.New(color.FgBlue).Printf("🐚 Detected shell: %s\n", shell)
	}

	// Setup completion based on shell
	switch shell {
	case "zsh":
		return setupZshCompletion()
	case "bash":
		return setupBashCompletion()
	case "fish":
		return setupFishCompletion()
	default:
		color.New(color.FgYellow).Printf("⚠️  Shell '%s' not fully supported\n", shell)
		fmt.Println()
		showManualCompletionInstructions()
		return nil
	}
}

// detectShell detects the current shell
func detectShell() string {
	// Check SHELL environment variable
	shell := os.Getenv("SHELL")
	if shell != "" {
		if strings.Contains(shell, "zsh") {
			return "zsh"
		} else if strings.Contains(shell, "bash") {
			return "bash"
		} else if strings.Contains(shell, "fish") {
			return "fish"
		}
	}

	// Check if we're running in a specific shell
	if os.Getenv("ZSH_VERSION") != "" {
		return "zsh"
	} else if os.Getenv("BASH_VERSION") != "" {
		return "bash"
	}

	return ""
}

// setupZshCompletion sets up autocompletion for zsh
func setupZshCompletion() error {
	var completionDir string

	// Try different zsh completion directories
	possibleDirs := []string{
		"/opt/homebrew/share/zsh/site-functions", // Apple Silicon Homebrew
		"/usr/local/share/zsh/site-functions",    // Intel Homebrew
		"/usr/share/zsh/site-functions",          // System zsh
		os.ExpandEnv("$HOME/.zsh/completions"),   // User directory
	}

	// Find the first writable directory
	for _, dir := range possibleDirs {
		if isWritableDir(dir) {
			completionDir = dir
			break
		}
	}

	// If no system directory is writable, create user directory
	if completionDir == "" {
		userDir := os.ExpandEnv("$HOME/.zsh/completions")
		if err := os.MkdirAll(userDir, 0755); err == nil {
			completionDir = userDir
		}
	}

	if completionDir == "" {
		return fmt.Errorf("could not find writable zsh completion directory")
	}

	// Generate completion script
	completionFile := filepath.Join(completionDir, "_git-assist")

	if !quiet {
		color.New(color.FgBlue).Printf("📁 Installing to: %s\n", completionFile)
	}

	// Get completion script
	cmd := exec.Command(os.Args[0], "completion", "zsh")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to generate completion script: %w", err)
	}

	// Write completion script
	if err := os.WriteFile(completionFile, output, 0644); err != nil {
		return fmt.Errorf("failed to write completion file: %w", err)
	}

	color.New(color.FgGreen).Println("✅ Zsh completion installed successfully!")

	// Check if user needs to add completion directory to fpath
	if strings.Contains(completionDir, ".zsh/completions") {
		fmt.Println()
		color.New(color.FgYellow).Println("⚠️  Add this to your ~/.zshrc:")
		fmt.Printf("   fpath=(~/.zsh/completions $fpath)\n")
		fmt.Printf("   autoload -U compinit && compinit\n")
	}

	fmt.Println()
	color.New(color.FgCyan).Println("💡 Restart your shell or run:")
	fmt.Println("   exec zsh")

	return nil
}

// setupBashCompletion sets up autocompletion for bash
func setupBashCompletion() error {
	// Try system-wide first, then user-specific
	possibleDirs := []string{
		"/etc/bash_completion.d",
		"/usr/local/etc/bash_completion.d",
		os.ExpandEnv("$HOME/.bash_completion.d"),
	}

	var completionDir string
	for _, dir := range possibleDirs {
		if isWritableDir(dir) {
			completionDir = dir
			break
		}
	}

	// Create user directory if needed
	if completionDir == "" {
		userDir := os.ExpandEnv("$HOME/.bash_completion.d")
		if err := os.MkdirAll(userDir, 0755); err == nil {
			completionDir = userDir
		}
	}

	if completionDir == "" {
		return fmt.Errorf("could not find writable bash completion directory")
	}

	completionFile := filepath.Join(completionDir, "git-assist")

	if !quiet {
		color.New(color.FgBlue).Printf("📁 Installing to: %s\n", completionFile)
	}

	// Get completion script
	cmd := exec.Command(os.Args[0], "completion", "bash")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to generate completion script: %w", err)
	}

	// Write completion script
	if err := os.WriteFile(completionFile, output, 0644); err != nil {
		return fmt.Errorf("failed to write completion file: %w", err)
	}

	color.New(color.FgGreen).Println("✅ Bash completion installed successfully!")

	// Check if user needs to source the file
	if strings.Contains(completionDir, ".bash_completion.d") {
		fmt.Println()
		color.New(color.FgYellow).Println("⚠️  Add this to your ~/.bashrc:")
		fmt.Printf("   source ~/.bash_completion.d/git-assist\n")
	}

	fmt.Println()
	color.New(color.FgCyan).Println("💡 Restart your shell or run:")
	fmt.Println("   source ~/.bashrc")

	return nil
}

// setupFishCompletion sets up autocompletion for fish
func setupFishCompletion() error {
	completionDir := os.ExpandEnv("$HOME/.config/fish/completions")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(completionDir, 0755); err != nil {
		return fmt.Errorf("failed to create fish completion directory: %w", err)
	}

	completionFile := filepath.Join(completionDir, "git-assist.fish")

	if !quiet {
		color.New(color.FgBlue).Printf("📁 Installing to: %s\n", completionFile)
	}

	// Get completion script
	cmd := exec.Command(os.Args[0], "completion", "fish")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to generate completion script: %w", err)
	}

	// Write completion script
	if err := os.WriteFile(completionFile, output, 0644); err != nil {
		return fmt.Errorf("failed to write completion file: %w", err)
	}

	color.New(color.FgGreen).Println("✅ Fish completion installed successfully!")
	fmt.Println()
	color.New(color.FgCyan).Println("💡 Completion will be available in new fish sessions")

	return nil
}

// isWritableDir checks if a directory exists and is writable
func isWritableDir(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}

	if !info.IsDir() {
		return false
	}

	// Try to create a temporary file to test write permissions
	testFile := filepath.Join(dir, ".git-assist-test")
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	file.Close()
	os.Remove(testFile)

	return true
}

// showManualCompletionInstructions shows manual setup instructions
func showManualCompletionInstructions() {
	color.New(color.FgCyan, color.Bold).Println("Manual Setup Instructions:")
	fmt.Println()

	color.New(color.FgGreen).Println("Zsh (macOS, many Linux):")
	fmt.Println("  git-assist completion zsh > $(brew --prefix)/share/zsh/site-functions/_git-assist")
	fmt.Println("  exec zsh")
	fmt.Println()

	color.New(color.FgBlue).Println("Bash (Linux):")
	fmt.Println("  sudo git-assist completion bash > /etc/bash_completion.d/git-assist")
	fmt.Println("  source ~/.bashrc")
	fmt.Println()

	color.New(color.FgMagenta).Println("Fish:")
	fmt.Println("  git-assist completion fish > ~/.config/fish/completions/git-assist.fish")
	fmt.Println()

	color.New(color.FgYellow).Println("PowerShell (Windows):")
	fmt.Println("  git-assist completion powershell > git-assist.ps1")
	fmt.Println("  # Add to PowerShell profile")
}
