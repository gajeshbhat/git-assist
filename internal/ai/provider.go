// Package ai provides AI model integration for git-assist.
//
// This package defines interfaces and implementations for different AI providers
// like Ollama, OpenAI, LocalAI, etc. It abstracts the AI interaction so we can
// easily switch between different models and providers.
package ai

import (
	"context"
	"fmt"
)

// Provider defines the interface that all AI providers must implement.
// This allows us to swap between different AI services (Ollama, OpenAI, etc.)
// without changing the rest of our code.
type Provider interface {
	// GenerateText sends a prompt to the AI model and returns the response
	GenerateText(ctx context.Context, prompt string) (string, error)

	// GenerateCommitMessage is a specialized method for commit message generation
	// It takes a git diff and returns a commit message
	GenerateCommitMessage(ctx context.Context, diff string) (string, error)

	// Configure sets up the provider with the given configuration
	Configure(config Config) error

	// IsAvailable checks if the provider is ready to use
	IsAvailable() bool

	// GetModelInfo returns information about the current model
	GetModelInfo() ModelInfo
}

// Config holds configuration for an AI provider
type Config struct {
	// Type of provider: "ollama", "openai", "localai", "custom"
	Type string `json:"type"`

	// Model name (e.g., "codellama:7b", "gpt-4", "mistral:7b")
	Model string `json:"model"`

	// Endpoint URL for the AI service (e.g., "http://localhost:11434")
	Endpoint string `json:"endpoint"`

	// API key for cloud providers (optional for local providers)
	APIKey string `json:"api_key,omitempty"`

	// Additional parameters for the model
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// ModelInfo contains information about the AI model
type ModelInfo struct {
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	Size        string `json:"size"` // e.g., "7B", "13B"
	Description string `json:"description"`
	Available   bool   `json:"available"`
}

// CommitMessageRequest contains the context for generating a commit message
type CommitMessageRequest struct {
	Diff            string            `json:"diff"`             // Git diff content
	Files           []string          `json:"files"`            // List of changed files
	Branch          string            `json:"branch"`           // Current branch name
	PreviousCommits []string          `json:"previous_commits"` // Recent commit messages for context
	Rules           map[string]string `json:"rules"`            // Commit message rules/preferences
}

// CommitMessageResponse contains the generated commit message and metadata
type CommitMessageResponse struct {
	Message     string            `json:"message"`     // The generated commit message
	Confidence  float64           `json:"confidence"`  // AI confidence score (0-1)
	Reasoning   string            `json:"reasoning"`   // Why this message was chosen
	Suggestions []string          `json:"suggestions"` // Alternative suggestions
	Metadata    map[string]string `json:"metadata"`    // Additional metadata
}

// ProviderManager manages multiple AI providers and handles fallbacks
type ProviderManager struct {
	providers []Provider
	primary   Provider
	fallback  Provider
}

// NewProviderManager creates a new provider manager
func NewProviderManager() *ProviderManager {
	return &ProviderManager{
		providers: make([]Provider, 0),
	}
}

// AddProvider adds a provider to the manager
func (pm *ProviderManager) AddProvider(provider Provider) {
	pm.providers = append(pm.providers, provider)

	// Set as primary if it's the first available provider
	if pm.primary == nil && provider.IsAvailable() {
		pm.primary = provider
	}
}

// SetPrimary sets the primary provider
func (pm *ProviderManager) SetPrimary(provider Provider) error {
	if !provider.IsAvailable() {
		return fmt.Errorf("provider is not available")
	}
	pm.primary = provider
	return nil
}

// SetFallback sets the fallback provider
func (pm *ProviderManager) SetFallback(provider Provider) {
	pm.fallback = provider
}

// GenerateCommitMessage generates a commit message using the primary provider,
// falling back to the fallback provider if the primary fails
func (pm *ProviderManager) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	if pm.primary != nil && pm.primary.IsAvailable() {
		result, err := pm.primary.GenerateCommitMessage(ctx, diff)
		if err == nil {
			return result, nil
		}
		// Log the error but continue to fallback
		fmt.Printf("Primary provider failed: %v\n", err)
	}

	if pm.fallback != nil && pm.fallback.IsAvailable() {
		return pm.fallback.GenerateCommitMessage(ctx, diff)
	}

	return "", fmt.Errorf("no available AI providers")
}

// GetAvailableProviders returns a list of available providers
func (pm *ProviderManager) GetAvailableProviders() []Provider {
	available := make([]Provider, 0)
	for _, provider := range pm.providers {
		if provider.IsAvailable() {
			available = append(available, provider)
		}
	}
	return available
}

// GetPrimaryProvider returns the current primary provider
func (pm *ProviderManager) GetPrimaryProvider() Provider {
	return pm.primary
}
