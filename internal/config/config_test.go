package config

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	manager, err := NewManager()

	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	// Test loading default config
	config, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Test default values
	if config.AI.Provider != "ollama" {
		t.Errorf("Expected default provider 'ollama', got '%s'", config.AI.Provider)
	}

	if config.AI.Model != "codellama:7b" {
		t.Errorf("Expected default model 'codellama:7b', got '%s'", config.AI.Model)
	}

	if config.AI.Endpoint != "http://localhost:11434" {
		t.Errorf("Expected default endpoint 'http://localhost:11434', got '%s'", config.AI.Endpoint)
	}
}

func TestGetAIConfig(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	aiConfig, err := manager.GetAIConfig()
	if err != nil {
		t.Fatalf("GetAIConfig() failed: %v", err)
	}

	if aiConfig == nil {
		t.Fatal("GetAIConfig() returned nil")
	}

	// Test default values
	if aiConfig.Provider != "ollama" {
		t.Errorf("Expected default provider 'ollama', got '%s'", aiConfig.Provider)
	}
}
