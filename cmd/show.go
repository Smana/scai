package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <deployment-id>",
	Short: "Show detailed deployment information",
	Long: `Display detailed information about a specific deployment, including configuration,
outputs, warnings, and optimizations.

Example:
  scia show abc123de-f456-7890-abcd-ef1234567890
  scia show abc123de --json`,
	Args: cobra.ExactArgs(1),
	RunE: runShow,
}

func init() {
	rootCmd.AddCommand(showCmd)

	// Show-specific flags
	showCmd.Flags().Bool("json", false, "Output as JSON")
}

func runShow(cmd *cobra.Command, args []string) error {
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

	// Check if JSON output requested
	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		// Output as JSON
		data, err := json.MarshalIndent(deployment, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Display human-readable format
	pterm.Println()
	pterm.DefaultHeader.WithFullWidth().Printf("DEPLOYMENT: %s", deployment.AppName)
	pterm.Println()

	// Basic information
	pterm.DefaultSection.Println("📋 Basic Information")
	pterm.Printf("   ID:           %s\n", deployment.ID)
	pterm.Printf("   App Name:     %s\n", deployment.AppName)
	pterm.Printf("   Status:       %s %s\n", getStatusIcon(deployment.Status), deployment.Status)
	pterm.Printf("   Strategy:     %s\n", deployment.Strategy)
	pterm.Printf("   Region:       %s\n", deployment.Region)
	pterm.Println()

	// Repository information
	pterm.DefaultSection.Println("📦 Repository")
	pterm.Printf("   URL:          %s\n", deployment.RepoURL)
	if deployment.RepoCommitSHA != "" {
		pterm.Printf("   Commit:       %s\n", deployment.RepoCommitSHA)
	}
	pterm.Println()

	// User prompt
	if deployment.UserPrompt != "" {
		pterm.DefaultSection.Println("💬 User Prompt")
		pterm.Printf("   %s\n", deployment.UserPrompt)
		pterm.Println()
	}

	// Terraform state
	pterm.DefaultSection.Println("🔧 Terraform")
	pterm.Printf("   State Key:    %s\n", deployment.TerraformStateKey)
	if deployment.TerraformDir != "" {
		pterm.Printf("   Directory:    %s\n", deployment.TerraformDir)
	}
	pterm.Println()

	// Configuration
	if deployment.Config != nil {
		pterm.DefaultSection.Println("⚙️  Configuration")
		pterm.Printf("   Framework:    %s\n", deployment.Config.Framework)
		pterm.Printf("   Language:     %s\n", deployment.Config.Language)
		pterm.Printf("   Port:         %d\n", deployment.Config.Port)
		if deployment.Config.InstanceType != "" {
			pterm.Printf("   Instance:     %s\n", deployment.Config.InstanceType)
		}
		if deployment.Config.StartCommand != "" {
			pterm.Printf("   Start Cmd:    %s\n", deployment.Config.StartCommand)
		}
		pterm.Println()
	}

	// Outputs
	if len(deployment.Outputs) > 0 {
		pterm.DefaultSection.Println("🔗 Outputs")
		for key, value := range deployment.Outputs {
			pterm.Printf("   %s: %s\n", key, value)
		}
		pterm.Println()
	}

	// Warnings
	if len(deployment.Warnings) > 0 {
		pterm.DefaultSection.Println("⚠️  Warnings")
		for _, warning := range deployment.Warnings {
			pterm.Printf("   • %s\n", warning)
		}
		pterm.Println()
	}

	// Optimizations
	if len(deployment.Optimizations) > 0 {
		pterm.DefaultSection.Println("💡 Optimization Suggestions")
		for _, opt := range deployment.Optimizations {
			pterm.Printf("   • %s\n", opt)
		}
		pterm.Println()
	}

	// Error message (if failed)
	if deployment.ErrorMessage != "" {
		pterm.DefaultSection.Println("❌ Error")
		pterm.Printf("   %s\n", deployment.ErrorMessage)
		pterm.Println()
	}

	// Timestamps
	pterm.DefaultSection.Println("🕐 Timestamps")
	pterm.Printf("   Created:      %s\n", deployment.CreatedAt.Format("2006-01-02 15:04:05 MST"))
	pterm.Printf("   Updated:      %s\n", deployment.UpdatedAt.Format("2006-01-02 15:04:05 MST"))
	if deployment.DeployedAt != nil {
		pterm.Printf("   Deployed:     %s\n", deployment.DeployedAt.Format("2006-01-02 15:04:05 MST"))
	}
	if deployment.DestroyedAt != nil {
		pterm.Printf("   Destroyed:    %s\n", deployment.DestroyedAt.Format("2006-01-02 15:04:05 MST"))
	}
	pterm.Println()

	return nil
}
