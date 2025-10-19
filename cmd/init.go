package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/Smana/scai/internal/backend"
	"github.com/Smana/scai/internal/cloud"
	"github.com/Smana/scai/internal/config"
	"github.com/Smana/scai/internal/llm"
	"github.com/Smana/scai/internal/requirements"
)

const (
	providerOllama = "ollama"
	providerGemini = "gemini"
	providerOpenAI = "openai"
	regionUSEast1  = "us-east-1"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize SCAI configuration",
	Long: `Interactive wizard to help onboard new users by configuring:
- LLM provider (Ollama, Gemini, or OpenAI)
- Cloud provider (AWS or GCP)
- Default region
- Terraform backend (S3 bucket)
- Requirements check (OpenTofu, Docker, etc.)

The configuration will be saved to ~/.scai.yaml`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	fmt.Println("🚀 SCAI Configuration Wizard")
	fmt.Println("This wizard will help you set up SCAI for the first time.")
	fmt.Println()

	// Check if config already exists
	if config.ConfigExists() {
		var overwrite bool
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Configuration file already exists").
					Description("Do you want to overwrite it?").
					Value(&overwrite),
			),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("failed to get confirmation: %w", err)
		}

		if !overwrite {
			fmt.Println("\n✓ Configuration unchanged")
			return nil
		}
	}

	// Initialize configuration with defaults
	cfg := config.DefaultConfig()

	// Step 1: LLM Provider Selection
	if err := configureLLMProvider(cfg); err != nil {
		return fmt.Errorf("llm configuration failed: %w", err)
	}

	// Step 2: Cloud Provider Selection
	if err := configureCloudProvider(ctx, cfg); err != nil {
		return fmt.Errorf("cloud configuration failed: %w", err)
	}

	// Step 3: Terraform Backend Configuration
	if err := configureTerraformBackend(ctx, cfg); err != nil {
		return fmt.Errorf("terraform backend configuration failed: %w", err)
	}

	// Step 4: Requirements Check
	if err := checkRequirements(cfg); err != nil {
		return fmt.Errorf("requirements check failed: %w", err)
	}

	// Validate configuration
	if err := config.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Write configuration
	if err := config.WriteConfig(cfg); err != nil {
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	// Display summary
	displaySummary(cfg)

	return nil
}

func configureLLMProvider(cfg *config.Config) error {
	fmt.Println("📋 Step 1: LLM Provider Configuration")
	fmt.Println()

	var provider string
	providerForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select LLM Provider").
				Description("Choose the LLM provider for natural language parsing").
				Options(
					huh.NewOption("Ollama (Local/Docker)", "ollama"),
					huh.NewOption("Google Gemini", "gemini"),
					huh.NewOption("OpenAI", "openai"),
				).
				Value(&provider),
		),
	)

	if err := providerForm.Run(); err != nil {
		return err
	}

	cfg.LLM.Provider = provider

	// Provider-specific configuration
	switch provider {
	case providerOllama:
		return configureOllama(cfg)
	case providerGemini:
		return configureGemini(cfg)
	case providerOpenAI:
		return configureOpenAI(cfg)
	}

	return nil
}

func configureOllama(cfg *config.Config) error {
	var useDocker bool
	dockerForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[bool]().
				Title("Ollama Deployment").
				Description("How should Ollama run?").
				Options(
					huh.NewOption("Docker (Recommended)", true),
					huh.NewOption("Remote Server", false),
				).
				Value(&useDocker),
		),
	)

	if err := dockerForm.Run(); err != nil {
		return err
	}

	cfg.LLM.Ollama.UseDocker = useDocker

	if !useDocker {
		var url string
		urlForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Ollama Server URL").
					Description("Enter the Ollama server address").
					Value(&url).
					Placeholder("http://localhost:11434"),
			),
		)

		if err := urlForm.Run(); err != nil {
			return err
		}

		if url == "" {
			url = defaultOllamaURL
		}
		cfg.LLM.Ollama.URL = url

		// Test connection
		if !llm.IsOllamaAccessible(url) {
			fmt.Printf("\n⚠️  Warning: Could not connect to Ollama at %s\n", url)
			fmt.Println("   Make sure Ollama is running before using SCAI")
		}
	} else {
		cfg.LLM.Ollama.URL = "http://localhost:11434"
	}

	// Model selection
	var model string
	modelForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Ollama Model").
				Description("Model name for code analysis").
				Value(&model).
				Placeholder("qwen2.5-coder:7b"),
		),
	)

	if err := modelForm.Run(); err != nil {
		return err
	}

	if model == "" {
		model = "qwen2.5-coder:7b"
	}
	cfg.LLM.Ollama.Model = model

	return nil
}

