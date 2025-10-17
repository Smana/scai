package ui

import (
	"fmt"

	"github.com/Smana/scia/internal/deployer"
	"github.com/Smana/scia/internal/llm"
	"github.com/Smana/scia/internal/parser"
	"github.com/Smana/scia/internal/types"
	"github.com/pterm/pterm"
)

// ConfirmOrModify shows the plan and allows confirmation or modification
func ConfirmOrModify(plan *DeploymentPlan, analysis *types.Analysis, config *deployer.DeployConfig, llmClient *llm.Client, autoApprove bool) (bool, *deployer.DeployConfig, error) {
	// Display the plan
	if err := DisplayPlanTable(plan); err != nil {
		return false, config, fmt.Errorf("failed to display plan: %w", err)
	}

	// Skip confirmation if --yes flag is set
	if autoApprove {
		pterm.Success.Println("Auto-confirmed with --yes flag")
		return true, config, nil
	}

	pterm.Println()

	// Interactive modification loop
	for {
		// Offer modification option
		pterm.Info.Println("You can:")
		pterm.Println("  • Type 'yes' or 'y' to proceed with deployment")
		pterm.Println("  • Type 'no' or 'n' to cancel")
		pterm.Println("  • Describe changes in natural language (e.g., 'use t3.large instance', 'change to 5 nodes')")
		pterm.Println()

		// Get user input
		userInput, err := pterm.DefaultInteractiveTextInput.
			WithDefaultText("Your choice").
			Show()

		if err != nil {
			return false, config, fmt.Errorf("failed to read input: %w", err)
		}

		userInput = pterm.FgLightWhite.Sprint(userInput)
		response := pterm.FgCyan.Sprint(userInput)

		// Check for yes/no
		if response == "yes" || response == "y" {
			pterm.Success.Println("✓ Deployment confirmed")
			return true, config, nil
		}

		if response == "no" || response == "n" {
			return false, config, nil
		}

		// User wants to modify - use LLM to understand the request
		pterm.Info.Printf("Processing modification request: %s\n", response)
		pterm.Println()

		// Use LLM to parse modification
		modifiedConfig, err := parser.ModifyPlanWithNaturalLanguage(llmClient, config, response)
		if err != nil {
			pterm.Warning.Printf("Could not understand modification: %v\n", err)
			pterm.Warning.Println("Please try rephrasing or use specific values")
			pterm.Println()
			continue
		}

		// Apply modifications to config
		parser.ApplyConfig(config, modifiedConfig)

		// Rebuild plan with modified config
		appName := plan.AppName
		plan = BuildDeploymentPlan(config.Strategy, config.AWSRegion, appName, analysis, config)

		// Show updated plan
		pterm.Println()
		pterm.Success.Println("✓ Plan updated based on your request")
		pterm.Println()

		if err := DisplayPlanTable(plan); err != nil {
			return false, config, fmt.Errorf("failed to display updated plan: %w", err)
		}

		pterm.Println()
	}
}
