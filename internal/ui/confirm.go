package ui

import (
	"fmt"

	"github.com/pterm/pterm"
)

// ConfirmDeployment displays the deployment plan and prompts for confirmation
// Returns: confirmed (bool), error
func ConfirmDeployment(plan *DeploymentPlan, skipConfirmation bool) (bool, error) {
	// Display the plan
	if err := DisplayPlanTable(plan); err != nil {
		return false, fmt.Errorf("failed to display plan: %w", err)
	}

	// Skip confirmation if --yes flag is set
	if skipConfirmation {
		pterm.Success.Println("Auto-confirmed with --yes flag")
		return true, nil
	}

	pterm.Println()

	// Confirmation prompt
	result, err := pterm.DefaultInteractiveConfirm.
		WithDefaultText("Do you want to proceed with this deployment?").
		WithDefaultValue(false).
		WithConfirmText("Yes").
		WithRejectText("No").
		Show()

	if err != nil {
		return false, fmt.Errorf("confirmation prompt failed: %w", err)
	}

	return result, nil
}

// DisplayPlanTable renders a beautiful table showing the deployment plan
func DisplayPlanTable(plan *DeploymentPlan) error {
	// Display header
	pterm.DefaultHeader.WithFullWidth().
		WithBackgroundStyle(pterm.NewStyle(pterm.BgBlue)).
		WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).
		Println("📋 DEPLOYMENT PLAN")

	pterm.Println()

	// Display summary information
	summaryData := [][]string{
		{pterm.LightCyan("Strategy:"), pterm.Bold.Sprint(plan.Strategy)},
		{pterm.LightCyan("Region:"), pterm.Yellow(plan.Region)},
		{pterm.LightCyan("Application:"), pterm.Green(plan.AppName)},
	}

	for _, row := range summaryData {
		pterm.Printf("  %s %s\n", row[0], row[1])
	}

	pterm.Println()
	pterm.DefaultSection.Println("Resources to be Created")
	pterm.Println()

	// Build table data
	tableData := pterm.TableData{
		{
			pterm.Bold.Sprint("Resource Type"),
			pterm.Bold.Sprint("Name"),
			pterm.Bold.Sprint("Configuration"),
			pterm.Bold.Sprint("Value"),
		},
	}

	for _, resource := range plan.Resources {
		// Add resource type and name in the first row
		firstRow := true

		for key, value := range resource.Parameters {
			if firstRow {
				// Resource type and name in bold/colored
				resourceType := resource.Type
				if resource.Important {
					resourceType = pterm.LightMagenta(resourceType + " *")
				} else {
					resourceType = pterm.LightCyan(resourceType)
				}

				tableData = append(tableData, []string{
					resourceType,
					pterm.Yellow(resource.Name),
					"  " + pterm.LightBlue(key),
					pterm.Green(value),
				})
				firstRow = false
			} else {
				// Subsequent parameters indented
				tableData = append(tableData, []string{
					"",
					"",
					"  " + pterm.LightBlue(key),
					pterm.Green(value),
				})
			}
		}

		// Add a separator row for readability
		tableData = append(tableData, []string{"", "", "", ""})
	}

	// Render the table
	err := pterm.DefaultTable.
		WithHasHeader().
		WithHeaderRowSeparator("-").
		WithBoxed(true).
		WithData(tableData).
		Render()

	if err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	pterm.Println()
	pterm.Info.Println("* = Important resources (will incur costs)")
	pterm.Println()

	// Display cost warning for expensive strategies
	if plan.Strategy == "kubernetes" {
		pterm.Warning.Println("⚠️  EKS clusters incur charges (~$0.10/hour for control plane + node costs)")
	}

	return nil
}
