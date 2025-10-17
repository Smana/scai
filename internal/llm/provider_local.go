package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// LocalProvider implements Provider for local GGUF models
// Uses llama.cpp compatible server (e.g., llama-server, text-generation-webui)
type LocalProvider struct {
	modelPath  string
	serverURL  string
	httpClient *http.Client
	verbose    bool
}

// NewLocalProvider creates a new local GGUF model provider
func NewLocalProvider(modelPath, serverURL string, verbose bool) (*LocalProvider, error) {
	if serverURL == "" {
		serverURL = "http://localhost:8080" // llama.cpp default port
	}

	// Check if model file exists
	if modelPath != "" {
		if _, err := os.Stat(modelPath); err != nil {
			return nil, fmt.Errorf("model file not found: %s", modelPath)
		}
	}

	return &LocalProvider{
		modelPath: modelPath,
		serverURL: serverURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // Local models can be slow on CPU
		},
		verbose: verbose,
	}, nil
}

// Name returns the provider name
func (p *LocalProvider) Name() string {
	return "local"
}

// IsAvailable checks if local server is accessible
func (p *LocalProvider) IsAvailable(ctx context.Context) bool {
	// Check if server is running
	req, err := http.NewRequestWithContext(ctx, "GET", p.serverURL+"/health", nil)
	if err != nil {
		return false
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode == http.StatusOK
}

// Generate sends a prompt to local server and returns the response
func (p *LocalProvider) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	// Build request payload for llama.cpp compatible server
	payload := map[string]interface{}{
		"prompt":      req.Prompt,
		"temperature": req.Temperature,
		"n_predict":   req.MaxTokens,
		"top_p":       req.TopP,
		"top_k":       req.TopK,
	}

	// Add system message if provided
	if req.System != "" {
		payload["system_prompt"] = req.System
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		p.serverURL+"/completion",
		bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("local server request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("local server error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract generated text
	generatedText := ""
	if content, ok := result["content"].(string); ok {
		generatedText = content
	} else if text, ok := result["text"].(string); ok {
		generatedText = text
	}

	// Extract token counts if available
	tokensPrompt := 0
	tokensTotal := 0
	if tp, ok := result["tokens_evaluated"].(float64); ok {
		tokensPrompt = int(tp)
	}
	if tt, ok := result["tokens_predicted"].(float64); ok {
		tokensTotal = int(tt)
	}

	modelName := filepath.Base(p.modelPath)
	if modelName == "" {
		modelName = "local-model"
	}

	return &GenerateResponse{
		Text:         generatedText,
		Model:        modelName,
		TokensPrompt: tokensPrompt,
		TokensTotal:  tokensTotal,
	}, nil
}

// ListModels returns available local models
func (p *LocalProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	var models []ModelInfo

	// If model path is provided, return it
	if p.modelPath != "" {
		modelName := filepath.Base(p.modelPath)
		models = append(models, ModelInfo{
			Name:         modelName,
			Provider:     "local",
			Size:         extractSizeFromFilename(modelName),
			Type:         extractModelType(modelName),
			IsLocal:      true,
			IsDownloaded: true,
		})
	}

	// Try to query server for model list
	req, err := http.NewRequestWithContext(ctx, "GET", p.serverURL+"/v1/models", nil)
	if err != nil {
		return models, nil // Return what we have
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return models, nil // Return what we have
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return models, nil
	}

	// Parse server model list
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return models, nil
	}

	// Extract models from server response
	if data, ok := result["data"].([]interface{}); ok {
		for _, item := range data {
			if modelData, ok := item.(map[string]interface{}); ok {
				if modelID, ok := modelData["id"].(string); ok {
					models = append(models, ModelInfo{
						Name:         modelID,
						Provider:     "local",
						Size:         extractSizeFromFilename(modelID),
						Type:         extractModelType(modelID),
						IsLocal:      true,
						IsDownloaded: true,
					})
				}
			}
		}
	}

	return models, nil
}

// Helper function to extract size from GGUF filename
// e.g., "mistral-7b-instruct-v0.2.Q4_K_M.gguf" -> "7B-Q4"
func extractSizeFromFilename(filename string) string {
	// Look for patterns like 7b, 13b, etc.
	size := extractModelSize(filename)

	// Look for quantization info (Q4, Q5, etc.)
	for i := 0; i < len(filename)-1; i++ {
		if filename[i] == 'Q' && i+1 < len(filename) {
			// Extract quantization type
			quant := ""
			for j := i; j < len(filename) && filename[j] != '.'; j++ {
				quant += string(filename[j])
			}
			if quant != "" {
				if size != "unknown" {
					return size + "-" + quant
				}
				return quant
			}
		}
	}

	return size
}
