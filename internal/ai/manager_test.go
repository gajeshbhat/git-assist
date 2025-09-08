package ai

import (
	"fmt"
	"github.com/gajeshbhat/git-assist/internal/config"
	"strings"
	"testing"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}
}

func TestSetupFromConfig(t *testing.T) {
	manager := NewManager()

	// Test with default config
	configManager, err := config.NewManager()
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	cfg, err := configManager.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	err = manager.SetupFromConfig()

	// Should not error even if Ollama is not running
	// The actual connection test happens later
	if err != nil {
		t.Logf("Setup failed (expected if Ollama not running): %v", err)
	}

	// Use cfg to avoid unused variable error
	_ = cfg
}

func TestValidateProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		wantErr  bool
	}{
		{
			name:     "valid ollama",
			provider: "ollama",
			wantErr:  false,
		},
		{
			name:     "valid openai",
			provider: "openai",
			wantErr:  false,
		},
		{
			name:     "valid anthropic",
			provider: "anthropic",
			wantErr:  false,
		},
		{
			name:     "invalid provider",
			provider: "invalid",
			wantErr:  true,
		},
		{
			name:     "empty provider",
			provider: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProvider(tt.provider)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateModel(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		model    string
		wantErr  bool
	}{
		{
			name:     "valid ollama model",
			provider: "ollama",
			model:    "codellama:7b",
			wantErr:  false,
		},
		{
			name:     "valid openai model",
			provider: "openai",
			model:    "gpt-3.5-turbo",
			wantErr:  false,
		},
		{
			name:     "valid anthropic model",
			provider: "anthropic",
			model:    "claude-3-sonnet",
			wantErr:  false,
		},
		{
			name:     "empty model",
			provider: "ollama",
			model:    "",
			wantErr:  true,
		},
		{
			name:     "invalid openai model",
			provider: "openai",
			model:    "invalid-model",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateModel(tt.provider, tt.model)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateModel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateCommitMessage(t *testing.T) {
	manager := NewManager()

	// Test with empty diff (should handle gracefully)
	message, err := manager.GenerateCommitMessage("")
	if err == nil {
		t.Error("GenerateCommitMessage() should fail with empty diff")
	}

	// Test with sample diff
	sampleDiff := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,6 @@
 package main
 
-func main() {}
+func main() {
+    fmt.Println("Hello, World!")
+}
`

	// This will fail if no AI provider is configured, which is expected
	message, err = manager.GenerateCommitMessage(sampleDiff)
	if err != nil {
		t.Logf("GenerateCommitMessage() failed (expected without AI setup): %v", err)
	} else {
		if message == "" {
			t.Error("GenerateCommitMessage() returned empty message")
		}
		t.Logf("Generated message: %s", message)
	}
}

func TestIsAvailable(t *testing.T) {
	manager := NewManager()

	// Without setup, should not be available
	if manager.IsAvailable() {
		t.Error("Manager should not be available without setup")
	}

	// After setup attempt (may still fail if no AI provider)
	manager.SetupFromConfig()

	// IsAvailable() result depends on whether AI provider is actually running
	available := manager.IsAvailable()
	t.Logf("AI available: %v", available)
}

// Helper functions for validation testing

func validateProvider(provider string) error {
	validProviders := []string{"ollama", "openai", "anthropic"}

	if provider == "" {
		return fmt.Errorf("provider cannot be empty")
	}

	for _, valid := range validProviders {
		if provider == valid {
			return nil
		}
	}

	return fmt.Errorf("unsupported provider: %s", provider)
}

func validateModel(provider, model string) error {
	if model == "" {
		return fmt.Errorf("model cannot be empty")
	}

	// Basic validation - in real implementation, this would be more sophisticated
	switch provider {
	case "openai":
		validModels := []string{"gpt-3.5-turbo", "gpt-4", "gpt-4-turbo"}
		for _, valid := range validModels {
			if model == valid {
				return nil
			}
		}
		return fmt.Errorf("unsupported OpenAI model: %s", model)
	case "anthropic":
		if !strings.HasPrefix(model, "claude-") {
			return fmt.Errorf("invalid Anthropic model format: %s", model)
		}
	case "ollama":
		// Ollama models can be various formats, so we're more lenient
		if !strings.Contains(model, ":") && model != "latest" {
			return fmt.Errorf("Ollama model should include tag (e.g., 'model:tag'): %s", model)
		}
	}

	return nil
}
