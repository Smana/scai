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
const ConfigExtractionPrompt = `You are a deployment configuration expert. Extract deployment parameters from the user's natural language request.

**User Request:** %s

**Your Task:**
Analyze the request and extract any deployment configuration parameters mentioned.

**Available Parameters:**
- strategy: vm, kubernetes, serverless
- region: AWS regions (us-east-1, eu-west-1, ap-south-1, etc.)
- ec2_instance_type: AWS EC2 instance types (e.g., t3.large, r5.xlarge, m5.2xlarge)
- eks_node_type: AWS EC2 instance types for Kubernetes nodes
- eks_nodes: number of nodes (min, max, desired)
- lambda_memory: memory in MB (128-10240)
- volume_size: disk size in GB

**Response Format (JSON only):**
{
  "strategy": "vm|kubernetes|serverless",
  "region": "us-east-1",
  "ec2_instance_type": "t3.large",
  "eks_node_type": "t3.medium",
  "eks_min_nodes": 1,
  "eks_max_nodes": 3,
  "eks_desired_nodes": 2,
  "lambda_memory": 512,
  "volume_size": 20
}

**Important:**
- Only include parameters that are explicitly mentioned
- Preserve exact instance types when specified (e.g., "r5.large", "m5.xlarge", "t3.medium")
- If user says "3 nodes", set all node counts to 3
- Understand variations: "EKS", "Kubernetes", "K8s" all mean strategy=kubernetes
- Empty/missing fields mean "not specified"

**Respond with ONLY the JSON object, nothing else.**
`

// PlanModificationPrompt is for understanding plan modification requests
const PlanModificationPrompt = `You are a deployment configuration expert. The user wants to modify their deployment plan.

**Current Deployment Plan:**
Strategy: %s
Region: %s
%s

**User's Modification Request:** %s

**Your Task:**
Understand what the user wants to change and provide the updated configuration.

**Response Format (JSON only):**
{
  "strategy": "vm|kubernetes|serverless",
  "region": "us-east-1",
  "ec2_instance_type": "t3.large",
  "eks_node_type": "t3.medium",
  "eks_min_nodes": 1,
  "eks_max_nodes": 3,
  "eks_desired_nodes": 2,
  "lambda_memory": 512,
  "volume_size": 20
}

**Important:**
- Only include parameters that should be CHANGED
- Keep unmentioned parameters unchanged
- Understand variations: "bigger instance" → upgrade instance type
- "5 nodes" → update node counts
- "us-east-1" → update region

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

	// Parse JSON response
	config, err := parseConfigJSON(resp.Text)
	if err != nil {
		// If parsing fails, return empty config
		return &DeploymentConfig{CleanedPrompt: userPrompt}, nil
	}

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

	// Parse JSON response
	config, err := parseConfigJSON(resp.Text)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

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
		Strategy        string `json:"strategy"`
		Region          string `json:"region"`
		EC2InstanceType string `json:"ec2_instance_type"`
		EC2VolumeSize   int    `json:"volume_size"`
		EKSNodeType     string `json:"eks_node_type"`
		EKSMinNodes     int    `json:"eks_min_nodes"`
		EKSMaxNodes     int    `json:"eks_max_nodes"`
		EKSDesiredNodes int    `json:"eks_desired_nodes"`
		LambdaMemory    int    `json:"lambda_memory"`
		LambdaTimeout   int    `json:"lambda_timeout"`
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
		LambdaMemory:      rawConfig.LambdaMemory,
		LambdaTimeout:     rawConfig.LambdaTimeout,
		EKSNodeVolumeSize: rawConfig.EC2VolumeSize, // Same volume size
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
