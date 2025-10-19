package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Smana/scia/internal/analyzer"
	"github.com/Smana/scia/internal/deployer"
	"github.com/Smana/scia/internal/llm"
	"github.com/Smana/scia/internal/parser"
	"github.com/Smana/scia/internal/ui"
)

const (
	defaultOllamaURL   = "http://localhost:11434"
	providerTypeOllama = "ollama"
	providerTypeGemini = "gemini"
	providerTypeOpenAI = "openai"
	defaultAWSRegion   = "eu-west-3"
)

var deployCmd = &cobra.Command{
	Use:   "deploy [prompt] [repository_url_or_zip]",
	Short: "Deploy an application to AWS",
	Long: `SCIA (Smart Cloud Infrastructure Automation) analyzes code repositories,
determines optimal deployment strategies using AI, and automatically provisions
infrastructure using Terraform.

Example:
  scai deploy "Deploy this Flask app on AWS" https://github.com/user/flask-app
  scai deploy "Deploy microservices" /path/to/app.zip`,
	Args: cobra.ExactArgs(2),
	RunE: runDeploy,
}

func init() {
	rootCmd.AddCommand(deployCmd)

	// Deploy-specific flags
	deployCmd.Flags().String("strategy", "", "Force deployment strategy (vm, kubernetes, serverless)")
	deployCmd.Flags().String("region", "", "AWS region (overrides config)")
	deployCmd.Flags().BoolP("yes", "y", false, "Auto-approve deployment without confirmation prompt")

	// EC2 sizing parameters
	deployCmd.Flags().String("ec2-instance-type", "", "EC2 instance type (default: t3.micro)")
	deployCmd.Flags().Int("ec2-volume-size", 30, "EC2 root volume size in GB")

	// Lambda sizing parameters
	deployCmd.Flags().Int("lambda-memory", 512, "Lambda memory in MB (128-10240)")
	deployCmd.Flags().Int("lambda-timeout", 30, "Lambda timeout in seconds (1-900)")
	deployCmd.Flags().Int("lambda-reserved-concurrency", 0, "Lambda reserved concurrent executions (0 = unreserved)")

	// EKS sizing parameters
	deployCmd.Flags().String("eks-node-type", "t3.medium", "EKS node instance type")
	deployCmd.Flags().Int("eks-min-nodes", 1, "EKS minimum number of nodes")
	deployCmd.Flags().Int("eks-max-nodes", 3, "EKS maximum number of nodes")
	deployCmd.Flags().Int("eks-desired-nodes", 2, "EKS desired number of nodes")
	deployCmd.Flags().Int("eks-node-volume-size", 30, "EKS node volume size in GB")
}