// configureCloudLLMProvider is a helper to configure cloud-based LLM providers (Gemini, OpenAI)
// It handles the common pattern of API key input and model selection
func configureCloudLLMProvider(
	apiKeyTitle, apiKeyDescription string,
	modelTitle string,
	modelOptions []huh.Option[string],
) (apiKey string, model string, err error) {
	// API Key input
	apiKeyForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(apiKeyTitle).
				Description(apiKeyDescription).
				Value(&apiKey).
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("API key is required")
					}
					return nil
				}),
		),
	)

	if err := apiKeyForm.Run(); err != nil {
		return "", "", err
	}

	// Model selection
	modelForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(modelTitle).
				Description("Choose the model based on your needs").
				Options(modelOptions...).
				Value(&model),
		),
	)

	if err := modelForm.Run(); err != nil {
		return "", "", err
	}

	return apiKey, model, nil
}

func configureGemini(cfg *config.Config) error {
	apiKey, model, err := configureCloudLLMProvider(
		"Google AI Studio API Key",
		"Get your key at: https://aistudio.google.com/apikey",
		"Select Gemini Model",
		[]huh.Option[string]{
			huh.NewOption("gemini-2.0-pro-exp (Recommended)", "gemini-2.0-pro-exp"),
			huh.NewOption("gemini-2.0-flash", "gemini-2.0-flash"),
			huh.NewOption("gemini-2.5-pro", "gemini-2.5-pro"),
		},
	)
	if err != nil {
		return err
	}

	cfg.LLM.Gemini.APIKey = apiKey
	cfg.LLM.Gemini.Model = model

	return nil
}

func configureOpenAI(cfg *config.Config) error {
	apiKey, model, err := configureCloudLLMProvider(
		"OpenAI API Key",
		"Get your key at: https://platform.openai.com/api-keys",
		"Select OpenAI Model",
		[]huh.Option[string]{
			huh.NewOption("gpt-4o (Recommended)", "gpt-4o"),
			huh.NewOption("gpt-4o-mini", "gpt-4o-mini"),
			huh.NewOption("gpt-4", "gpt-4"),
		},
	)
	if err != nil {
		return err
	}

	cfg.LLM.OpenAI.APIKey = apiKey
	cfg.LLM.OpenAI.Model = model

	return nil
}

func configureCloudProvider(ctx context.Context, cfg *config.Config) error {
	fmt.Println("\n📋 Step 2: Cloud Provider Configuration")
	fmt.Println()

	var provider string
	providerForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Cloud Provider").
				Description("Choose your cloud platform").
				Options(
					huh.NewOption("AWS", "aws"),
					huh.NewOption("GCP (Coming Soon)", "gcp"),
				).
				Value(&provider),
		),
	)

	if err := providerForm.Run(); err != nil {
		return err
	}

	if provider == "gcp" {
		fmt.Println("\n⚠️  GCP support is not yet implemented. Please choose AWS.")
		return fmt.Errorf("GCP not yet supported")
	}

	cfg.Cloud.Provider = provider

	// AWS Region Selection - MANDATORY
	fmt.Println("\n🔐 Checking AWS credentials...")
	awsClient, err := cloud.NewAWSClient(ctx)
	if err != nil {
		fmt.Printf("\n❌ Error: Could not connect to AWS: %v\n\n", err)
		fmt.Println("AWS credentials are required to continue.")
		fmt.Println("Please configure your AWS credentials using one of these methods:")
		fmt.Println("  1. Run: aws configure")
		fmt.Println("  2. Set environment variables: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY")
		fmt.Println("  3. Use AWS SSO: aws sso login")
		fmt.Println()
		return fmt.Errorf("AWS credentials not configured")
	}

	fmt.Println("✓ AWS credentials verified")
	fmt.Println("\n🌍 Fetching available AWS regions...")
	regionOpts, err := awsClient.GetRegionForSelect(ctx)
	if err != nil {
		fmt.Printf("\n❌ Error: Could not fetch AWS regions: %v\n\n", err)
		fmt.Println("This is required to continue. Please check:")
		fmt.Println("  1. Your AWS credentials have permission to list regions (ec2:DescribeRegions)")
		fmt.Println("  2. Your network connection is working")
		fmt.Println()
		return fmt.Errorf("failed to fetch AWS regions: %w", err)
	}

	fmt.Printf("✓ Found %d available regions\n", len(regionOpts))

	// Build region options for huh select
	regionOptions := make([]huh.Option[string], 0, len(regionOpts))
	for _, region := range regionOpts {
		regionOptions = append(regionOptions, huh.NewOption(region.Code, region.Code))
	}

	var selectedRegion string
	regionForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select AWS Region").
				Description("Choose your default AWS region").
				Options(regionOptions...).
				Value(&selectedRegion).
				Height(15),
		),
	)

	if err := regionForm.Run(); err != nil {
		return err
	}

	if selectedRegion == "" {
		return fmt.Errorf("region selection is required")
	}

	cfg.Cloud.DefaultRegion = selectedRegion
	fmt.Printf("\n✓ Region set to: %s\n", selectedRegion)

	return nil
}

