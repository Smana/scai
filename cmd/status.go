package cmd

import (
	"context"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <deployment-id>",
	Short: "Check deployment status",
	Long: `Check the current status of a deployment.

Status values:
  - pending: Deployment created but not started
  - running: Deployment in progress
  - succeeded: Deployment completed successfully
  - failed: Deployment failed
  - destroyed: Deployment has been destroyed

Example:
  scia status abc123de-f456-7890-abcd-ef1234567890`,
	Args: cobra.ExactArgs(1),
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	if globalStore == nil {
		return fmt.Errorf("database not initialized")
	}

	ctx := context.Background()
	deploymentID := args[0]

	// Get deployment
	deployment, err := globalStore.Get(ctx, deploymentID)
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	// Display status
	pterm.Println()
	pterm.DefaultHeader.WithFullWidth().Printf("Status: %s", deployment.AppName)
	pterm.Println()

	pterm.Printf("Deployment: %s\n", deployment.AppName)
	pterm.Printf("ID:         %s\n", deployment.ID)
	pterm.Printf("Status:     %s %s\n", getStatusIcon(deployment.Status), deployment.Status)
	pterm.Printf("Strategy:   %s\n", deployment.Strategy)
	pterm.Printf("Region:     %s\n", deployment.Region)
	pterm.Println()

	// Display timestamps
	pterm.Printf("Created:    %s\n", deployment.CreatedAt.Format("2006-01-02 15:04:05 MST"))
	if deployment.DeployedAt != nil {
		pterm.Printf("Deployed:   %s\n", deployment.DeployedAt.Format("2006-01-02 15:04:05 MST"))
	}
	if deployment.DestroyedAt != nil {
		pterm.Printf("Destroyed:  %s\n", deployment.DestroyedAt.Format("2006-01-02 15:04:05 MST"))
	}

	// Display error if failed
	if deployment.ErrorMessage != "" {
		pterm.Println()
		pterm.Error.Println("Error:")
		pterm.Printf("  %s\n", deployment.ErrorMessage)
	}

	pterm.Println()

	return nil
}
