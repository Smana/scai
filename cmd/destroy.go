package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Smana/scia/internal/store"
	"github.com/Smana/scia/internal/terraform"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy <deployment-id>",
	Short: "Destroy a deployment",
	Long: `Destroy infrastructure for a specific deployment using Terraform destroy.
This will remove all AWS resources created for the deployment.

Example:
  scia destroy abc123de-f456-7890-abcd-ef1234567890
  scia destroy abc123de --yes`,
	Args: cobra.ExactArgs(1),
	RunE: runDestroy,
}

func init() {
	rootCmd.AddCommand(destroyCmd)

	// Destroy-specific flags
	destroyCmd.Flags().BoolP("yes", "y", false, "Auto-approve destroy without confirmation prompt")
}

func runDestroy(cmd *cobra.Command, args []string) error {
	if globalStore == nil {
		return fmt.Errorf("database not initialized")
	}

	ctx := context.Background()
	deploymentID := args[0]
	verbose := viper.GetBool("verbose")

	// Get deployment
	deployment, err := globalStore.Get(ctx, deploymentID)
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	// Check if already destroyed
	if deployment.Status == store.DeploymentStatusDestroyed {
		fmt.Printf("⚠️  Deployment %s is already destroyed\n", deploymentID)
		return nil
	}

	// Display deployment information
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Printf("  DESTROY DEPLOYMENT: %s\n", deployment.AppName)
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("   ID:           %s\n", deployment.ID)
	fmt.Printf("   App Name:     %s\n", deployment.AppName)
	fmt.Printf("   Strategy:     %s\n", deployment.Strategy)
	fmt.Printf("   Region:       %s\n", deployment.Region)
	fmt.Printf("   Status:       %s\n", deployment.Status)
	fmt.Println()

	// Get confirmation unless --yes flag is set
	autoApprove, _ := cmd.Flags().GetBool("yes")
	if !autoApprove {
		pterm.Warning.Println("This will destroy all infrastructure resources!")
		pterm.Println()

		response, err := pterm.DefaultInteractiveTextInput.
			WithDefaultText("Type 'yes' to confirm").
			Show()

		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		if strings.ToLower(strings.TrimSpace(response)) != "yes" {
			pterm.Info.Println("Destroy canceled")
			return nil
		}
		pterm.Println()
	} else {
		pterm.Success.Println("Auto-confirmed with --yes flag")
	}

	// Check if terraform directory exists
	if deployment.TerraformDir == "" {
		return fmt.Errorf("terraform directory not found in deployment record")
	}

	// Execute terraform destroy
	pterm.Info.Println("Destroying infrastructure...")
	if verbose {
		pterm.Debug.Printf("Terraform directory: %s\n", deployment.TerraformDir)
	}
	pterm.Info.Println("This may take several minutes...")
	pterm.Println()

	tfBin := viper.GetString("terraform.bin")
	// Always use verbose for destroy to show progress
	executor, err := terraform.NewExecutor(deployment.TerraformDir, tfBin, true)
	if err != nil {
		return fmt.Errorf("failed to create terraform executor: %w", err)
	}

	// Run terraform destroy
	if err := executor.Destroy(); err != nil {
		// Update deployment status to failed
		_ = globalStore.UpdateStatus(ctx, deploymentID, store.DeploymentStatusFailed,
			fmt.Sprintf("terraform destroy failed: %v", err))
		return fmt.Errorf("terraform destroy failed: %w", err)
	}

	// Update deployment status to destroyed
	if err := globalStore.UpdateStatus(ctx, deploymentID, store.DeploymentStatusDestroyed, ""); err != nil {
		// Log but don't fail
		if verbose {
			pterm.Warning.Printf("Failed to update deployment status: %v\n", err)
		}
	}

	pterm.Println()
	pterm.Success.Println("Deployment destroyed successfully!")
	pterm.Info.Printf("Deployment ID: %s\n", deploymentID)
	pterm.Println()

	return nil
}
