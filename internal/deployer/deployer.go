package deployer

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"

	"github.com/Smana/scia/internal/backend"
	"github.com/Smana/scia/internal/llm"
	"github.com/Smana/scia/internal/store"
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
	store     store.Store
}

// NewDeployer creates a new Deployer instance
func NewDeployer(config *DeployConfig, storeInstance store.Store) *Deployer {
	return &Deployer{
		config: config,
		store:  storeInstance,
	}
}

// SetLLMClient sets the LLM client for the deployer
func (d *Deployer) SetLLMClient(client *llm.Client) {
	d.llmClient = client
}

// Deploy executes the deployment workflow
func (d *Deployer) Deploy() (*types.DeploymentResult, error) {
	ctx := context.Background()

	// Generate unique deployment ID
	deploymentID := uuid.New().String()

	// Create deployment record with status "running"
	deployment := &store.Deployment{
		ID:                deploymentID,
		AppName:           d.extractAppName(),
		UserPrompt:        d.config.UserPrompt,
		RepoURL:           d.config.Analysis.RepoURL,
		RepoCommitSHA:     d.config.Analysis.CommitSHA,
		Strategy:          d.config.Strategy,
		Region:            d.config.AWSRegion,
		Status:            store.DeploymentStatusRunning,
		TerraformStateKey: fmt.Sprintf("deployments/%s/terraform.tfstate", deploymentID),
		TerraformDir:      "",
		Analysis:          d.config.Analysis,
		Config:            nil,
		Outputs:           make(map[string]string),
		Warnings:          []string{},
		Optimizations:     []string{},
		ErrorMessage:      "",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		DeployedAt:        nil,
		DestroyedAt:       nil,
	}

	if d.store != nil {
		if err := d.store.Create(ctx, deployment); err != nil {
			return nil, fmt.Errorf("failed to create deployment record: %w", err)
		}

		if d.config.Verbose {
			fmt.Printf("   Created deployment record: %s\n", deploymentID)
		}
	}

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
		AppDir:       d.config.Analysis.AppDir,
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
		// Update deployment status to failed
		if d.store != nil {
			_ = d.store.UpdateStatus(ctx, deploymentID, store.DeploymentStatusFailed, err.Error())
		}
		return nil, fmt.Errorf("failed to generate Terraform config: %w", err)
	}

	// Update deployment record with config and terraform directory
	deployment.Config = tfConfig
	deployment.TerraformDir = tfDir
	if d.store != nil {
		if err := d.store.Update(ctx, deployment); err != nil {
			return nil, fmt.Errorf("failed to update deployment record: %w", err)
		}
	}

	// Generate backend.tf for S3 state storage (if configured)
	if err := d.generateBackend(tfDir, deployment.TerraformStateKey); err != nil {
		// Update deployment status to failed
		if d.store != nil {
			_ = d.store.UpdateStatus(ctx, deploymentID, store.DeploymentStatusFailed, err.Error())
		}
		return nil, fmt.Errorf("failed to generate backend configuration: %w", err)
	}

	// Execute Terraform
	if d.config.Verbose {
		fmt.Printf("   Running Terraform...\n")
	}

	executor, err := terraform.NewExecutor(tfDir, d.config.TerraformBin, d.config.Verbose)
	if err != nil {
		// Update deployment status to failed
		if d.store != nil {
			_ = d.store.UpdateStatus(ctx, deploymentID, store.DeploymentStatusFailed, err.Error())
		}
		return nil, fmt.Errorf("failed to create terraform executor: %w", err)
	}

	if err := executor.Init(); err != nil {
		// Update deployment status to failed
		if d.store != nil {
			_ = d.store.UpdateStatus(ctx, deploymentID, store.DeploymentStatusFailed, fmt.Sprintf("terraform init failed: %v", err))
		}
		return nil, fmt.Errorf("terraform init failed: %w", err)
	}

	if err := executor.Apply(); err != nil {
		// Update deployment status to failed
		if d.store != nil {
			_ = d.store.UpdateStatus(ctx, deploymentID, store.DeploymentStatusFailed, fmt.Sprintf("terraform apply failed: %v", err))
		}
		return nil, fmt.Errorf("terraform apply failed: %w", err)
	}

	// Get outputs
	outputs, err := executor.Outputs()
	if err != nil {
		// Update deployment status to failed
		if d.store != nil {
			_ = d.store.UpdateStatus(ctx, deploymentID, store.DeploymentStatusFailed, fmt.Sprintf("failed to get outputs: %v", err))
		}
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

	// Update deployment record with success status and outputs
	deployment.Outputs = outputs
	deployment.Warnings = result.Warnings
	deployment.Optimizations = result.Optimizations
	if d.store != nil {
		if err := d.store.UpdateStatus(ctx, deploymentID, store.DeploymentStatusSucceeded, ""); err != nil {
			// Log but don't fail deployment
			if d.config.Verbose {
				fmt.Printf("   Warning: failed to update deployment status: %v\n", err)
			}
		}

		// Update full deployment record
		if err := d.store.Update(ctx, deployment); err != nil {
			// Log but don't fail deployment
			if d.config.Verbose {
				fmt.Printf("   Warning: failed to update deployment record: %v\n", err)
			}
		}

		if d.config.Verbose {
			fmt.Printf("   ✓ Deployment completed successfully: %s\n", deploymentID)
		}
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

// generateBackend generates the backend.tf file for S3 state storage
func (d *Deployer) generateBackend(tfDir string, deploymentStateKey string) error {
	// Read backend configuration from viper
	backendType := viper.GetString("terraform.backend.type")

	// Only generate backend.tf if S3 backend is configured
	if backendType != "s3" {
		if d.config.Verbose {
			fmt.Printf("   No S3 backend configured, using local state\n")
		}
		return nil
	}

	s3Bucket := viper.GetString("terraform.backend.s3_bucket")
	s3Region := viper.GetString("terraform.backend.s3_region")

	// Validate required fields
	if s3Bucket == "" || s3Region == "" {
		if d.config.Verbose {
			fmt.Printf("   S3 backend not fully configured, using local state\n")
		}
		return nil
	}

	// Use deployment-specific S3 key (e.g., deployments/<uuid>/terraform.tfstate)
	s3Key := deploymentStateKey

	if d.config.Verbose {
		fmt.Printf("   Configuring S3 backend: bucket=%s, region=%s, key=%s\n",
			s3Bucket, s3Region, s3Key)
	}

	// Generate backend.tf
	backendCfg := backend.BackendTFConfig{
		BucketName: s3Bucket,
		Region:     s3Region,
		Key:        s3Key,
	}

	backendFile, err := backend.WriteBackendTF(tfDir, backendCfg)
	if err != nil {
		return err
	}

	if d.config.Verbose {
		fmt.Printf("   ✓ Generated backend.tf at %s\n", backendFile)
	}

	return nil
}
