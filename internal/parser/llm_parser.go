package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/Smana/scia/internal/deployer"
	"github.com/Smana/scia/internal/llm"
)

const (
	maxLLMResponseSize = 10 * 1024 // 10KB max response
)

// ConfigExtractionPrompt is the template for extracting deployment config from natural language
// This matches the Terraform variables in types.TerraformConfig
const ConfigExtractionPrompt = `You are a deployment configuration expert. Extract deployment parameters from the user's natural language request.

**User Request:** %s

**Your Task:**
Analyze the request and extract any deployment configuration parameters mentioned.

**Available Parameters (matching Terraform variables):**

**Strategy & Region:**
- strategy: "vm", "kubernetes", or "serverless"
- region: AWS region (e.g., "eu-west-3", "us-east-1", "ap-south-1")

**EC2/VM Parameters (when strategy=vm):**
- ec2_instance_type: Instance type (e.g., "t3.micro", "t3.small", "t3.medium", "t3.large", "m5.large", "r5.xlarge")
- volume_size: Root volume size in GB (e.g., 30, 50, 100)

**EKS/Kubernetes Parameters (when strategy=kubernetes):**
- eks_node_type: Node instance type (e.g., "t3.medium", "t3.large", "m5.large")
- eks_min_nodes: Minimum number of nodes (integer)
- eks_max_nodes: Maximum number of nodes (integer)
- eks_desired_nodes: Desired number of nodes (integer)
- eks_node_volume_size: Node volume size in GB

**Lambda/Serverless Parameters (when strategy=serverless):**
- lambda_memory: Memory in MB (128-10240)
- lambda_timeout: Timeout in seconds (1-900)

**Response Format (JSON only):**
{
  "strategy": "vm",
  "region": "eu-west-3",
  "ec2_instance_type": "t3.medium",
  "volume_size": 30,
  "eks_node_type": "t3.medium",
  "eks_min_nodes": 1,
  "eks_max_nodes": 3,
  "eks_desired_nodes": 2,
  "eks_node_volume_size": 30,
  "lambda_memory": 512,
  "lambda_timeout": 30
}

**Important:**
- Only include parameters that are EXPLICITLY mentioned in the user's request
- Field names MUST match exactly: ec2_instance_type, volume_size, eks_node_type, etc.
- Instance types: preserve exact format (e.g., "t3.medium", not "T3.Medium" or "t3-medium")
- If user says "3 nodes", set eks_min_nodes, eks_max_nodes, and eks_desired_nodes all to 3
- Understand variations: "EKS"/"Kubernetes"/"K8s" → strategy="kubernetes", "VM"/"EC2" → strategy="vm"
- Omit fields that are not mentioned

**Respond with ONLY the JSON object, nothing else.**
`

// PlanModificationPrompt is for understanding plan modification requests
// This matches the Terraform variables in types.TerraformConfig
const PlanModificationPrompt = `You are a deployment configuration expert. The user wants to modify their deployment plan.

**Current Deployment Plan:**
Strategy: %s
Region: %s
%s

**User's Modification Request:** %s

**Your Task:**
Understand what the user wants to change and provide ONLY the changed parameters.

**Available Terraform Variables:**

**Strategy & Region:**
- strategy: "vm", "kubernetes", or "serverless"
- region: AWS region (e.g., "eu-west-3", "us-east-1")

**EC2/VM Parameters (when strategy=vm):**
- ec2_instance_type: Instance type (e.g., "t3.micro", "t3.small", "t3.medium", "t3.large", "m5.large")
- volume_size: Root volume size in GB

**EKS/Kubernetes Parameters (when strategy=kubernetes):**
- eks_node_type: Node instance type (e.g., "t3.medium", "t3.large")
- eks_min_nodes: Minimum number of nodes
- eks_max_nodes: Maximum number of nodes
- eks_desired_nodes: Desired number of nodes
- eks_node_volume_size: Node volume size in GB

**Lambda/Serverless Parameters (when strategy=serverless):**
- lambda_memory: Memory in MB (128-10240)
- lambda_timeout: Timeout in seconds (1-900)

**Parameter Extraction Examples:**
- "instance type t3.medium" → {"ec2_instance_type": "t3.medium"}
- "t3.large instance" → {"ec2_instance_type": "t3.large"}
- "change to t3.small" → {"ec2_instance_type": "t3.small"}
- "32GB disk" → {"volume_size": 32}
- "disk to 32GB" → {"volume_size": 32}
- "50 GB volume" → {"volume_size": 50}
- "5 nodes" → {"eks_desired_nodes": 5, "eks_min_nodes": 5, "eks_max_nodes": 5}
- "region eu-west-1" → {"region": "eu-west-1"}
- "32GB and t3.medium" → {"volume_size": 32, "ec2_instance_type": "t3.medium"}

**Response Format (JSON only - include ONLY changed parameters):**
{
  "ec2_instance_type": "t3.medium",
  "volume_size": 32
}

**Critical Requirements:**
- Field names MUST match EXACTLY: ec2_instance_type, volume_size, eks_node_type, etc.
- Instance types: exact format (e.g., "t3.medium", not "T3.Medium")
- Include ONLY parameters mentioned in the modification request
- Omit parameters that are not being changed
- If user says "instance type X", you MUST include "ec2_instance_type": "X"
- If user says "disk Y GB" or "Y GB", you MUST include "volume_size": Y

**Respond with ONLY the JSON object of CHANGED parameters, nothing else.**
`

