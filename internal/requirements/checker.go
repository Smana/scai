package requirements

import (
	"fmt"
	"os/exec"
	"strings"
)

const (
	providerOllama = "ollama"
	versionUnknown = "unknown"
)

// Requirement represents a system requirement
type Requirement struct {
	Name        string // Display name (e.g., "OpenTofu")
	Binary      string // Binary to check (e.g., "tofu")
	Required    bool   // Whether this is mandatory
	Installed   bool   // Whether it's installed
	Version     string // Version string (if available)
	Description string // What it's used for
}

// CheckRequirements checks all system requirements
func CheckRequirements(llmProvider string, useDocker bool) ([]Requirement, error) {
	requirements := []Requirement{
		{
			Name:        "OpenTofu/Terraform",
			Binary:      "tofu", // Check tofu first, can fallback to terraform
			Required:    true,
			Description: "Infrastructure as Code provisioning",
		},
		{
			Name:        "AWS CLI",
			Binary:      "aws",
			Required:    false,
			Description: "AWS credentials and configuration",
		},
	}

	// Add Docker requirement if using Ollama with Docker
	if llmProvider == providerOllama && useDocker {
		requirements = append(requirements, Requirement{
			Name:        "Docker",
			Binary:      "docker",
			Required:    true,
			Description: "Container runtime for Ollama",
		})
	}

	// Add Ollama requirement if using Ollama without Docker
	if llmProvider == providerOllama && !useDocker {
		requirements = append(requirements, Requirement{
			Name:        "Ollama",
			Binary:      "ollama",
			Required:    true,
			Description: "Local LLM runtime",
		})
	}

	// Check each requirement
	for i := range requirements {
		installed, version := checkBinary(requirements[i].Binary)
		requirements[i].Installed = installed
		requirements[i].Version = version

		// For terraform/tofu, try fallback
		if requirements[i].Binary == "tofu" && !installed {
			// Try terraform as fallback
			installed, version = checkBinary("terraform")
			if installed {
				requirements[i].Binary = "terraform"
				requirements[i].Installed = true
				requirements[i].Version = version
			}
		}
	}

	return requirements, nil
}

// checkBinary checks if a binary exists and gets its version
func checkBinary(binaryName string) (installed bool, version string) {
	// Check if binary exists in PATH
	path, err := exec.LookPath(binaryName)
	if err != nil {
		return false, ""
	}

	// Binary exists
	if path == "" {
		return false, ""
	}

	// Try to get version
	version = getVersion(binaryName)

	return true, version
}

// getVersion attempts to get the version of a binary
func getVersion(binaryName string) string {
	// Try common version flags
	versionFlags := []string{"--version", "-v", "version"}

	for _, flag := range versionFlags {
		// #nosec G204 -- binaryName is from a controlled list of system binaries (tofu, aws, docker, ollama)
		cmd := exec.Command(binaryName, flag)
		output, err := cmd.CombinedOutput()
		if err == nil && len(output) > 0 {
			// Parse first line of output
			lines := strings.Split(string(output), "\n")
			if len(lines) > 0 {
				version := strings.TrimSpace(lines[0])
				// Clean up common prefixes
				version = strings.TrimPrefix(version, binaryName+" ")
				version = strings.TrimPrefix(version, "version ")
				version = strings.TrimPrefix(version, "Version ")
				version = strings.TrimPrefix(version, "v")
				return version
			}
		}
	}

	return versionUnknown
}

// AllRequiredInstalled checks if all required binaries are installed
func AllRequiredInstalled(reqs []Requirement) bool {
	for _, req := range reqs {
		if req.Required && !req.Installed {
			return false
		}
	}
	return true
}

// GetMissingRequired returns a list of missing required binaries
func GetMissingRequired(reqs []Requirement) []string {
	var missing []string
	for _, req := range reqs {
		if req.Required && !req.Installed {
			missing = append(missing, req.Name)
		}
	}
	return missing
}

// FormatRequirementStatus formats a requirement for display
func FormatRequirementStatus(req Requirement) string {
	status := "✗"
	if req.Installed {
		status = "✓"
	}

	name := req.Name
	if req.Required {
		name += " (required)"
	}

	if req.Installed && req.Version != "" && req.Version != versionUnknown {
		return fmt.Sprintf("%s %s [%s]", status, name, req.Version)
	}

	return fmt.Sprintf("%s %s", status, name)
}
