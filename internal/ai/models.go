// Package ai/models handles AI model management (downloading, listing, etc.)
package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// ModelManager handles AI model operations
type ModelManager struct {
	endpoint string
	client   *http.Client
}

// NewModelManager creates a new model manager
func NewModelManager(endpoint string) *ModelManager {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}

	return &ModelManager{
		endpoint: endpoint,
		client: &http.Client{
			Timeout: 300 * time.Second, // 5 minutes for model operations
		},
	}
}

// InstalledModel represents a model that's installed locally
type InstalledModel struct {
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
	ModifiedAt time.Time `json:"modified_at"`
	Family     string    `json:"family,omitempty"`
	Format     string    `json:"format,omitempty"`
}

// ModelPullProgress represents progress during model download
type ModelPullProgress struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

// ProgressCallback is called during model download to report progress
type ProgressCallback func(progress ModelPullProgress)

// ListInstalledModels returns a list of models installed locally
func (mm *ModelManager) ListInstalledModels() ([]InstalledModel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", mm.endpoint+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := mm.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var response struct {
		Models []InstalledModel `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Models, nil
}

// IsModelInstalled checks if a specific model is installed
func (mm *ModelManager) IsModelInstalled(modelName string) (bool, error) {
	models, err := mm.ListInstalledModels()
	if err != nil {
		return false, err
	}

	for _, model := range models {
		if model.Name == modelName {
			return true, nil
		}
	}

	return false, nil
}

// PullModel downloads a model from the Ollama registry
func (mm *ModelManager) PullModel(modelName string, callback ProgressCallback) error {
	// Use ollama CLI for pulling models as it handles progress better
	return mm.pullModelViaCLI(modelName, callback)
}

// pullModelViaCLI uses the ollama CLI to pull models
func (mm *ModelManager) pullModelViaCLI(modelName string, callback ProgressCallback) error {
	cmd := exec.Command("ollama", "pull", modelName)

	// Get stdout pipe to read progress
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ollama pull: %w", err)
	}

	// Read progress line by line
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		// Parse progress information
		if callback != nil {
			progress := mm.parseProgressLine(line)
			callback(progress)
		}
	}

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ollama pull failed: %w", err)
	}

	return nil
}

// parseProgressLine parses a progress line from ollama pull output
func (mm *ModelManager) parseProgressLine(line string) ModelPullProgress {
	// Ollama pull output format varies, this is a simplified parser
	progress := ModelPullProgress{
		Status: strings.TrimSpace(line),
	}

	// Look for common progress indicators
	if strings.Contains(line, "pulling") {
		progress.Status = "downloading"
	} else if strings.Contains(line, "verifying") {
		progress.Status = "verifying"
	} else if strings.Contains(line, "success") {
		progress.Status = "complete"
	}

	return progress
}

// DeleteModel removes a model from local storage
func (mm *ModelManager) DeleteModel(modelName string) error {
	cmd := exec.Command("ollama", "rm", modelName)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete model %s: %s (%w)", modelName, string(output), err)
	}

	return nil
}

// GetModelInfo gets detailed information about a specific model
func (mm *ModelManager) GetModelInfo(modelName string) (*InstalledModel, error) {
	models, err := mm.ListInstalledModels()
	if err != nil {
		return nil, err
	}

	for _, model := range models {
		if model.Name == modelName {
			return &model, nil
		}
	}

	return nil, fmt.Errorf("model %s not found", modelName)
}

// SuggestBestModel suggests the best model based on system capabilities
func (mm *ModelManager) SuggestBestModel() string {
	// This is a simplified heuristic
	// In a real implementation, we'd check available RAM, CPU, etc.

	// Check what models are already installed
	models, err := mm.ListInstalledModels()
	if err == nil && len(models) > 0 {
		// Prefer code-focused models if available
		for _, model := range models {
			if strings.Contains(model.Name, "codellama") {
				return model.Name
			}
		}
		// Return the first available model
		return models[0].Name
	}

	// Default suggestion
	return "codellama:7b"
}

// ValidateModelName checks if a model name is valid
func (mm *ModelManager) ValidateModelName(modelName string) error {
	if modelName == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	// Basic validation - model names should follow pattern: name:tag
	if !strings.Contains(modelName, ":") {
		return fmt.Errorf("model name should include a tag (e.g., codellama:7b)")
	}

	parts := strings.Split(modelName, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid model name format, expected name:tag")
	}

	if parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("model name and tag cannot be empty")
	}

	return nil
}

// GetRecommendedModels returns a categorized list of recommended models
func (mm *ModelManager) GetRecommendedModels() map[string][]ModelRecommendation {
	return map[string][]ModelRecommendation{
		"code": {
			{Name: "codellama:7b", Size: "3.8GB", Description: "Fast code generation, good balance"},
			{Name: "codellama:13b", Size: "7.3GB", Description: "Better quality, slower"},
			{Name: "deepseek-coder:6.7b", Size: "3.8GB", Description: "Specialized for coding tasks"},
			{Name: "codegemma:7b", Size: "5.0GB", Description: "Google's code model"},
		},
		"general": {
			{Name: "mistral:7b", Size: "4.1GB", Description: "Fast and capable general model"},
			{Name: "llama3:8b", Size: "4.7GB", Description: "Meta's latest model"},
			{Name: "phi3:3.8b", Size: "2.3GB", Description: "Smaller, faster model"},
		},
		"large": {
			{Name: "codellama:34b", Size: "19GB", Description: "High quality, requires 32GB+ RAM"},
			{Name: "llama3:70b", Size: "40GB", Description: "Very high quality, requires 64GB+ RAM"},
		},
	}
}

// ModelRecommendation represents a recommended model
type ModelRecommendation struct {
	Name        string `json:"name"`
	Size        string `json:"size"`
	Description string `json:"description"`
}

// EstimateDownloadTime estimates how long a model will take to download
func (mm *ModelManager) EstimateDownloadTime(modelSize string) string {
	// This is a rough estimate based on typical internet speeds
	// In a real implementation, we might test download speed first

	switch {
	case strings.Contains(modelSize, "2.3GB"):
		return "5-15 minutes"
	case strings.Contains(modelSize, "3.8GB"):
		return "8-25 minutes"
	case strings.Contains(modelSize, "4.1GB"):
		return "10-30 minutes"
	case strings.Contains(modelSize, "7.3GB"):
		return "15-45 minutes"
	case strings.Contains(modelSize, "19GB"):
		return "45-120 minutes"
	case strings.Contains(modelSize, "40GB"):
		return "2-6 hours"
	default:
		return "varies"
	}
}
