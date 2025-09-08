// Package config handles git-assist configuration management
//
// This package manages configuration files, user preferences, and settings
// for git-assist. It handles both global and repository-specific configuration.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// GitAssistConfig represents the complete configuration for git-assist
type GitAssistConfig struct {
	Version     string            `json:"version"`
	AI          AIConfig          `json:"ai"`
	Practices   PracticesConfig   `json:"practices"`
	Preferences PreferencesConfig `json:"preferences"`
}

// AIConfig contains AI-related configuration
type AIConfig struct {
	Provider   string                 `json:"provider"` // "ollama", "openai", etc.
	Model      string                 `json:"model"`    // "codellama:7b", etc.
	Endpoint   string                 `json:"endpoint"` // "http://localhost:11434"
	APIKey     string                 `json:"api_key,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// PracticesConfig contains commit message practices configuration
type PracticesConfig struct {
	Style       string   `json:"style"`                 // "conventional", "angular", "gitmoji"
	MaxLength   int      `json:"max_length"`            // Maximum commit message length
	RequireType bool     `json:"require_type"`          // Require commit type prefix
	Rules       []string `json:"rules"`                 // Custom rules
	CustomFile  string   `json:"custom_file,omitempty"` // Path to custom rules file
}

// PreferencesConfig contains user preferences
type PreferencesConfig struct {
	OutputFormat string `json:"output_format"` // "text", "json"
	ColorOutput  bool   `json:"color_output"`  // Enable colored output
	AutoStage    bool   `json:"auto_stage"`    // Auto-stage changes before commit
	Verbose      bool   `json:"verbose"`       // Verbose output by default
}

// Manager handles configuration operations
type Manager struct {
	globalPath string
	repoPath   string
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	// Get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Global config path: ~/.git-assist/config.json
	globalPath := filepath.Join(homeDir, ".git-assist", "config.json")

	// Repository config path: .git/git-assist/config.json
	repoPath := filepath.Join(".git", "git-assist", "config.json")

	return &Manager{
		globalPath: globalPath,
		repoPath:   repoPath,
	}, nil
}

// Load loads configuration with hierarchy: repo > global > defaults
func (m *Manager) Load() (*GitAssistConfig, error) {
	// Start with default configuration
	config := m.getDefaultConfig()

	// Try to load global configuration
	if globalConfig, err := m.loadFromFile(m.globalPath); err == nil {
		m.mergeConfigs(config, globalConfig)
	}

	// Try to load repository configuration (overrides global)
	if repoConfig, err := m.loadFromFile(m.repoPath); err == nil {
		m.mergeConfigs(config, repoConfig)
	}

	return config, nil
}

// Save saves configuration to the appropriate location
func (m *Manager) Save(config *GitAssistConfig, global bool) error {
	var path string
	if global {
		path = m.globalPath
		// Ensure global config directory exists
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create global config directory: %w", err)
		}
	} else {
		path = m.repoPath
		// Ensure repository config directory exists
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create repository config directory: %w", err)
		}
	}

	return m.saveToFile(config, path)
}

// UpdateAI updates AI configuration
func (m *Manager) UpdateAI(model, endpoint string, global bool) error {
	config, err := m.Load()
	if err != nil {
		return fmt.Errorf("failed to load current config: %w", err)
	}

	if model != "" {
		config.AI.Model = model
	}

	if endpoint != "" {
		config.AI.Endpoint = endpoint
	}

	// Set provider based on endpoint
	if config.AI.Endpoint != "" {
		config.AI.Provider = "ollama" // Default to ollama for now
	}

	return m.Save(config, global)
}

// GetAIConfig returns the current AI configuration
func (m *Manager) GetAIConfig() (*AIConfig, error) {
	config, err := m.Load()
	if err != nil {
		return nil, err
	}

	return &config.AI, nil
}

// loadFromFile loads configuration from a specific file
func (m *Manager) loadFromFile(path string) (*GitAssistConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config GitAssistConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	return &config, nil
}

// saveToFile saves configuration to a specific file
func (m *Manager) saveToFile(config *GitAssistConfig, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", path, err)
	}

	return nil
}

// mergeConfigs merges source config into target config
func (m *Manager) mergeConfigs(target, source *GitAssistConfig) {
	// Merge AI config
	if source.AI.Provider != "" {
		target.AI.Provider = source.AI.Provider
	}
	if source.AI.Model != "" {
		target.AI.Model = source.AI.Model
	}
	if source.AI.Endpoint != "" {
		target.AI.Endpoint = source.AI.Endpoint
	}
	if source.AI.APIKey != "" {
		target.AI.APIKey = source.AI.APIKey
	}
	if source.AI.Parameters != nil {
		if target.AI.Parameters == nil {
			target.AI.Parameters = make(map[string]interface{})
		}
		for k, v := range source.AI.Parameters {
			target.AI.Parameters[k] = v
		}
	}

	// Merge practices config
	if source.Practices.Style != "" {
		target.Practices.Style = source.Practices.Style
	}
	if source.Practices.MaxLength > 0 {
		target.Practices.MaxLength = source.Practices.MaxLength
	}
	if source.Practices.Rules != nil {
		target.Practices.Rules = source.Practices.Rules
	}
	if source.Practices.CustomFile != "" {
		target.Practices.CustomFile = source.Practices.CustomFile
	}

	// Merge preferences config
	if source.Preferences.OutputFormat != "" {
		target.Preferences.OutputFormat = source.Preferences.OutputFormat
	}
	// Note: for boolean fields, we need to be careful about zero values
	// In a real implementation, we might use pointers or a different approach
}

// getDefaultConfig returns the default configuration
func (m *Manager) getDefaultConfig() *GitAssistConfig {
	return &GitAssistConfig{
		Version: "1.0",
		AI: AIConfig{
			Provider: "ollama",
			Model:    "codellama:7b",
			Endpoint: "http://localhost:11434",
			Parameters: map[string]interface{}{
				"temperature": 0.1,
				"top_p":       0.9,
				"top_k":       40,
			},
		},
		Practices: PracticesConfig{
			Style:       "conventional",
			MaxLength:   50,
			RequireType: true,
			Rules:       []string{"require-type", "max-length-50"},
		},
		Preferences: PreferencesConfig{
			OutputFormat: "text",
			ColorOutput:  true,
			AutoStage:    false,
			Verbose:      false,
		},
	}
}
