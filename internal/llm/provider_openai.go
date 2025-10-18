package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/openai/openai-go"
)

// OpenAIProvider implements Provider for OpenAI
type OpenAIProvider struct {
	client       *openai.Client
	apiKey       string
	defaultModel string
	verbose      bool
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, defaultModel string, verbose bool) (*OpenAIProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("openai API key is required")
	}

	if defaultModel == "" {
		defaultModel = "gpt-4o"
	}

	// TODO: Determine correct client initialization
	// The openai-go SDK has different API than expected
	client := &openai.Client{}

	return &OpenAIProvider{
		client:       client,
		apiKey:       apiKey,
		defaultModel: defaultModel,
		verbose:      verbose,
	}, nil
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// IsAvailable checks if OpenAI API is accessible
func (p *OpenAIProvider) IsAvailable(ctx context.Context) bool {
	// Try to get a specific model as a health check with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := p.client.Models.Get(ctx, p.defaultModel)
	return err == nil
}

// Generate sends a prompt to OpenAI and returns the response
func (p *OpenAIProvider) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	_ = req.Model      // TODO: use this when implementing
	_ = p.defaultModel // TODO: use this when implementing

	// TODO: Implement OpenAI API calls - requires API testing
	// The openai-go SDK has complex types that need actual API key testing to implement correctly
	return nil, fmt.Errorf("openai provider not yet fully implemented - API signature needs testing with valid API key")
}

// ListModels returns available OpenAI models
func (p *OpenAIProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	models := []ModelInfo{
		{
			Name:         "gpt-4o",
			Provider:     "openai",
			Size:         "Unknown",
			Type:         "code",
			IsLocal:      false,
			IsDownloaded: true,
		},
		{
			Name:         "gpt-4o-mini",
			Provider:     "openai",
			Size:         "Unknown",
			Type:         "general",
			IsLocal:      false,
			IsDownloaded: true,
		},
		{
			Name:         "gpt-4",
			Provider:     "openai",
			Size:         "Unknown",
			Type:         "general",
			IsLocal:      false,
			IsDownloaded: true,
		},
	}

	return models, nil
}
