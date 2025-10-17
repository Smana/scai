package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HuggingFaceProvider implements Provider for HuggingFace Inference API
type HuggingFaceProvider struct {
	apiToken     string
	endpoint     string
	defaultModel string
	httpClient   *http.Client
	verbose      bool
}

// NewHuggingFaceProvider creates a new HuggingFace provider
func NewHuggingFaceProvider(apiToken, defaultModel string, verbose bool) (*HuggingFaceProvider, error) {
	if defaultModel == "" {
		defaultModel = "mistralai/Mistral-7B-Instruct-v0.2"
	}

	return &HuggingFaceProvider{
		apiToken:     apiToken,
		endpoint:     "https://api-inference.huggingface.co/models",
		defaultModel: defaultModel,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		verbose: verbose,
	}, nil
}

// Name returns the provider name
func (p *HuggingFaceProvider) Name() string {
	return "huggingface"
}

// IsAvailable checks if HuggingFace API is accessible
func (p *HuggingFaceProvider) IsAvailable(ctx context.Context) bool {
	// Check if API token is set
	if p.apiToken == "" {
		return false
	}

	// Try a simple API call
	req, err := http.NewRequestWithContext(ctx, "GET", p.endpoint+"/"+p.defaultModel, nil)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Bearer "+p.apiToken)
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode == http.StatusOK
}

// Generate sends a prompt to HuggingFace and returns the response
func (p *HuggingFaceProvider) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	model := req.Model
	if model == "" {
		model = p.defaultModel
	}

	// Build request payload
	payload := map[string]interface{}{
		"inputs": req.Prompt,
		"parameters": map[string]interface{}{
			"temperature":    req.Temperature,
			"max_new_tokens": req.MaxTokens,
			"top_p":          req.TopP,
		},
	}

	// Add system message if provided
	if req.System != "" {
		payload["inputs"] = fmt.Sprintf("%s\n\nUser: %s", req.System, req.Prompt)
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		p.endpoint+"/"+model,
		bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.apiToken)
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("huggingface API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("huggingface API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("empty response from HuggingFace")
	}

	// Extract generated text
	generatedText := ""
	if text, ok := result[0]["generated_text"].(string); ok {
		generatedText = text
	}

	return &GenerateResponse{
		Text:  generatedText,
		Model: model,
	}, nil
}

// ListModels returns available HuggingFace models
// Note: This is a simplified implementation. Full implementation would
// query the HuggingFace Hub API for model listings
func (p *HuggingFaceProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	// Predefined list of popular models
	// In a full implementation, this would query the HF Hub API
	popularModels := []string{
		"mistralai/Mistral-7B-Instruct-v0.2",
		"meta-llama/Llama-2-7b-chat-hf",
		"codellama/CodeLlama-7b-Instruct-hf",
		"bigcode/starcoder",
	}

	var models []ModelInfo
	for _, modelName := range popularModels {
		models = append(models, ModelInfo{
			Name:         modelName,
			Provider:     "huggingface",
			Size:         extractModelSize(modelName),
			Type:         extractModelType(modelName),
			IsLocal:      false,
			IsDownloaded: false,
		})
	}

	return models, nil
}
