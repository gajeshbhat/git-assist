package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/gajeshbhat/git-assist/internal/config"
)

// Manager handles AI operations for git-assist
// It manages different AI providers and provides a simple interface
// for the rest of the application to use AI features
type Manager struct {
	providerManager *ProviderManager
	defaultTimeout  time.Duration
	configManager   *config.Manager
}

// NewManager creates a new AI manager
func NewManager() *Manager {
	configManager, _ := config.NewManager() // Ignore error for now

	return &Manager{
		providerManager: NewProviderManager(),
		defaultTimeout:  30 * time.Second, // Default 30 second timeout
		configManager:   configManager,
	}
}

// SetupOllama configures and adds an Ollama provider
func (m *Manager) SetupOllama(model string, endpoint string) error {
	config := Config{
		Type:     "ollama",
		Model:    model,
		Endpoint: endpoint,
	}

	provider := NewOllamaProvider(config)

	// Test if the provider is available
	if !provider.IsAvailable() {
		return fmt.Errorf("ollama is not available at %s - make sure Ollama is running", endpoint)
	}

	// Configure the provider
	if err := provider.Configure(config); err != nil {
		return fmt.Errorf("failed to configure Ollama provider: %w", err)
	}

	// Add to provider manager
	m.providerManager.AddProvider(provider)

	// Set as primary if it's the first provider
	if m.providerManager.GetPrimaryProvider() == nil {
		return m.providerManager.SetPrimary(provider)
	}

	return nil
}

// SetupFromConfig configures AI providers from saved configuration
func (m *Manager) SetupFromConfig() error {
	if m.configManager == nil {
		return fmt.Errorf("config manager not initialized")
	}

	aiConfig, err := m.configManager.GetAIConfig()
	if err != nil {
		return fmt.Errorf("failed to load AI config: %w", err)
	}

	// Setup provider based on configuration
	switch aiConfig.Provider {
	case "ollama":
		return m.SetupOllama(aiConfig.Model, aiConfig.Endpoint)
	default:
		return fmt.Errorf("unsupported AI provider: %s", aiConfig.Provider)
	}
}

// GenerateCommitMessage generates a commit message using the configured AI provider
func (m *Manager) GenerateCommitMessage(diff string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.defaultTimeout)
	defer cancel()

	return m.providerManager.GenerateCommitMessage(ctx, diff)
}

// GenerateCommitMessageWithContext generates a commit message with a custom context
func (m *Manager) GenerateCommitMessageWithContext(ctx context.Context, diff string) (string, error) {
	return m.providerManager.GenerateCommitMessage(ctx, diff)
}

// IsAvailable checks if any AI provider is available
func (m *Manager) IsAvailable() bool {
	providers := m.providerManager.GetAvailableProviders()
	return len(providers) > 0
}

// GetAvailableModels returns information about available AI models
func (m *Manager) GetAvailableModels() []ModelInfo {
	providers := m.providerManager.GetAvailableProviders()
	models := make([]ModelInfo, 0, len(providers))

	for _, provider := range providers {
		models = append(models, provider.GetModelInfo())
	}

	return models
}

// GetPrimaryModel returns information about the primary AI model
func (m *Manager) GetPrimaryModel() *ModelInfo {
	primary := m.providerManager.GetPrimaryProvider()
	if primary == nil {
		return nil
	}

	info := primary.GetModelInfo()
	return &info
}

// SetTimeout sets the default timeout for AI operations
func (m *Manager) SetTimeout(timeout time.Duration) {
	m.defaultTimeout = timeout
}

// DefaultConfig returns a default configuration for common AI setups
func DefaultConfig() Config {
	return Config{
		Type:     "ollama",
		Model:    "codellama:7b",
		Endpoint: "http://localhost:11434",
		Parameters: map[string]interface{}{
			"temperature": 0.1, // Low temperature for more consistent output
			"top_p":       0.9,
			"top_k":       40,
		},
	}
}

// RecommendedModels returns a list of recommended models for different use cases
func RecommendedModels() map[string][]string {
	return map[string][]string{
		"code": {
			"codellama:7b",        // Good balance of speed and quality
			"codellama:13b",       // Better quality, slower
			"deepseek-coder:6.7b", // Specialized for coding
		},
		"general": {
			"mistral:7b", // Fast and capable
			"llama2:7b",  // Well-tested
			"phi3:3.8b",  // Smaller, faster
		},
		"large": {
			"codellama:34b", // High quality, requires more resources
			"llama2:70b",    // Very high quality, slow
		},
	}
}

// ValidateModel checks if a model name is in a valid format
func ValidateModel(model string) error {
	if model == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	// Basic validation - model names should contain at least one character
	// and optionally a tag after a colon (e.g., "codellama:7b")
	if len(model) < 1 {
		return fmt.Errorf("model name is too short")
	}

	return nil
}

// SuggestModel suggests a model based on system capabilities
// This is a simple heuristic - in a real implementation, we might
// check available RAM, CPU, etc.
func SuggestModel() string {
	// For now, suggest a balanced model
	// In the future, we could check system resources and suggest accordingly
	return "codellama:7b"
}
