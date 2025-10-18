package config

// Config represents the SCIA configuration structure
type Config struct {
	LLM       LLMConfig       `yaml:"llm"`
	Cloud     CloudConfig     `yaml:"cloud"`
	Terraform TerraformConfig `yaml:"terraform"`
}

// LLMConfig holds LLM provider configuration
type LLMConfig struct {
	Provider string       `yaml:"provider"` // ollama, gemini, openai
	Ollama   OllamaConfig `yaml:"ollama,omitempty"`
	Gemini   GeminiConfig `yaml:"gemini,omitempty"`
	OpenAI   OpenAIConfig `yaml:"openai,omitempty"`
}

// OllamaConfig holds Ollama-specific configuration
type OllamaConfig struct {
	URL       string `yaml:"url,omitempty"`        // http://localhost:11434 or remote URL
	Model     string `yaml:"model,omitempty"`      // qwen2.5-coder:7b
	UseDocker bool   `yaml:"use_docker,omitempty"` // Whether to use Docker
}

// GeminiConfig holds Google Gemini configuration
type GeminiConfig struct {
	APIKey string `yaml:"api_key,omitempty"` // Google AI Studio API key
	Model  string `yaml:"model,omitempty"`   // gemini-2.0-pro-exp or gemini-2.0-flash
}

// OpenAIConfig holds OpenAI configuration
type OpenAIConfig struct {
	APIKey string `yaml:"api_key,omitempty"` // OpenAI API key
	Model  string `yaml:"model,omitempty"`   // gpt-4o or gpt-4o-mini
}

// CloudConfig holds cloud provider configuration
type CloudConfig struct {
	Provider      string `yaml:"provider"`       // aws, gcp
	DefaultRegion string `yaml:"default_region"` // AWS region (e.g., us-east-1)
}

// TerraformConfig holds Terraform/OpenTofu configuration
type TerraformConfig struct {
	Backend BackendConfig `yaml:"backend"`
	Binary  string        `yaml:"binary"` // tofu or terraform
}

// BackendConfig holds Terraform backend configuration
type BackendConfig struct {
	Type     string `yaml:"type"`      // s3
	S3Bucket string `yaml:"s3_bucket"` // S3 bucket name for state
	S3Region string `yaml:"s3_region"` // S3 bucket region
	S3Key    string `yaml:"s3_key"`    // State file path in bucket
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		LLM: LLMConfig{
			Provider: "ollama",
			Ollama: OllamaConfig{
				URL:       "http://localhost:11434",
				Model:     "qwen2.5-coder:7b",
				UseDocker: true,
			},
			Gemini: GeminiConfig{
				Model: "gemini-2.0-pro-exp",
			},
			OpenAI: OpenAIConfig{
				Model: "gpt-4o",
			},
		},
		Cloud: CloudConfig{
			Provider:      "aws",
			DefaultRegion: "eu-west-3",
		},
		Terraform: TerraformConfig{
			Backend: BackendConfig{
				Type:  "s3",
				S3Key: "terraform.tfstate",
			},
			Binary: "tofu",
		},
	}
}
