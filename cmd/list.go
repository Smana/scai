package cmd

import (
	"context"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/Smana/scai/internal/store"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all deployments",
	Long: `List all deployments with optional filtering by region, strategy, status, or app name.

Example:
  scia list
  scia list --region us-east-1
  scia list --strategy vm
  scia list --status succeeded
  scia list --app hello-world`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	// List-specific flags
	listCmd.Flags().String("region", "", "Filter by AWS region")
	listCmd.Flags().String("strategy", "", "Filter by deployment strategy (vm, kubernetes, serverless)")
	listCmd.Flags().String("status", "", "Filter by deployment status (pending, running, succeeded, failed, destroyed)")
	listCmd.Flags().String("app", "", "Filter by application name")
}

func runList(cmd *cobra.Command, args []string) error {
	if globalStore == nil {
		return fmt.Errorf("database not initialized")
	}

	ctx := context.Background()

	// Build filter from flags
	filter := &store.DeploymentFilter{}

	if region, _ := cmd.Flags().GetString("region"); region != "" {
		filter.Region = region
	}
	if strategy, _ := cmd.Flags().GetString("strategy"); strategy != "" {
		filter.Strategy = strategy
	}
	if status, _ := cmd.Flags().GetString("status"); status != "" {
		filter.Status = store.DeploymentStatus(status)
	}
	if app, _ := cmd.Flags().GetString("app"); app != "" {
		filter.AppName = app
	}

	// Query deployments
	deployments, err := globalStore.List(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	// Display results
	if len(deployments) == 0 {
		pterm.Info.Println("No deployments found.")
		return nil
	}

	pterm.DefaultHeader.WithFullWidth().Printf("Found %d deployment(s)", len(deployments))
	pterm.Println()

	// Prepare table data
	tableData := pterm.TableData{
		{"ID", "APP NAME", "STRATEGY", "REGION", "STATUS", "CREATED"},
	}

	for _, dep := range deployments {
		// Format creation time
		createdTime := dep.CreatedAt.Format("2006-01-02 15:04")

		// Truncate app name if too long
		appName := dep.AppName
		if len(appName) > 20 {
			appName = appName[:17] + "..."
		}

		// Add status indicator
		statusIcon := getStatusIcon(dep.Status)

		tableData = append(tableData, []string{
			dep.ID,
			appName,
			dep.Strategy,
			dep.Region,
			fmt.Sprintf("%s %s", statusIcon, dep.Status),
			createdTime,
		})
	}

	// Render table
	if err := pterm.DefaultTable.WithHasHeader().WithData(tableData).Render(); err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	pterm.Println()
	pterm.Info.Println("Use 'scia show <deployment-id>' to see detailed information")

	return nil
}

// getStatusIcon returns an emoji icon for the deployment status
func getStatusIcon(status store.DeploymentStatus) string {
	switch status {
	case store.DeploymentStatusPending:
		return "⏳"
	case store.DeploymentStatusRunning:
		return "🔄"
	case store.DeploymentStatusSucceeded:
		return "✅"
	case store.DeploymentStatusFailed:
		return "❌"
	case store.DeploymentStatusDestroyed:
		return "🗑️"
	default:
		return "❓"
	}
}
