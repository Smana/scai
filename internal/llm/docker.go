package llm

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	OllamaContainerName = "scia-ollama"
	OllamaImage         = "ollama/ollama"
	OllamaPort          = "11434"
	OllamaDockerURL     = "http://localhost:11434"
)

// IsDockerAvailable checks if Docker is installed and running
func IsDockerAvailable() bool {
	cmd := exec.Command("docker", "ps")
	err := cmd.Run()
	return err == nil
}

// IsOllamaContainerRunning checks if the SCIA Ollama container is running
func IsOllamaContainerRunning() bool {
	cmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", OllamaContainerName), "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == OllamaContainerName
}

// StartOllamaContainer starts the Ollama Docker container
func StartOllamaContainer(verbose bool) error {
	// Check if container exists but is stopped
	checkCmd := exec.Command("docker", "ps", "-a", "--filter", fmt.Sprintf("name=%s", OllamaContainerName), "--format", "{{.Names}}")
	output, _ := checkCmd.Output()

	if strings.TrimSpace(string(output)) == OllamaContainerName {
		// Container exists, just start it
		if verbose {
			fmt.Printf("Starting existing Ollama container...\n")
		}
		cmd := exec.Command("docker", "start", OllamaContainerName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start existing container: %w", err)
		}
	} else {
		// Create new container with security options
		if verbose {
			fmt.Printf("Creating Ollama container...\n")
		}
		cmd := exec.Command("docker", "run", "-d",
			"--name", OllamaContainerName,
			"-p", fmt.Sprintf("%s:%s", OllamaPort, OllamaPort),
			"-v", "ollama-data:/root/.ollama",
			"--security-opt", "no-new-privileges:true",
			"--memory", "8g", // Limit memory to 8GB
			"--cpus", "4.0", // Limit to 4 CPUs
			OllamaImage,
		)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to create container: %w\nOutput: %s", err, string(output))
		}
	}

	// Wait for container to be ready
	if verbose {
		fmt.Printf("Waiting for Ollama to be ready...\n")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for Ollama to start")
		default:
			if IsOllamaAccessible(OllamaDockerURL) {
				if verbose {
					fmt.Printf("✓ Ollama container is ready\n")
				}
				return nil
			}
			time.Sleep(1 * time.Second)
		}
	}
}

// EnsureModelAvailable ensures the specified model is pulled
func EnsureModelAvailable(model string, verbose bool) error {
	// Check if model exists
	checkCmd := exec.Command("docker", "exec", OllamaContainerName, "ollama", "list")
	output, err := checkCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	// Check if model is already present
	if strings.Contains(string(output), model) {
		if verbose {
			fmt.Printf("✓ Model %s is already available\n", model)
		}
		return nil
	}

	// Pull the model
	if verbose {
		fmt.Printf("Pulling model %s (this may take a while)...\n", model)
	}

	pullCmd := exec.Command("docker", "exec", OllamaContainerName, "ollama", "pull", model)

	if verbose {
		// Show progress to user
		pullCmd.Stdout = os.Stdout
		pullCmd.Stderr = os.Stderr
	} else {
		// Suppress progress
		pullCmd.Stdout = nil
		pullCmd.Stderr = nil
	}

	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("failed to pull model %s: %w", model, err)
	}

	if verbose {
		fmt.Printf("✓ Model %s is ready\n", model)
	}
	return nil
}

// IsOllamaAccessible checks if Ollama is accessible at the given URL
func IsOllamaAccessible(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/version", url), nil)
	if err != nil {
		return false
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode == http.StatusOK
}

// SetupOllamaDocker ensures Ollama Docker container is running with the required model
func SetupOllamaDocker(model string, verbose bool) (string, error) {
	if !IsDockerAvailable() {
		return "", fmt.Errorf("Docker is not available")
	}

	if verbose {
		fmt.Println("🐳 Setting up Ollama with Docker...")
	}

	// Check if container is already running
	if !IsOllamaContainerRunning() {
		if err := StartOllamaContainer(verbose); err != nil {
			return "", err
		}
	} else if verbose {
		fmt.Printf("✓ Ollama container is already running\n")
	}

	// Ensure model is available
	if err := EnsureModelAvailable(model, verbose); err != nil {
		return "", err
	}

	return OllamaDockerURL, nil
}
