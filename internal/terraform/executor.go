package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Executor handles Terraform/OpenTofu command execution
type Executor struct {
	workDir string
	tfBin   string
	verbose bool
}

// NewExecutor creates a new Terraform executor with path validation
func NewExecutor(workDir, tfBin string, verbose bool) (*Executor, error) {
	// Validate tfBin is an absolute path or in PATH
	validatedBin, err := validateTerraformBinary(tfBin)
	if err != nil {
		return nil, fmt.Errorf("invalid terraform binary: %w", err)
	}

	return &Executor{
		workDir: workDir,
		tfBin:   validatedBin,
		verbose: verbose,
	}, nil
}

// validateTerraformBinary ensures the binary is safe to execute
func validateTerraformBinary(bin string) (string, error) {
	// Allow only specific binary names
	allowedBinaries := map[string]bool{
		"terraform": true,
		"tofu":      true,
	}

	baseName := filepath.Base(bin)
	if !allowedBinaries[baseName] {
		return "", fmt.Errorf("binary name must be 'terraform' or 'tofu', got: %s", baseName)
	}

	// Check if binary exists and is executable
	absPath, err := exec.LookPath(bin)
	if err != nil {
		return "", fmt.Errorf("binary not found in PATH: %w", err)
	}

	// Verify it's a regular file
	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("cannot stat binary: %w", err)
	}

	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("not a regular file: %s", absPath)
	}

	return absPath, nil
}

// Init initializes Terraform in the working directory
func (e *Executor) Init() error {
	args := []string{"init", "-reconfigure"}
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

	// Parse JSON output properly
	// Format: {"output_name": {"value": "output_value", "type": "string", "sensitive": false}}
	var rawOutputs map[string]struct {
		Value     interface{} `json:"value"`
		Type      string      `json:"type"`
		Sensitive bool        `json:"sensitive"`
	}

	if err := json.Unmarshal(output, &rawOutputs); err != nil {
		return nil, fmt.Errorf("failed to parse terraform outputs: %w", err)
	}

	outputs := make(map[string]string, len(rawOutputs))
	for key, val := range rawOutputs {
		// Convert value to string
		switch v := val.Value.(type) {
		case string:
			outputs[key] = v
		case float64:
			outputs[key] = fmt.Sprintf("%.0f", v)
		case bool:
			outputs[key] = fmt.Sprintf("%t", v)
		default:
			// For complex types (arrays, objects), marshal back to JSON
			jsonBytes, _ := json.Marshal(v)
			outputs[key] = string(jsonBytes)
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
		// Stream output in real-time to stdout/stderr
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Run command with live output
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command failed: %s %s\nError: %w",
				e.tfBin, strings.Join(args, " "), err)
		}
		return nil
	}

	// Non-verbose mode: capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %s %s\nError: %w\nOutput: %s",
			e.tfBin, strings.Join(args, " "), err, string(output))
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