func configureTerraformBackend(ctx context.Context, cfg *config.Config) error {
	fmt.Println("\n📋 Step 3: Terraform Backend Configuration")
	fmt.Println()

	// Ask if they want to create a new bucket or use an existing one
	var useExisting bool
	bucketChoiceForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[bool]().
				Title("S3 Bucket Configuration").
				Description(fmt.Sprintf("Choose S3 bucket option for Terraform state in %s", cfg.Cloud.DefaultRegion)).
				Options(
					huh.NewOption("Create new S3 bucket", false),
					huh.NewOption("Use existing S3 bucket", true),
				).
				Value(&useExisting),
		),
	)

	if err := bucketChoiceForm.Run(); err != nil {
		return err
	}

	var bucketName string
	var bucketRegion string = cfg.Cloud.DefaultRegion

	if useExisting {
		// List existing S3 buckets
		fmt.Println("\n🪣 Fetching S3 buckets...")

		// Create S3 manager (region doesn't matter for ListBuckets)
		s3Manager, err := backend.NewS3Manager(ctx, bucketRegion)
		if err != nil {
			return fmt.Errorf("failed to connect to S3: %w\nPlease ensure your AWS credentials are configured correctly", err)
		}

		buckets, err := s3Manager.ListBuckets(ctx)
		if err != nil {
			return fmt.Errorf("failed to list S3 buckets: %w\nPlease check your AWS permissions (s3:ListAllMyBuckets)", err)
		}

		if len(buckets) == 0 {
			fmt.Println("\n⚠️  No S3 buckets found in your AWS account")
			fmt.Println("   You can create a new bucket instead")
			return fmt.Errorf("no existing buckets available")
		}

		fmt.Printf("✓ Found %d buckets\n", len(buckets))

		// Build bucket options for selection
		bucketOptions := make([]huh.Option[string], 0, len(buckets))
		for _, bucket := range buckets {
			bucketOptions = append(bucketOptions, huh.NewOption(bucket, bucket))
		}

		var selectedBucket string
		bucketForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select S3 Bucket").
					Description("Choose an existing bucket for Terraform state").
					Options(bucketOptions...).
					Value(&selectedBucket).
					Height(15),
			),
		)

		if err := bucketForm.Run(); err != nil {
			return err
		}

		if selectedBucket == "" {
			return fmt.Errorf("bucket selection is required")
		}

		bucketName = selectedBucket

		// Get the bucket's region (required for backend configuration)
		// Note: ListBuckets doesn't return region, so we need to query it
		fmt.Println("\n🔍 Determining bucket region...")

		// Try to get bucket location
		// We need to create a new S3 manager without region specified
		tempManager, err := backend.NewS3Manager(ctx, regionUSEast1) // us-east-1 works globally
		if err != nil {
			return fmt.Errorf("failed to connect to S3: %w", err)
		}

		// Get bucket location
		locationResp, err := tempManager.GetBucketLocation(ctx, bucketName)
		if err != nil {
			return fmt.Errorf("failed to get bucket location: %w\nPlease ensure you have access to bucket '%s'", err, bucketName)
		}

		bucketRegion = locationResp
		fmt.Printf("✓ Bucket '%s' is in region: %s\n", bucketName, bucketRegion)
	} else {
		// Create new bucket
		var newBucketName string
		newBucketForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("New S3 Bucket Name").
					Description(fmt.Sprintf("Bucket for Terraform state in %s (must be globally unique)", bucketRegion)).
					Value(&newBucketName).
					Placeholder("my-terraform-state-bucket").
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("bucket name is required")
						}
						if len(s) < 3 || len(s) > 63 {
							return fmt.Errorf("bucket name must be 3-63 characters")
						}
						return nil
					}),
			),
		)

		if err := newBucketForm.Run(); err != nil {
			return err
		}

		bucketName = newBucketName

		// Check if bucket exists
		s3Manager, err := backend.NewS3Manager(ctx, bucketRegion)
		if err != nil {
			fmt.Printf("\n⚠️  Warning: Could not connect to S3: %v\n", err)
			cfg.Terraform.Backend.S3Bucket = bucketName
			cfg.Terraform.Backend.S3Region = bucketRegion
			return nil
		}

		exists, err := s3Manager.BucketExists(ctx, bucketName)
		if err != nil {
			fmt.Printf("\n⚠️  Warning: Could not check bucket: %v\n", err)
			cfg.Terraform.Backend.S3Bucket = bucketName
			cfg.Terraform.Backend.S3Region = bucketRegion
			return nil
		}

		if exists {
			fmt.Printf("\n✓ Bucket '%s' already exists and will be used for state storage\n", bucketName)
		} else {
			fmt.Printf("\n📦 Bucket '%s' does not exist\n", bucketName)

			var createBucket bool
			confirmForm := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("Create S3 Bucket?").
						Description("Create the bucket with versioning, encryption, and lifecycle policies?").
						Value(&createBucket),
				),
			)

			if err := confirmForm.Run(); err != nil {
				return err
			}

			if createBucket {
				fmt.Println("\n🔨 Configuring S3 bucket with security best practices...")
				created, err := s3Manager.CreateStateBucket(ctx, bucketName)
				if err != nil {
					return fmt.Errorf("failed to configure bucket: %w", err)
				}

				if created {
					fmt.Printf("✓ Bucket '%s' created successfully with:\n", bucketName)
				} else {
					fmt.Printf("✓ Bucket '%s' already exists, configured with:\n", bucketName)
				}
				fmt.Println("  - Versioning enabled")
				fmt.Println("  - Server-side encryption (AES256)")
				fmt.Println("  - Public access blocked")
				fmt.Println("  - Lifecycle policy (7-day lock file retention)")
			}
		}
	}

	// Set the backend configuration
	cfg.Terraform.Backend.S3Bucket = bucketName
	cfg.Terraform.Backend.S3Region = bucketRegion

	return nil
}

