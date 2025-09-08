package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OllamaProvider implements the Provider interface for Ollama
// Ollama is a local AI platform that makes it easy to run language models locally
type OllamaProvider struct {
	config   Config
	client   *http.Client
	endpoint string
	model    string
}

// OllamaRequest represents a request to the Ollama API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	// Options for controlling the model behavior
	Options map[string]interface{} `json:"options,omitempty"`
}

// OllamaResponse represents a response from the Ollama API
type OllamaResponse struct {
	Model     string `json:"model"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	Context   []int  `json:"context,omitempty"`
	CreatedAt string `json:"created_at"`
}

// OllamaModelInfo represents model information from Ollama
type OllamaModelInfo struct {
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
	ModifiedAt time.Time `json:"modified_at"`
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(config Config) *OllamaProvider {
	// Default endpoint if not specified
	endpoint := config.Endpoint
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}

	return &OllamaProvider{
		config:   config,
		endpoint: endpoint,
		model:    config.Model,
		client: &http.Client{
			Timeout: 60 * time.Second, // 60 second timeout for AI requests
		},
	}
}

// Configure sets up the Ollama provider with the given configuration
func (o *OllamaProvider) Configure(config Config) error {
	o.config = config
	o.model = config.Model

	if config.Endpoint != "" {
		o.endpoint = config.Endpoint
	}

	// Validate that we can connect to Ollama
	if !o.IsAvailable() {
		return fmt.Errorf("cannot connect to Ollama at %s", o.endpoint)
	}

	return nil
}

// IsAvailable checks if Ollama is running and accessible
func (o *OllamaProvider) IsAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to ping the Ollama API
	req, err := http.NewRequestWithContext(ctx, "GET", o.endpoint+"/api/tags", nil)
	if err != nil {
		return false
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// GetModelInfo returns information about the current model
func (o *OllamaProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		Name:        o.model,
		Provider:    "ollama",
		Size:        o.extractModelSize(o.model),
		Description: fmt.Sprintf("Ollama model: %s", o.model),
		Available:   o.IsAvailable(),
	}
}

// extractModelSize extracts the model size from the model name
// e.g., "codellama:7b" -> "7B"
func (o *OllamaProvider) extractModelSize(modelName string) string {
	parts := strings.Split(modelName, ":")
	if len(parts) > 1 {
		size := strings.ToUpper(parts[1])
		if !strings.HasSuffix(size, "B") {
			size += "B"
		}
		return size
	}
	return "Unknown"
}

// GenerateText sends a prompt to Ollama and returns the response
func (o *OllamaProvider) GenerateText(ctx context.Context, prompt string) (string, error) {
	// Create the request payload
	request := OllamaRequest{
		Model:  o.model,
		Prompt: prompt,
		Stream: false, // We want the complete response, not streaming
	}

	// Add any additional options from config
	if o.config.Parameters != nil {
		request.Options = o.config.Parameters
	}

	// Convert to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", o.endpoint+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return strings.TrimSpace(ollamaResp.Response), nil
}

// GenerateCommitMessage generates a commit message from a git diff using Ollama
func (o *OllamaProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	// Create a specialized prompt for commit message generation
	prompt := o.buildCommitPrompt(diff)

	// Generate the response
	response, err := o.GenerateText(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate commit message: %w", err)
	}

	// Clean up the response
	commitMessage := o.cleanCommitMessage(response)

	return commitMessage, nil
}

// buildCommitPrompt creates a specialized prompt for commit message generation
func (o *OllamaProvider) buildCommitPrompt(diff string) string {
	return fmt.Sprintf(`You are an expert software developer. Generate a concise, clear commit message for the following git diff.

Rules:
- Use conventional commit format: type(scope): description
- Types: feat, fix, docs, style, refactor, test, chore
- Keep the first line under 50 characters
- Be specific about what changed
- Don't include file names unless necessary

Git diff:
%s

Generate only the commit message, nothing else:`, diff)
}

// cleanCommitMessage cleans up the AI response to extract just the commit message
func (o *OllamaProvider) cleanCommitMessage(response string) string {
	// Remove common AI response prefixes/suffixes
	response = strings.TrimSpace(response)

	// Remove quotes if the AI wrapped the message in quotes
	if strings.HasPrefix(response, `"`) && strings.HasSuffix(response, `"`) {
		response = strings.Trim(response, `"`)
	}

	// Take only the first line if there are multiple lines
	lines := strings.Split(response, "\n")
	if len(lines) > 0 {
		response = strings.TrimSpace(lines[0])
	}

	return response
}
