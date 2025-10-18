package llm

import (
	"context"
	"fmt"
)

// Provider defines the interface for LLM providers
// Supports: Ollama (local), HuggingFace (API), Local GGUF models
type Provider interface {
	// Generate sends a prompt and returns the response
	Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)

	// ListModels returns available models for this provider
	ListModels(ctx context.Context) ([]ModelInfo, error)

	// Name returns the provider name
	Name() string

	// IsAvailable checks if the provider is accessible
	IsAvailable(ctx context.Context) bool
}

// GenerateRequest is provider-agnostic generation request
type GenerateRequest struct {
	Model       string                 // Model name/identifier
	Prompt      string                 // Text prompt
	System      string                 // System message (optional)
	Temperature float64                // Sampling temperature (0.0-2.0)
	MaxTokens   int                    // Maximum tokens to generate
	TopP        float64                // Nucleus sampling threshold
	TopK        int                    // Top-K sampling
	Options     map[string]interface{} // Provider-specific options
}

// GenerateResponse is provider-agnostic generation response
type GenerateResponse struct {
	Text         string // Generated text
	Model        string // Model used
	TokensPrompt int    // Tokens in prompt
	TokensTotal  int    // Total tokens
	Error        error  // Error if any
}

// ModelInfo describes an available model
type ModelInfo struct {
	Name         string   // Model identifier
	Provider     string   // Provider name (ollama, huggingface, local)
	Size         string   // Model size (e.g., "7B", "13B")
	Type         string   // Model type (e.g., "instruct", "chat", "code")
	Tags         []string // Additional tags
	IsLocal      bool     // Whether model is available locally
	IsDownloaded bool     // Whether model is fully downloaded
}

// ProviderConfig holds provider-specific configuration
type ProviderConfig struct {
	// Provider type: "ollama", "gemini", "openai", "huggingface", "local"
	Type string

	// Ollama configuration
	OllamaURL   string // Default: http://localhost:11434
	OllamaModel string // Default model for Ollama

	// Gemini configuration
	GeminiAPIKey string // Google AI Studio API key
	GeminiModel  string // Default model (gemini-2.0-pro-exp)

	// OpenAI configuration
	OpenAIAPIKey string // OpenAI API key
	OpenAIModel  string // Default model (gpt-4o)

	// HuggingFace configuration
	HFToken    string // HuggingFace API token (optional)
	HFEndpoint string // Custom endpoint (optional)
	HFModel    string // Default model

	// Local GGUF configuration
	LocalModelPath string // Path to local GGUF model file
	LocalServerURL string // llama.cpp compatible server URL

	// General settings
	DefaultModel string  // Fallback model name
	Timeout      int     // Request timeout in seconds
	Retries      int     // Number of retries
	Temperature  float64 // Default temperature
}

// ProviderManager manages multiple LLM providers with fallback
type ProviderManager struct {
	providers []Provider
	config    *ProviderConfig
	verbose   bool
}

// NewProviderManager creates a manager with configured providers
func NewProviderManager(config *ProviderConfig, verbose bool) (*ProviderManager, error) {
	pm := &ProviderManager{
		config:  config,
		verbose: verbose,
	}

	// Initialize providers based on config
	var providers []Provider

	// Add Ollama if configured
	if config.Type == "ollama" || config.Type == "" {
		ollamaProvider, err := NewOllamaProvider(config.OllamaURL, config.OllamaModel, verbose)
		if err == nil {
			providers = append(providers, ollamaProvider)
		}
	}

	// Add Gemini if configured
	if config.Type == "gemini" {
		geminiProvider, err := NewGeminiProvider(config.GeminiAPIKey, config.GeminiModel, verbose)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Gemini provider: %w", err)
		}
		providers = append(providers, geminiProvider)
	}

	// Add OpenAI if configured
	if config.Type == "openai" {
		openaiProvider, err := NewOpenAIProvider(config.OpenAIAPIKey, config.OpenAIModel, verbose)
		if err == nil {
			providers = append(providers, openaiProvider)
		}
	}

	// Add HuggingFace if configured
	if config.Type == "huggingface" {
		hfProvider, err := NewHuggingFaceProvider(config.HFToken, config.HFModel, verbose)
		if err == nil {
			providers = append(providers, hfProvider)
		}
	}

	// Add local GGUF if configured
	if config.Type == "local" && config.LocalModelPath != "" {
		localProvider, err := NewLocalProvider(config.LocalModelPath, config.LocalServerURL, verbose)
		if err == nil {
			providers = append(providers, localProvider)
		}
	}

	if len(providers) == 0 {
		return nil, ErrNoProvidersAvailable
	}

	pm.providers = providers
	return pm, nil
}

// Generate tries providers in order until success
func (pm *ProviderManager) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	var lastErr error

	for _, provider := range pm.providers {
		// Check if provider is available
		if !provider.IsAvailable(ctx) {
			if pm.verbose {
				logger.Printf("Provider %s not available, trying next...", provider.Name())
			}
			continue
		}

		// Try generation
		resp, err := provider.Generate(ctx, req)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if pm.verbose {
			logger.Printf("Provider %s failed: %v, trying next...", provider.Name(), err)
		}
	}

	if lastErr == nil {
		lastErr = ErrNoProvidersAvailable
	}

	return nil, lastErr
}

// ListAllModels returns models from all available providers
func (pm *ProviderManager) ListAllModels(ctx context.Context) ([]ModelInfo, error) {
	var allModels []ModelInfo

	for _, provider := range pm.providers {
		if !provider.IsAvailable(ctx) {
			continue
		}

		models, err := provider.ListModels(ctx)
		if err != nil {
			continue
		}

		allModels = append(allModels, models...)
	}

	return allModels, nil
}

// GetBestProvider returns the first available provider
func (pm *ProviderManager) GetBestProvider(ctx context.Context) Provider {
	for _, provider := range pm.providers {
		if provider.IsAvailable(ctx) {
			return provider
		}
	}
	return nil
}