func runDeploy(cmd *cobra.Command, args []string) error {
	userPrompt := args[0]
	repoSource := args[1]

	// Get configuration
	verbose := viper.GetBool("verbose")

	// Initialize LLM provider
	providerManager, providerConfig, err := initializeLLMProvider(verbose)
	if err != nil {
		return err
	}

	// Create LLM client from the configured provider manager
	llmClient := llm.NewClientWithManager(providerManager, providerConfig)

	// Parse natural language prompt for configuration using LLM
	var parsedConfig *parser.DeploymentConfig
	parsedConfig, err = parser.ParseConfigFromPrompt(llmClient, userPrompt)
	if err != nil && verbose {
		fmt.Printf("Warning: Could not parse prompt configuration: %v\n", err)
	}

	if verbose && parsedConfig != nil {
		fmt.Println("🔍 Detected configuration from prompt:")
		if parsedConfig.Strategy != "" {
			fmt.Printf("   Strategy: %s\n", parsedConfig.Strategy)
		}
		if parsedConfig.Region != "" {
			fmt.Printf("   Region: %s\n", parsedConfig.Region)
		}
		if parsedConfig.EC2InstanceType != "" {
			fmt.Printf("   EC2 Instance: %s\n", parsedConfig.EC2InstanceType)
		}
		if parsedConfig.EKSNodeType != "" {
			fmt.Printf("   EKS Node Type: %s\n", parsedConfig.EKSNodeType)
		}
		if parsedConfig.EKSDesiredNodes > 0 {
			fmt.Printf("   EKS Nodes: %d (min: %d, max: %d)\n", parsedConfig.EKSDesiredNodes, parsedConfig.EKSMinNodes, parsedConfig.EKSMaxNodes)
		}
		fmt.Println()
	}

	// Get remaining configuration
	workDir := viper.GetString("workdir")
	awsRegion := viper.GetString("cloud.default_region")
	tfBin := viper.GetString("terraform.bin")

	// Override with parsed config (natural language takes precedence)
	if parsedConfig.Region != "" {
		awsRegion = parsedConfig.Region
	}

	// Override region if flag provided (flags have highest priority)
	if region, _ := cmd.Flags().GetString("region"); region != "" {
		awsRegion = region
	}

	if verbose {
		fmt.Printf("🚀 SCIA Deployment Starting...\n")
		fmt.Printf("   User Prompt: %s\n", userPrompt)
		fmt.Printf("   Repository: %s\n", repoSource)
		fmt.Printf("   Work Directory: %s\n", workDir)
		fmt.Printf("   AWS Region: %s\n", awsRegion)
		fmt.Printf("   Terraform Binary: %s\n", tfBin)
		fmt.Println()
	}

	// Create work directory
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}

	// Step 1: Analyze repository
	fmt.Println("📊 Analyzing repository...")
	analyzer := analyzer.NewAnalyzer(workDir, verbose)
	analysis, err := analyzer.Analyze(repoSource)
	if err != nil {
		return fmt.Errorf("repository analysis failed: %w", err)
	}

	if verbose {
		fmt.Printf("   Framework: %s\n", analysis.Framework)
		fmt.Printf("   Language: %s\n", analysis.Language)
		fmt.Printf("   Port: %d\n", analysis.Port)
		fmt.Printf("   Dependencies: %d\n", len(analysis.Dependencies))
		fmt.Printf("   Docker: %v\n", analysis.HasDockerfile)
		fmt.Println()
	}

	// Step 2: Determine deployment strategy
	fmt.Println("🤖 Determining deployment strategy...")

	var strategy string
	forcedStrategy, _ := cmd.Flags().GetString("strategy")

	// Check if strategy was specified in natural language
	if parsedConfig != nil && parsedConfig.Strategy != "" && forcedStrategy == "" {
		strategy = parsedConfig.Strategy
		fmt.Printf("   Strategy from prompt: %s\n", strategy)
	} else if forcedStrategy != "" {
		strategy = forcedStrategy
		fmt.Printf("   Using forced strategy: %s\n", strategy)
	} else {
		// Use LLM client to determine strategy based on code analysis
		strategy, err = llmClient.DetermineStrategy(parsedConfig.CleanedPrompt, analysis)
		if err != nil {
			return fmt.Errorf("failed to determine strategy: %w", err)
		}
		fmt.Printf("   Recommended strategy: %s\n", strategy)
	}
	fmt.Println()

	// Extract app name for deployment plan
	appName := extractAppName(repoSource)

	// Step 2.5: Build deployment plan and get confirmation
	fmt.Println("📋 Preparing deployment plan...")
	fmt.Println()

	// Extract sizing parameters from flags
	ec2InstanceType, _ := cmd.Flags().GetString("ec2-instance-type")
	ec2VolumeSize, _ := cmd.Flags().GetInt("ec2-volume-size")
	lambdaMemory, _ := cmd.Flags().GetInt("lambda-memory")
	lambdaTimeout, _ := cmd.Flags().GetInt("lambda-timeout")
	lambdaReservedConcurrency, _ := cmd.Flags().GetInt("lambda-reserved-concurrency")
	eksNodeType, _ := cmd.Flags().GetString("eks-node-type")
	eksMinNodes, _ := cmd.Flags().GetInt("eks-min-nodes")
	eksMaxNodes, _ := cmd.Flags().GetInt("eks-max-nodes")
	eksDesiredNodes, _ := cmd.Flags().GetInt("eks-desired-nodes")
	eksNodeVolumeSize, _ := cmd.Flags().GetInt("eks-node-volume-size")

	// Apply parsed config from natural language (if not overridden by flags)
	if parsedConfig != nil {
		if ec2InstanceType == "" && parsedConfig.EC2InstanceType != "" {
			ec2InstanceType = parsedConfig.EC2InstanceType
		}
		if parsedConfig.EC2VolumeSize > 0 {
			ec2VolumeSize = parsedConfig.EC2VolumeSize
		}
		if eksNodeType == "t3.medium" && parsedConfig.EKSNodeType != "" {
			eksNodeType = parsedConfig.EKSNodeType
		}
		if parsedConfig.EKSMinNodes > 0 {
			eksMinNodes = parsedConfig.EKSMinNodes
		}
		if parsedConfig.EKSMaxNodes > 0 {
			eksMaxNodes = parsedConfig.EKSMaxNodes
		}
		if parsedConfig.EKSDesiredNodes > 0 {
			eksDesiredNodes = parsedConfig.EKSDesiredNodes
		}
		if parsedConfig.LambdaMemory > 0 {
			lambdaMemory = parsedConfig.LambdaMemory
		}
		if parsedConfig.LambdaTimeout > 0 {
			lambdaTimeout = parsedConfig.LambdaTimeout
		}
	}

	// Create temporary config for plan building
	planConfig := &deployer.DeployConfig{
		Strategy:                  strategy,
		Analysis:                  analysis,
		AWSRegion:                 awsRegion,
		EC2InstanceType:           ec2InstanceType,
		EC2VolumeSize:             ec2VolumeSize,
		LambdaMemory:              lambdaMemory,
		LambdaTimeout:             lambdaTimeout,
		LambdaReservedConcurrency: lambdaReservedConcurrency,
		EKSNodeType:               eksNodeType,
		EKSMinNodes:               eksMinNodes,
		EKSMaxNodes:               eksMaxNodes,
		EKSDesiredNodes:           eksDesiredNodes,
		EKSNodeVolumeSize:         eksNodeVolumeSize,
	}

	// Build deployment plan
	plan := ui.BuildDeploymentPlan(strategy, awsRegion, appName, analysis, planConfig)

	// Get --yes flag
	autoApprove, _ := cmd.Flags().GetBool("yes")

	// Show plan and get confirmation (with interactive modification support)
	confirmed, updatedConfig, err := ui.ConfirmOrModify(plan, analysis, planConfig, llmClient, autoApprove)
	if err != nil {
		return fmt.Errorf("deployment confirmation failed: %w", err)
	}

	if !confirmed {
		fmt.Println()
		fmt.Println("❌ Deployment canceled by user")
		return nil
	}

	// Use updated config from modification loop
	planConfig = updatedConfig

	fmt.Println()

	// Step 3: Deploy infrastructure (extend planConfig)
	planConfig.UserPrompt = userPrompt
	planConfig.WorkDir = workDir
	planConfig.TerraformBin = tfBin
	planConfig.Verbose = verbose
	planConfig.LLMProvider = providerConfig.Type
	planConfig.LLMModel = getLLMModel(providerConfig)

	deployConfig := planConfig

	d := deployer.NewDeployer(deployConfig, globalStore)
	d.SetLLMClient(llmClient)
	result, err := d.Deploy()
	if err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	// Step 4: Display results
	fmt.Println()
	fmt.Println("✅ Deployment Complete!")
	fmt.Println()
	fmt.Println("📋 Deployment Summary:")
	fmt.Printf("   Strategy: %s\n", result.Strategy)
	fmt.Printf("   Region: %s\n", result.Region)

	if len(result.Outputs) > 0 {
		fmt.Println()
		fmt.Println("🔗 Access URLs:")
		for key, value := range result.Outputs {
			fmt.Printf("   %s: %s\n", key, value)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Println()
		fmt.Println("⚠️  Warnings:")
		for _, warning := range result.Warnings {
			fmt.Printf("   %s\n", warning)
		}
	}

	if len(result.Optimizations) > 0 {
		fmt.Println()
		fmt.Println("💡 Optimization Suggestions:")
		for _, opt := range result.Optimizations {
			fmt.Printf("   %s\n", opt)
		}
	}

	if result.TerraformDir != "" {
		fmt.Println()
		fmt.Printf("📁 Terraform files: %s\n", result.TerraformDir)
	}

	fmt.Println()
	fmt.Println("🎉 Success! Your application is now deployed.")

	return nil
}

