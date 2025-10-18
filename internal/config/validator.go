package config

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// AWS region pattern (e.g., us-east-1, eu-west-3)
	awsRegionPattern = regexp.MustCompile(`^[a-z]{2}-[a-z]+-\d$`)

	// S3 bucket name validation
	// Bucket names must be 3-63 characters, lowercase, no underscores
	s3BucketPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`)
)

// ValidateConfig validates the entire configuration
func ValidateConfig(cfg *Config) error {
	// Validate LLM configuration
	if err := validateLLM(&cfg.LLM); err != nil {
		return fmt.Errorf("llm config invalid: %w", err)
	}

	// Validate Cloud configuration
	if err := validateCloud(&cfg.Cloud); err != nil {
		return fmt.Errorf("cloud config invalid: %w", err)
	}

	// Validate Terraform configuration
	if err := validateTerraform(&cfg.Terraform); err != nil {
		return fmt.Errorf("terraform config invalid: %w", err)
	}

	return nil
}

// validateLLM validates LLM provider configuration
//
//nolint:gocyclo // Validation logic requires checking each provider's specific requirements
func validateLLM(llm *LLMConfig) error {
	// Provider must be set
	if llm.Provider == "" {
		return fmt.Errorf("llm provider is required")
	}

	// Validate provider is one of the supported types
	validProviders := []string{"ollama", "gemini", "openai"}
	if !contains(validProviders, llm.Provider) {
		return fmt.Errorf("llm provider must be one of: %s", strings.Join(validProviders, ", "))
	}

	// Provider-specific validation
	switch llm.Provider {
	case "ollama":
		if llm.Ollama.URL == "" {
			return fmt.Errorf("ollama url is required when using ollama provider")
		}
		if llm.Ollama.Model == "" {
			return fmt.Errorf("ollama model is required when using ollama provider")
		}
	case "gemini":
		if llm.Gemini.APIKey == "" {
			return fmt.Errorf("gemini api_key is required when using gemini provider")
		}
		if llm.Gemini.Model == "" {
			return fmt.Errorf("gemini model is required when using gemini provider")
		}
	case "openai":
		if llm.OpenAI.APIKey == "" {
			return fmt.Errorf("openai api_key is required when using openai provider")
		}
		if llm.OpenAI.Model == "" {
			return fmt.Errorf("openai model is required when using openai provider")
		}
	}

	return nil
}

// validateCloud validates cloud provider configuration
func validateCloud(cloud *CloudConfig) error {
	// Provider must be set
	if cloud.Provider == "" {
		return fmt.Errorf("cloud provider is required")
	}

	// Validate provider is supported
	validProviders := []string{"aws", "gcp"}
	if !contains(validProviders, cloud.Provider) {
		return fmt.Errorf("cloud provider must be one of: %s", strings.Join(validProviders, ", "))
	}

	// AWS-specific validation
	if cloud.Provider == "aws" {
		if cloud.DefaultRegion == "" {
			return fmt.Errorf("default_region is required for aws provider")
		}

		// Basic format validation for AWS region
		if !awsRegionPattern.MatchString(cloud.DefaultRegion) {
			return fmt.Errorf("invalid aws region format: %s (expected format: us-east-1)", cloud.DefaultRegion)
		}
	}

	return nil
}

// validateTerraform validates Terraform configuration
func validateTerraform(tf *TerraformConfig) error {
	// Binary must be set
	if tf.Binary == "" {
		return fmt.Errorf("terraform binary is required")
	}

	// Validate binary is either terraform or tofu
	if tf.Binary != "terraform" && tf.Binary != "tofu" {
		return fmt.Errorf("terraform binary must be 'terraform' or 'tofu'")
	}

	// Validate backend configuration
	if err := validateBackend(&tf.Backend); err != nil {
		return fmt.Errorf("backend config invalid: %w", err)
	}

	return nil
}

// validateBackend validates Terraform backend configuration
func validateBackend(backend *BackendConfig) error {
	// Type must be set
	if backend.Type == "" {
		return fmt.Errorf("backend type is required")
	}

	// Only S3 backend is supported currently
	if backend.Type != "s3" {
		return fmt.Errorf("only 's3' backend is supported")
	}

	// S3-specific validation
	if backend.S3Bucket == "" {
		return fmt.Errorf("s3_bucket is required for s3 backend")
	}

	// Validate S3 bucket name format
	if !s3BucketPattern.MatchString(backend.S3Bucket) {
		return fmt.Errorf("invalid s3 bucket name: %s (must be 3-63 lowercase alphanumeric characters with hyphens)", backend.S3Bucket)
	}

	if backend.S3Region == "" {
		return fmt.Errorf("s3_region is required for s3 backend")
	}

	// Validate S3 region format
	if !awsRegionPattern.MatchString(backend.S3Region) {
		return fmt.Errorf("invalid s3 region format: %s (expected format: us-east-1)", backend.S3Region)
	}

	if backend.S3Key == "" {
		return fmt.Errorf("s3_key is required for s3 backend")
	}

	return nil
}

// contains checks if a string slice contains a value
func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