// ParseConfigFromPrompt uses LLM to extract deployment configuration from natural language
func ParseConfigFromPrompt(llmClient *llm.Client, userPrompt string) (*DeploymentConfig, error) {
	if llmClient == nil {
		return &DeploymentConfig{CleanedPrompt: userPrompt}, nil
	}

	ctx := context.Background()

	// Build the prompt
	prompt := fmt.Sprintf(ConfigExtractionPrompt, userPrompt)

	// Generate using LLM
	req := &llm.GenerateRequest{
		Prompt:      prompt,
		Temperature: 0.1, // Low temperature for structured output
		MaxTokens:   300,
	}

	resp, err := llmClient.Generate(ctx, req)
	if err != nil {
		// If LLM fails, return empty config
		return &DeploymentConfig{CleanedPrompt: userPrompt}, nil
	}

	// Validate response size before parsing
	if len(resp.Text) > maxLLMResponseSize {
		log.Printf("Warning: LLM response exceeds max size (%d bytes), truncating", len(resp.Text))
		resp.Text = resp.Text[:maxLLMResponseSize]
	}

	// Log the LLM response for debugging
	log.Printf("LLM initial config response: %s", resp.Text)

	// Parse JSON response
	config, err := parseConfigJSON(resp.Text)
	if err != nil {
		// If parsing fails, return empty config
		log.Printf("Warning: Failed to parse LLM response as JSON: %v", err)
		return &DeploymentConfig{CleanedPrompt: userPrompt}, nil
	}

	// Log what was extracted
	log.Printf("Extracted initial config - EC2 Instance: %s, Volume: %dGB, Strategy: %s, Region: %s",
		config.EC2InstanceType, config.EC2VolumeSize, config.Strategy, config.Region)

	config.CleanedPrompt = userPrompt // Keep original prompt for context
	return config, nil
}

// ModifyPlanWithNaturalLanguage uses LLM to understand plan modification requests
func ModifyPlanWithNaturalLanguage(llmClient *llm.Client, currentConfig *deployer.DeployConfig, userRequest string) (*DeploymentConfig, error) {
	if llmClient == nil {
		return nil, fmt.Errorf("LLM client not available")
	}

	ctx := context.Background()

	// Build current plan description
	planDesc := buildCurrentPlanDescription(currentConfig)

	// Build the prompt
	prompt := fmt.Sprintf(PlanModificationPrompt,
		currentConfig.Strategy,
		currentConfig.AWSRegion,
		planDesc,
		userRequest,
	)

	// Generate using LLM
	req := &llm.GenerateRequest{
		Prompt:      prompt,
		Temperature: 0.1, // Low temperature for structured output
		MaxTokens:   300,
	}

	resp, err := llmClient.Generate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to parse modification request: %w", err)
	}

	// Log the LLM response for debugging
	log.Printf("LLM modification response: %s", resp.Text)

	// Parse JSON response
	config, err := parseConfigJSON(resp.Text)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Log what was extracted
	log.Printf("Extracted config - EC2 Instance: %s, Volume: %dGB, Strategy: %s, Region: %s",
		config.EC2InstanceType, config.EC2VolumeSize, config.Strategy, config.Region)

	return config, nil
}

// buildCurrentPlanDescription creates a human-readable description of the current plan
func buildCurrentPlanDescription(config *deployer.DeployConfig) string {
	var parts []string

	switch config.Strategy {
	case "vm":
		if config.EC2InstanceType != "" {
			parts = append(parts, fmt.Sprintf("EC2 Instance: %s", config.EC2InstanceType))
		}
		if config.EC2VolumeSize > 0 {
			parts = append(parts, fmt.Sprintf("Volume: %dGB", config.EC2VolumeSize))
		}

	case "kubernetes":
		if config.EKSNodeType != "" {
			parts = append(parts, fmt.Sprintf("Node Type: %s", config.EKSNodeType))
		}
		parts = append(parts, fmt.Sprintf("Nodes: %d (min: %d, max: %d)",
			config.EKSDesiredNodes, config.EKSMinNodes, config.EKSMaxNodes))
		if config.EKSNodeVolumeSize > 0 {
			parts = append(parts, fmt.Sprintf("Node Volume: %dGB", config.EKSNodeVolumeSize))
		}

	case "serverless":
		if config.LambdaMemory > 0 {
			parts = append(parts, fmt.Sprintf("Memory: %dMB", config.LambdaMemory))
		}
		if config.LambdaTimeout > 0 {
			parts = append(parts, fmt.Sprintf("Timeout: %ds", config.LambdaTimeout))
		}
	}

	return strings.Join(parts, ", ")
}

