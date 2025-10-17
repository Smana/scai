package deployer

import (
	"fmt"
	"path/filepath"

	"github.com/Smana/scia/internal/llm"
	"github.com/Smana/scia/internal/terraform"
	"github.com/Smana/scia/internal/types"
)

// DeployConfig contains deployment configuration
type DeployConfig struct {
	Strategy     string
	Analysis     *types.Analysis
	UserPrompt   string
	WorkDir      string
	AWSRegion    string
	TerraformBin string
	Verbose      bool

	// EC2 sizing
	EC2InstanceType string
	EC2VolumeSize   int

	// Lambda sizing
	LambdaMemory              int
	LambdaTimeout             int
	LambdaReservedConcurrency int

	// EKS sizing
	EKSNodeType       string
	EKSMinNodes       int
	EKSMaxNodes       int
	EKSDesiredNodes   int
	EKSNodeVolumeSize int
}

// Deployer orchestrates the deployment process
type Deployer struct {
	config    *DeployConfig
	llmClient *llm.Client
}

// NewDeployer creates a new Deployer instance
func NewDeployer(config *DeployConfig) *Deployer {
	return &Deployer{
		config: config,
	}
}

// Deploy executes the deployment workflow
func (d *Deployer) Deploy() (*types.DeploymentResult, error) {
	// Create terraform directory
	tfDir := filepath.Join(d.config.WorkDir, "terraform")

	if d.config.Verbose {
		fmt.Printf("   Creating Terraform configuration...\n")
	}

	// Generate Terraform configuration based on strategy
	generator := terraform.NewGenerator(tfDir, d.config.Verbose)

	tfConfig := &types.TerraformConfig{
		Strategy:     d.config.Strategy,
		AppName:      d.extractAppName(),
		Region:       d.config.AWSRegion,
		Framework:    d.config.Analysis.Framework,
		Language:     d.config.Analysis.Language,
		Port:         d.config.Analysis.Port,
		RepoURL:      d.config.Analysis.RepoURL,
		StartCommand: d.config.Analysis.StartCommand,
		EnvVars:      d.config.Analysis.EnvVars,

		// EC2 sizing
		VolumeSize: d.config.EC2VolumeSize,

		// Lambda sizing
		LambdaMemory:              d.config.LambdaMemory,
		LambdaTimeout:             d.config.LambdaTimeout,
		LambdaReservedConcurrency: d.config.LambdaReservedConcurrency,

		// EKS sizing
		EKSNodeType:       d.config.EKSNodeType,
		EKSMinNodes:       d.config.EKSMinNodes,
		EKSMaxNodes:       d.config.EKSMaxNodes,
		EKSDesiredNodes:   d.config.EKSDesiredNodes,
		EKSNodeVolumeSize: d.config.EKSNodeVolumeSize,
	}

	// Set EC2 instance type if provided or use LLM suggestion
	if d.config.EC2InstanceType != "" {
		tfConfig.InstanceType = d.config.EC2InstanceType
	} else if d.llmClient != nil {
		tfConfig.InstanceType = d.llmClient.SuggestInstanceType(d.config.Analysis)
	} else {
		tfConfig.InstanceType = "t3.micro" // Default
	}

	if err := generator.Generate(tfConfig); err != nil {
		return nil, fmt.Errorf("failed to generate Terraform config: %w", err)
	}

	// Execute Terraform
	if d.config.Verbose {
		fmt.Printf("   Running Terraform...\n")
	}

	executor, err := terraform.NewExecutor(tfDir, d.config.TerraformBin, d.config.Verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to create terraform executor: %w", err)
	}

	if err := executor.Init(); err != nil {
		return nil, fmt.Errorf("terraform init failed: %w", err)
	}

	if err := executor.Apply(); err != nil {
		return nil, fmt.Errorf("terraform apply failed: %w", err)
	}

	// Get outputs
	outputs, err := executor.Outputs()
	if err != nil {
		return nil, fmt.Errorf("failed to get terraform outputs: %w", err)
	}

	// Build deployment result
	result := &types.DeploymentResult{
		Strategy:      d.config.Strategy,
		Region:        d.config.AWSRegion,
		Outputs:       outputs,
		TerraformDir:  tfDir,
		Warnings:      []string{},
		Optimizations: []string{},
	}

	// Add warnings and optimizations if LLM client available
	if d.llmClient != nil {
		result.Warnings = d.llmClient.ValidateDeploymentRequirements(d.config.Analysis, d.config.Strategy)
		result.Optimizations = d.llmClient.SuggestOptimizations(d.config.Analysis, d.config.Strategy)
	}

	return result, nil
}

// extractAppName extracts application name from repository URL or path
func (d *Deployer) extractAppName() string {
	// Extract from repo URL: https://github.com/user/repo-name -> repo-name
	repoURL := d.config.Analysis.RepoURL

	// Simple extraction logic
	if repoURL != "" {
		// Remove .git suffix if present
		if len(repoURL) > 4 && repoURL[len(repoURL)-4:] == ".git" {
			repoURL = repoURL[:len(repoURL)-4]
		}

		// Get last path component
		for i := len(repoURL) - 1; i >= 0; i-- {
			if repoURL[i] == '/' {
				return repoURL[i+1:]
			}
		}
		return repoURL
	}

	return "scia-app"
}
