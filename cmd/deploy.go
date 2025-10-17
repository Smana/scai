package cmd

import (
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
	defaultOllamaURL = "http://localhost:11434"
)

var deployCmd = &cobra.Command{
	Use:   "deploy [prompt] [repository_url_or_zip]",
	Short: "Deploy an application to AWS",
	Long: `SCIA (Smart Cloud Infrastructure Automation) analyzes code repositories,
determines optimal deployment strategies using AI, and automatically provisions
infrastructure using Terraform.

Example:
  scia deploy "Deploy this Flask app on AWS" https://github.com/Arvo-AI/hello_world
  scia deploy "Deploy microservices" /path/to/app.zip`,
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
	deployCmd.Flags().Int("ec2-volume-size", 20, "EC2 root volume size in GB")

	// Lambda sizing parameters
	deployCmd.Flags().Int("lambda-memory", 512, "Lambda memory in MB (128-10240)")
	deployCmd.Flags().Int("lambda-timeout", 30, "Lambda timeout in seconds (1-900)")
	deployCmd.Flags().Int("lambda-reserved-concurrency", 0, "Lambda reserved concurrent executions (0 = unreserved)")

	// EKS sizing parameters
	deployCmd.Flags().String("eks-node-type", "t3.medium", "EKS node instance type")
	deployCmd.Flags().Int("eks-min-nodes", 1, "EKS minimum number of nodes")
	deployCmd.Flags().Int("eks-max-nodes", 3, "EKS maximum number of nodes")
	deployCmd.Flags().Int("eks-desired-nodes", 2, "EKS desired number of nodes")
	deployCmd.Flags().Int("eks-node-volume-size", 20, "EKS node volume size in GB")
}

func runDeploy(cmd *cobra.Command, args []string) error {
	userPrompt := args[0]
	repoSource := args[1]

	// Get configuration
	verbose := viper.GetBool("verbose")

	// Ensure Ollama is available before proceeding
	ollamaURL, ollamaModel, err := ensureOllamaAvailable(verbose)
	if err != nil {
		return err
	}

	// Initialize LLM client for config parsing
	llmClient := llm.NewClient(ollamaURL, ollamaModel)

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
	awsRegion := viper.GetString("aws.region")
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

	deployConfig := planConfig

	d := deployer.NewDeployer(deployConfig)
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

// ensureOllamaAvailable ensures Ollama is running and accessible
func ensureOllamaAvailable(verbose bool) (url string, model string, err error) {
	model = viper.GetString("ollama.model")
	configuredURL := viper.GetString("ollama.url")
	useDocker := viper.GetBool("ollama.use_docker")

	// Priority 1: Check if remote/configured URL is accessible
	if configuredURL != defaultOllamaURL {
		if verbose {
			fmt.Printf("🔍 Checking remote Ollama at %s...\n", configuredURL)
		}
		if llm.IsOllamaAccessible(configuredURL) {
			if verbose {
				fmt.Printf("✓ Connected to remote Ollama\n\n")
			}
			return configuredURL, model, nil
		}
		return "", "", fmt.Errorf(`❌ Ollama not available at configured URL: %s

Please ensure Ollama is running at the remote URL or update your configuration:
  export SCIA_OLLAMA_URL=http://your-server:11434

Or remove the configuration to use Docker.`, configuredURL)
	}

	// Priority 2: Try Docker (if enabled)
	if useDocker && llm.IsDockerAvailable() {
		if verbose {
			fmt.Println("🐳 Checking Docker Ollama...")
		}

		url, err = llm.SetupOllamaDocker(model, verbose)
		if err == nil {
			if verbose {
				fmt.Println()
			}
			return url, model, nil
		}

		if verbose {
			fmt.Printf("Warning: Docker setup failed: %v\n", err)
		}
	}

	// Priority 3: Try localhost
	if verbose {
		fmt.Println("🔍 Checking local Ollama...")
	}
	if llm.IsOllamaAccessible(defaultOllamaURL) {
		if verbose {
			fmt.Printf("✓ Connected to local Ollama\n\n")
		}
		return defaultOllamaURL, model, nil
	}

	// All options failed - return helpful error
	return "", "", fmt.Errorf(`❌ Ollama LLM is not available!

SCIA requires Ollama for natural language parsing and deployment decisions.

Options to fix:

1. 🐳 Use Docker Ollama (recommended):
   docker run -d --name scia-ollama -p 11434:11434 -v ollama-data:/root/.ollama ollama/ollama
   docker exec scia-ollama ollama pull %s

2. 🌐 Use remote Ollama server:
   export SCIA_OLLAMA_URL=http://remote-server:11434

3. 💻 Start Ollama locally:
   ollama serve
   ollama pull %s

After starting Ollama, run your command again.`, model, model)
}