// parseConfigJSON parses the LLM's JSON response into a DeploymentConfig
func parseConfigJSON(jsonText string) (*DeploymentConfig, error) {
	// Extract JSON from response (LLM might add extra text)
	jsonText = extractJSON(jsonText)

	var rawConfig struct {
		Strategy          string `json:"strategy"`
		Region            string `json:"region"`
		EC2InstanceType   string `json:"ec2_instance_type"`
		EC2VolumeSize     int    `json:"volume_size"`
		EKSNodeType       string `json:"eks_node_type"`
		EKSMinNodes       int    `json:"eks_min_nodes"`
		EKSMaxNodes       int    `json:"eks_max_nodes"`
		EKSDesiredNodes   int    `json:"eks_desired_nodes"`
		EKSNodeVolumeSize int    `json:"eks_node_volume_size"`
		LambdaMemory      int    `json:"lambda_memory"`
		LambdaTimeout     int    `json:"lambda_timeout"`
	}

	if err := json.Unmarshal([]byte(jsonText), &rawConfig); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	config := &DeploymentConfig{
		Strategy:          rawConfig.Strategy,
		Region:            rawConfig.Region,
		EC2InstanceType:   rawConfig.EC2InstanceType,
		EC2VolumeSize:     rawConfig.EC2VolumeSize,
		EKSNodeType:       rawConfig.EKSNodeType,
		EKSMinNodes:       rawConfig.EKSMinNodes,
		EKSMaxNodes:       rawConfig.EKSMaxNodes,
		EKSDesiredNodes:   rawConfig.EKSDesiredNodes,
		EKSNodeVolumeSize: rawConfig.EKSNodeVolumeSize,
		LambdaMemory:      rawConfig.LambdaMemory,
		LambdaTimeout:     rawConfig.LambdaTimeout,
	}

	return config, nil
}

// extractJSON finds and extracts JSON object from text
func extractJSON(text string) string {
	// Find first { and last }
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")

	if start == -1 || end == -1 || start >= end {
		return "{}" // Return empty JSON instead of raw text
	}

	extracted := text[start : end+1]

	// Validate it's parseable JSON
	var test interface{}
	if err := json.Unmarshal([]byte(extracted), &test); err != nil {
		return "{}" // Return empty JSON on parse failure
	}

	return extracted
}

// ApplyConfig applies parsed configuration to deployer config
func ApplyConfig(deployConfig *deployer.DeployConfig, parsedConfig *DeploymentConfig) {
	if parsedConfig == nil {
		return
	}

	if parsedConfig.Strategy != "" {
		deployConfig.Strategy = parsedConfig.Strategy
	}

	if parsedConfig.Region != "" {
		deployConfig.AWSRegion = parsedConfig.Region
	}

	if parsedConfig.EC2InstanceType != "" {
		deployConfig.EC2InstanceType = parsedConfig.EC2InstanceType
	}

	if parsedConfig.EC2VolumeSize > 0 {
		deployConfig.EC2VolumeSize = parsedConfig.EC2VolumeSize
	}

	if parsedConfig.EKSNodeType != "" {
		deployConfig.EKSNodeType = parsedConfig.EKSNodeType
	}

	if parsedConfig.EKSMinNodes > 0 {
		deployConfig.EKSMinNodes = parsedConfig.EKSMinNodes
	}

	if parsedConfig.EKSMaxNodes > 0 {
		deployConfig.EKSMaxNodes = parsedConfig.EKSMaxNodes
	}

	if parsedConfig.EKSDesiredNodes > 0 {
		deployConfig.EKSDesiredNodes = parsedConfig.EKSDesiredNodes
	}

	if parsedConfig.EKSNodeVolumeSize > 0 {
		deployConfig.EKSNodeVolumeSize = parsedConfig.EKSNodeVolumeSize
	}

	if parsedConfig.LambdaMemory > 0 {
		deployConfig.LambdaMemory = parsedConfig.LambdaMemory
	}

	if parsedConfig.LambdaTimeout > 0 {
		deployConfig.LambdaTimeout = parsedConfig.LambdaTimeout
	}
}