// extractAppName extracts application name from repository URL or path
func extractAppName(repoSource string) string {
	// Remove .git suffix if present
	if len(repoSource) > 4 && repoSource[len(repoSource)-4:] == ".git" {
		repoSource = repoSource[:len(repoSource)-4]
	}

	// Remove trailing slash if present
	repoSource = strings.TrimSuffix(repoSource, "/")

	// Get last path component
	for i := len(repoSource) - 1; i >= 0; i-- {
		if repoSource[i] == '/' {
			name := repoSource[i+1:]
			// Replace underscores with hyphens for Kubernetes compatibility
			return strings.ReplaceAll(name, "_", "-")
		}
	}

	// Fallback
	return "scia-app"
}

// initializeLLMProvider initializes the LLM provider based on configuration
// Returns the ProviderManager and its config for creating a Client
func initializeLLMProvider(verbose bool) (*llm.ProviderManager, *llm.ProviderConfig, error) {
	ctx := context.Background()

	// Get provider type from config
	providerType := viper.GetString("llm.provider")
	if providerType == "" {
		providerType = providerTypeOllama // Default to ollama for backward compatibility
	}

	// Build provider config
	providerConfig := &llm.ProviderConfig{
		Type: providerType,

		// Ollama configuration
		OllamaURL:   viper.GetString("llm.ollama.url"),
		OllamaModel: viper.GetString("llm.ollama.model"),

		// Gemini configuration
		GeminiAPIKey: viper.GetString("llm.gemini.api_key"),
		GeminiModel:  viper.GetString("llm.gemini.model"),

		// OpenAI configuration
		OpenAIAPIKey: viper.GetString("llm.openai.api_key"),
		OpenAIModel:  viper.GetString("llm.openai.model"),
	}

	// Special handling for Ollama - ensure it's available
	if providerType == providerTypeOllama {
		useDocker := viper.GetBool("llm.ollama.use_docker")
		configuredURL := providerConfig.OllamaURL

		// Priority 1: Check if remote/configured URL is accessible
		if configuredURL != defaultOllamaURL {
			if verbose {
				fmt.Printf("🔍 Checking remote Ollama at %s...\n", configuredURL)
			}
			if llm.IsOllamaAccessible(configuredURL) {
				if verbose {
					fmt.Printf("✓ Connected to remote Ollama\n\n")
				}
			} else {
				return nil, nil, fmt.Errorf(`❌ Ollama not available at configured URL: %s

Please ensure Ollama is running or update your configuration with 'scia init'`, configuredURL)
			}
		} else {
			// Priority 2: Try Docker (if enabled)
			if useDocker && llm.IsDockerAvailable() {
				if verbose {
					fmt.Println("🐳 Checking Docker Ollama...")
				}

				url, err := llm.SetupOllamaDocker(providerConfig.OllamaModel, verbose)
				if err == nil {
					providerConfig.OllamaURL = url
					if verbose {
						fmt.Println()
					}
				} else if verbose {
					fmt.Printf("Warning: Docker setup failed: %v\n", err)
				}
			} else if llm.IsOllamaAccessible(defaultOllamaURL) {
				// Priority 3: Try localhost
				if verbose {
					fmt.Println("🔍 Checking local Ollama...")
					fmt.Printf("✓ Connected to local Ollama\n\n")
				}
			} else {
				return nil, nil, fmt.Errorf(`❌ Ollama LLM is not available!

Run 'scia init' to configure an LLM provider, or start Ollama:
  docker run -d --name scia-ollama -p 11434:11434 -v ollama-data:/root/.ollama ollama/ollama
  docker exec scia-ollama ollama pull %s`, providerConfig.OllamaModel)
			}
		}
	}

	// Create provider manager
	providerManager, err := llm.NewProviderManager(providerConfig, verbose)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize LLM provider: %w", err)
	}

	// Check if provider is available
	bestProvider := providerManager.GetBestProvider(ctx)
	if bestProvider == nil {
		return nil, nil, fmt.Errorf(`❌ LLM provider '%s' is not available!

No accessible LLM providers found. Run 'scia init' to configure a provider.`, providerType)
	}

	if !bestProvider.IsAvailable(ctx) {
		return nil, nil, fmt.Errorf(`❌ LLM provider '%s' is not available!

Run 'scia init' to configure a different LLM provider.`, providerType)
	}

	if verbose {
		fmt.Printf("✓ Using LLM provider: %s\n\n", providerType)
	}

	return providerManager, providerConfig, nil
}

// getLLMModel returns the active model name based on provider type
func getLLMModel(config *llm.ProviderConfig) string {
	switch config.Type {
	case providerTypeOllama:
		return config.OllamaModel
	case providerTypeGemini:
		return config.GeminiModel
	case providerTypeOpenAI:
		return config.OpenAIModel
	default:
		return ""
	}
}
