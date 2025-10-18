package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var outputsCmd = &cobra.Command{
	Use:   "outputs <deployment-id>",
	Short: "Show deployment outputs",
	Long: `Display Terraform outputs for a specific deployment, such as URLs, IPs, and other resource information.

Example:
  scia outputs abc123de-f456-7890-abcd-ef1234567890
  scia outputs abc123de --json`,
	Args: cobra.ExactArgs(1),
	RunE: runOutputs,
}

func init() {
	rootCmd.AddCommand(outputsCmd)

	// Outputs-specific flags
	outputsCmd.Flags().Bool("json", false, "Output as JSON")
}

func runOutputs(cmd *cobra.Command, args []string) error {
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

	// Check if deployment has outputs
	if len(deployment.Outputs) == 0 {
		pterm.Info.Printf("No outputs found for deployment %s\n", deploymentID)
		return nil
	}

	// Check if JSON output requested
	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		// Output as JSON
		data, err := json.MarshalIndent(deployment.Outputs, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Display human-readable format
	pterm.Println()
	pterm.DefaultHeader.WithFullWidth().Printf("Outputs: %s", deployment.AppName)
	pterm.Println()

	// Find longest key for alignment
	maxKeyLen := 0
	for key := range deployment.Outputs {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}
	}

	// Display outputs
	for key, value := range deployment.Outputs {
		pterm.Printf("  %-*s = %s\n", maxKeyLen, key, value)
	}
	pterm.Println()

	return nil
}
