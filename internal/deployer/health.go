package deployer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// InstanceInfo contains information about an EC2 instance
type InstanceInfo struct {
	InstanceID string
	PublicIP   string
	PrivateIP  string
	State      string
}

// GetASGInstance retrieves the public IP of the first running instance in an ASG
func GetASGInstance(ctx context.Context, asgName, region string, verbose bool) (*InstanceInfo, error) {
	if verbose {
		fmt.Printf("   Looking up instance in ASG: %s\n", asgName)
	}

	// Get instance IDs from ASG
	// #nosec G204 -- AWS CLI with controlled arguments (region and asgName are from Terraform outputs)
	cmd := exec.CommandContext(ctx, "aws", "autoscaling", "describe-auto-scaling-groups",
		"--auto-scaling-group-names", asgName,
		"--region", region,
		"--query", "AutoScalingGroups[0].Instances[?HealthStatus=='Healthy' && LifecycleState=='InService'].InstanceId",
		"--output", "json")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get ASG instances: %w", err)
	}

	var instanceIDs []string
	if err := json.Unmarshal(output, &instanceIDs); err != nil {
		return nil, fmt.Errorf("failed to parse instance IDs: %w", err)
	}

	if len(instanceIDs) == 0 {
		return nil, fmt.Errorf("no healthy instances found in ASG")
	}

	instanceID := instanceIDs[0]
	if verbose {
		fmt.Printf("   Found instance: %s\n", instanceID)
	}

	// Get instance details
	// #nosec G204 -- AWS CLI with controlled arguments (region and instanceID are validated by AWS SDK)
	cmd = exec.CommandContext(ctx, "aws", "ec2", "describe-instances",
		"--instance-ids", instanceID,
		"--region", region,
		"--query", "Reservations[0].Instances[0].{PublicIpAddress:PublicIpAddress,PrivateIpAddress:PrivateIpAddress,State:State.Name}",
		"--output", "json")

	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get instance details: %w", err)
	}

	var result struct {
		PublicIpAddress  string `json:"PublicIpAddress"`
		PrivateIpAddress string `json:"PrivateIpAddress"`
		State            string `json:"State"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse instance details: %w", err)
	}

	return &InstanceInfo{
		InstanceID: instanceID,
		PublicIP:   result.PublicIpAddress,
		PrivateIP:  result.PrivateIpAddress,
		State:      result.State,
	}, nil
}

// WaitForASGInstance waits for an instance to be running in the ASG
func WaitForASGInstance(ctx context.Context, asgName, region string, timeout time.Duration, verbose bool) (*InstanceInfo, error) {
	if verbose {
		fmt.Printf("   Waiting for instance to be ready (timeout: %v)...\n", timeout)
	}

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			info, err := GetASGInstance(ctx, asgName, region, false)
			if err == nil && info.State == "running" && info.PublicIP != "" {
				if verbose {
					fmt.Printf("   ✓ Instance is running: %s (IP: %s)\n", info.InstanceID, info.PublicIP)
				}
				return info, nil
			}

			if verbose && err != nil {
				fmt.Printf("   Still waiting for instance... (%v)\n", err)
			}

			if time.Now().After(deadline) {
				return nil, fmt.Errorf("timeout waiting for instance to be ready")
			}
		}
	}
}

// WaitForApplicationReady waits for the application to respond to HTTP requests
func WaitForApplicationReady(ctx context.Context, url string, timeout time.Duration, verbose bool) error {
	if verbose {
		fmt.Printf("   Waiting for application to be ready at %s (timeout: %v)...\n", url, timeout)
	}

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	attempt := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			attempt++
			resp, err := client.Get(url)
			if err == nil {
				_ = resp.Body.Close()
				if resp.StatusCode < 500 {
					if verbose {
						fmt.Printf("   ✓ Application is ready! (HTTP %d)\n", resp.StatusCode)
					}
					return nil
				}
				if verbose {
					fmt.Printf("   Attempt %d: Received HTTP %d, waiting...\n", attempt, resp.StatusCode)
				}
			} else if verbose {
				fmt.Printf("   Attempt %d: %v\n", attempt, err)
			}

			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for application to be ready")
			}
		}
	}
}

// GetApplicationURL constructs the application URL and waits for it to be ready
func GetApplicationURL(ctx context.Context, asgName, region string, port int, verbose bool) (string, error) {
	// Wait for instance to be running (5 minute timeout)
	info, err := WaitForASGInstance(ctx, asgName, region, 5*time.Minute, verbose)
	if err != nil {
		return "", fmt.Errorf("failed to get running instance: %w", err)
	}

	// Construct URL
	url := fmt.Sprintf("http://%s:%d", info.PublicIP, port)

	// Wait for application to be ready (5 minute timeout)
	if err := WaitForApplicationReady(ctx, url, 5*time.Minute, verbose); err != nil {
		// Return URL even if health check fails, with a warning
		return url, fmt.Errorf("application may not be ready yet: %w (URL: %s)", err, url)
	}

	return url, nil
}

// ParsePort converts a string port to int
func ParsePort(portStr string) (int, error) {
	port, err := strconv.Atoi(strings.TrimSpace(portStr))
	if err != nil {
		return 0, fmt.Errorf("invalid port: %w", err)
	}
	return port, nil
}
