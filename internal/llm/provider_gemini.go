package llm

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/genai"
)

// GeminiProvider implements Provider for Google Gemini
type GeminiProvider struct {
	client       *genai.Client
	apiKey       string
	defaultModel string
	verbose      bool
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(apiKey, defaultModel string, verbose bool) (*GeminiProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("gemini API key is required")
	}

	if defaultModel == "" {
		defaultModel = "gemini-2.0-pro-exp"
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI, // Use Gemini Developer API (not Vertex AI)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiProvider{
		client:       client,
		apiKey:       apiKey,
		defaultModel: defaultModel,
		verbose:      verbose,
	}, nil
}

// Name returns the provider name
func (p *GeminiProvider) Name() string {
	return "gemini"
}

// IsAvailable checks if Gemini API is accessible
func (p *GeminiProvider) IsAvailable(ctx context.Context) bool {
	// Try to list models as a health check with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := p.client.Models.List(ctx, nil)
	if err != nil {
		if p.verbose {
			logger.Printf("Gemini availability check failed: %v", err)
		}
		return false
	}

	if p.verbose {
		logger.Printf("Gemini API is available")
	}

	return true
}

// Generate sends a prompt to Gemini and returns the response
func (p *GeminiProvider) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	// Use requested model or fall back to default
	modelName := req.Model
	if modelName == "" {
		modelName = p.defaultModel
	}

	// Build the prompt - combine system and user message
	prompt := req.Prompt
	if req.System != "" {
		prompt = req.System + "\n\n" + req.Prompt
	}

	// Build generation config
	config := &genai.GenerateContentConfig{}

	if req.Temperature > 0 {
		config.Temperature = genai.Ptr(float32(req.Temperature))
	}

	if req.MaxTokens > 0 {
		config.MaxOutputTokens = int32(req.MaxTokens)
	}

	if req.TopP > 0 {
		config.TopP = genai.Ptr(float32(req.TopP))
	}

	// Generate content
	if p.verbose {
		logger.Printf("Gemini: Generating with model %s (temp=%.2f, max_tokens=%d)",
			modelName, req.Temperature, req.MaxTokens)
	}

	resp, err := p.client.Models.GenerateContent(ctx, modelName, genai.Text(prompt), config)
	if err != nil {
		return nil, fmt.Errorf("gemini generation failed: %w", err)
	}

	// Extract text from response
	text := resp.Text()
	if text == "" {
		return nil, fmt.Errorf("gemini returned empty response")
	}

	if p.verbose {
		logger.Printf("Gemini: Generated %d characters", len(text))
	}

	return &GenerateResponse{
		Text:         text,
		Model:        modelName,
		TokensPrompt: 0, // Gemini SDK doesn't expose token counts easily in basic response
		TokensTotal:  0,
	}, nil
}

// ListModels returns available Gemini models
func (p *GeminiProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	models := []ModelInfo{
		{
			Name:         "gemini-2.0-pro-exp",
			Provider:     "gemini",
			Size:         "Unknown",
			Type:         "code",
			IsLocal:      false,
			IsDownloaded: true,
		},
		{
			Name:         "gemini-2.0-flash",
			Provider:     "gemini",
			Size:         "Unknown",
			Type:         "general",
			IsLocal:      false,
			IsDownloaded: true,
		},
		{
			Name:         "gemini-2.5-pro",
			Provider:     "gemini",
			Size:         "Unknown",
			Type:         "code",
			IsLocal:      false,
			IsDownloaded: true,
		},
	}

	return models, nil
}
