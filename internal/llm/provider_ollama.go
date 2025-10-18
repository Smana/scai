package llm

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
)

var logger = log.Default()

// OllamaProvider implements Provider for Ollama
type OllamaProvider struct {
	client       *api.Client
	baseURL      string
	defaultModel string
	verbose      bool
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(baseURL, defaultModel string, verbose bool) (*OllamaProvider, error) {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	if defaultModel == "" {
		defaultModel = "qwen2.5-coder:7b"
	}

	// Parse URL
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Ollama URL: %w", err)
	}

	// Create client
	client := api.NewClient(u, http.DefaultClient)

	return &OllamaProvider{
		client:       client,
		baseURL:      baseURL,
		defaultModel: defaultModel,
		verbose:      verbose,
	}, nil
}

// Name returns the provider name
func (p *OllamaProvider) Name() string {
	return "ollama"
}

// IsAvailable checks if Ollama is accessible
func (p *OllamaProvider) IsAvailable(ctx context.Context) bool {
	// Try to list models as a health check
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := p.client.List(ctx)
	return err == nil
}

// Generate sends a prompt to Ollama and returns the response
func (p *OllamaProvider) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	model := req.Model
	if model == "" {
		model = p.defaultModel
	}

	// Build Ollama request
	ollamaReq := &api.GenerateRequest{
		Model:  model,
		Prompt: req.Prompt,
		System: req.System,
		Options: map[string]interface{}{
			"temperature": req.Temperature,
			"num_predict": req.MaxTokens,
		},
	}

	// Add custom options if provided
	if req.Options != nil {
		for k, v := range req.Options {
			ollamaReq.Options[k] = v
		}
	}

	// Collect response
	var fullResponse string
	var promptTokens, totalTokens int

	err := p.client.Generate(ctx, ollamaReq, func(resp api.GenerateResponse) error {
		fullResponse += resp.Response

		// Track tokens if available
		if resp.PromptEvalCount > 0 {
			promptTokens = resp.PromptEvalCount
		}
		if resp.EvalCount > 0 {
			totalTokens += resp.EvalCount
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("ollama generation failed: %w", err)
	}

	return &GenerateResponse{
		Text:         fullResponse,
		Model:        model,
		TokensPrompt: promptTokens,
		TokensTotal:  totalTokens,
	}, nil
}

// ListModels returns available Ollama models
func (p *OllamaProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	resp, err := p.client.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list ollama models: %w", err)
	}

	models := make([]ModelInfo, 0, len(resp.Models))
	for _, model := range resp.Models {
		models = append(models, ModelInfo{
			Name:         model.Name,
			Provider:     "ollama",
			Size:         extractModelSize(model.Name),
			Type:         extractModelType(model.Name),
			IsLocal:      true,
			IsDownloaded: true,
		})
	}

	return models, nil
}

// Helper functions

func extractModelSize(modelName string) string {
	// Extract size from model name (e.g., "qwen2.5-coder:7b" -> "7b")
	// Simple heuristic - look for patterns like :7b, :13b, etc.
	if len(modelName) > 2 {
		// Check for :Xb pattern
		for i := len(modelName) - 1; i >= 0; i-- {
			if modelName[i] == ':' && i+1 < len(modelName) {
				return modelName[i+1:]
			}
		}
	}
	return "unknown"
}

func extractModelType(modelName string) string {
	// Determine model type from name
	switch {
	case strings.Contains(modelName, "code"):
		return "code"
	case strings.Contains(modelName, "instruct"):
		return "instruct"
	case strings.Contains(modelName, "chat"):
		return "chat"
	default:
		return "general"
	}
}
