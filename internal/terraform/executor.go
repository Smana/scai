package terraform

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Executor handles Terraform/OpenTofu command execution
type Executor struct {
	workDir string
	tfBin   string
	verbose bool
}

// NewExecutor creates a new Terraform executor
func NewExecutor(workDir, tfBin string, verbose bool) *Executor {
	return &Executor{
		workDir: workDir,
		tfBin:   tfBin,
		verbose: verbose,
	}
}

// Init initializes Terraform in the working directory
func (e *Executor) Init() error {
	args := []string{"init"}
	if !e.verbose {
		args = append(args, "-input=false")
	}

	return e.runCommand(args...)
}

// Plan runs terraform plan
func (e *Executor) Plan() error {
	args := []string{"plan", "-input=false"}
	if !e.verbose {
		args = append(args, "-no-color")
	}

	return e.runCommand(args...)
}

// Apply runs terraform apply with auto-approve
func (e *Executor) Apply() error {
	args := []string{"apply", "-auto-approve", "-input=false"}
	if !e.verbose {
		args = append(args, "-no-color")
	}

	return e.runCommand(args...)
}

// Destroy runs terraform destroy
func (e *Executor) Destroy() error {
	args := []string{"destroy", "-auto-approve", "-input=false"}
	if !e.verbose {
		args = append(args, "-no-color")
	}

	return e.runCommand(args...)
}

// Outputs retrieves terraform outputs as a map
func (e *Executor) Outputs() (map[string]string, error) {
	cmd := exec.Command(e.tfBin, "output", "-json")
	cmd.Dir = e.workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		// If no outputs exist, return empty map
		if strings.Contains(string(output), "no outputs") {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("failed to get outputs: %w\nOutput: %s", err, string(output))
	}

	// Parse JSON output
	outputs := make(map[string]string)

	// Simple JSON parsing for outputs
	// Format: {"output_name": {"value": "output_value"}}
	outputStr := string(output)
	if outputStr == "{}\n" || outputStr == "{}" {
		return outputs, nil
	}

	// Extract outputs using simple string parsing
	// This is a simplified approach - in production, use encoding/json
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "\"value\":") {
			// Extract key-value pairs
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				key := strings.Trim(parts[0], " \"")
				value := strings.Trim(parts[1], " \",")
				if key != "value" && key != "type" && key != "sensitive" {
					outputs[key] = value
				}
			}
		}
	}

	return outputs, nil
}

// runCommand executes a terraform command
func (e *Executor) runCommand(args ...string) error {
	cmd := exec.Command(e.tfBin, args...)
	cmd.Dir = e.workDir

	if e.verbose {
		fmt.Printf("   Executing: %s %s\n", e.tfBin, strings.Join(args, " "))
		cmd.Stdout = nil // Let it print to stdout
		cmd.Stderr = nil // Let it print to stderr
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %s %s\nError: %w\nOutput: %s",
			e.tfBin, strings.Join(args, " "), err, string(output))
	}

	if e.verbose && len(output) > 0 {
		fmt.Printf("%s\n", string(output))
	}

	return nil
}

// Validate runs terraform validate
func (e *Executor) Validate() error {
	args := []string{"validate"}
	if !e.verbose {
		args = append(args, "-json")
	}

	return e.runCommand(args...)
}

// GetState retrieves the current terraform state
func (e *Executor) GetState() (string, error) {
	cmd := exec.Command(e.tfBin, "show", "-json")
	cmd.Dir = e.workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get state: %w", err)
	}

	return string(output), nil
}

// Version returns the terraform/tofu version
func (e *Executor) Version(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, e.tfBin, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}

	return string(output), nil
}