func checkRequirements(cfg *config.Config) error {
	fmt.Println("\n📋 Step 4: Requirements Check")
	fmt.Println()

	useDocker := cfg.LLM.Provider == providerOllama && cfg.LLM.Ollama.UseDocker
	reqs, err := requirements.CheckRequirements(cfg.LLM.Provider, useDocker)
	if err != nil {
		return err
	}

	fmt.Println("Checking system requirements:")
	for _, req := range reqs {
		fmt.Printf("  %s\n", requirements.FormatRequirementStatus(req))
	}

	missing := requirements.GetMissingRequired(reqs)
	if len(missing) > 0 {
		fmt.Println("\n⚠️  Missing required dependencies:")
		for _, name := range missing {
			fmt.Printf("  - %s\n", name)
		}
		fmt.Println("\nPlease install missing dependencies before using SCAI.")
	}

	return nil
}

func displaySummary(cfg *config.Config) {
	fmt.Println("\n" + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ Configuration Complete!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	fmt.Println("\n📋 Configuration Summary:")
	fmt.Printf("  LLM Provider: %s\n", cfg.LLM.Provider)

	switch cfg.LLM.Provider {
	case providerOllama:
		fmt.Printf("    Model: %s\n", cfg.LLM.Ollama.Model)
		if cfg.LLM.Ollama.UseDocker {
			fmt.Println("    Mode: Docker")
		} else {
			fmt.Printf("    URL: %s\n", cfg.LLM.Ollama.URL)
		}
	case providerGemini:
		fmt.Printf("    Model: %s\n", cfg.LLM.Gemini.Model)
	case providerOpenAI:
		fmt.Printf("    Model: %s\n", cfg.LLM.OpenAI.Model)
	}

	fmt.Printf("\n  Cloud Provider: %s\n", cfg.Cloud.Provider)
	fmt.Printf("    Default Region: %s\n", cfg.Cloud.DefaultRegion)

	fmt.Printf("\n  Terraform Backend:\n")
	fmt.Printf("    Type: %s\n", cfg.Terraform.Backend.Type)
	fmt.Printf("    S3 Bucket: %s\n", cfg.Terraform.Backend.S3Bucket)
	fmt.Printf("    S3 Region: %s\n", cfg.Terraform.Backend.S3Region)

	home, _ := os.UserHomeDir()
	fmt.Printf("\n📁 Configuration saved to: %s/.scai.yaml\n", home)

	fmt.Println("\n🎉 Next Steps:")
	fmt.Println("  1. Run 'scia deploy' to deploy your first application")
	fmt.Println("  2. Use --verbose flag for detailed output")
	fmt.Println("  3. Check the documentation for more examples")
	fmt.Println()
}
