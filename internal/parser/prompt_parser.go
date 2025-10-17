package parser

import (
	"regexp"
	"strconv"
	"strings"
)

// DeploymentConfig holds parsed configuration from natural language
type DeploymentConfig struct {
	Strategy              string
	Region                string
	EC2InstanceType       string
	EC2VolumeSize         int
	LambdaMemory          int
	LambdaTimeout         int
	EKSNodeType           string
	EKSMinNodes           int
	EKSMaxNodes           int
	EKSDesiredNodes       int
	EKSNodeVolumeSize     int
	CleanedPrompt         string // Prompt with config keywords removed
}

// ParsePrompt extracts deployment configuration from natural language prompt
func ParsePrompt(prompt string) *DeploymentConfig {
	config := &DeploymentConfig{
		CleanedPrompt: prompt, // Will be cleaned as we extract
	}

	promptLower := strings.ToLower(prompt)

	// Extract strategy
	config.Strategy = extractStrategy(promptLower)

	// Extract region
	config.Region = extractRegion(promptLower)

	// Extract instance types
	config.EC2InstanceType = extractEC2InstanceType(promptLower)
	config.EKSNodeType = extractEKSNodeType(promptLower)

	// Extract node counts
	config.EKSMinNodes, config.EKSMaxNodes, config.EKSDesiredNodes = extractNodeCounts(promptLower)

	// Extract memory/storage
	config.LambdaMemory = extractLambdaMemory(promptLower)
	config.EC2VolumeSize = extractVolumeSize(promptLower)
	config.EKSNodeVolumeSize = extractVolumeSize(promptLower)

	// Extract timeout
	config.LambdaTimeout = extractTimeout(promptLower)

	// Clean the prompt (remove extracted config)
	config.CleanedPrompt = cleanPrompt(prompt, config)

	return config
}

// extractStrategy identifies deployment strategy from keywords
func extractStrategy(prompt string) string {
	// EKS/Kubernetes patterns
	if regexp.MustCompile(`\b(eks|kubernetes|k8s)\b`).MatchString(prompt) {
		return "kubernetes"
	}

	// Serverless/Lambda patterns
	if regexp.MustCompile(`\b(lambda|serverless|function)\b`).MatchString(prompt) {
		return "serverless"
	}

	// VM/EC2 patterns
	if regexp.MustCompile(`\b(ec2|vm|virtual machine|instance)\b`).MatchString(prompt) {
		return "vm"
	}

	return ""
}

// extractRegion extracts AWS region from prompt
func extractRegion(prompt string) string {
	// Pattern: us-east-1, eu-west-2, ap-south-1, etc.
	re := regexp.MustCompile(`\b(us|eu|ap|sa|ca|me|af)-(east|west|south|north|central|northeast|southeast)-[1-9]\b`)
	match := re.FindString(prompt)
	return match
}

// extractEC2InstanceType extracts EC2 instance type
func extractEC2InstanceType(prompt string) string {
	// Pattern: t3.micro, t3.small, t3.medium, t3.large, t3.xlarge, t3.2xlarge, etc.
	// Also support other families: t2, m5, c5, r5, etc.
	re := regexp.MustCompile(`\b(t2|t3|t4g|m5|m6i|c5|c6i|r5|r6i)\.(micro|nano|small|medium|large|xlarge|2xlarge|4xlarge|8xlarge|16xlarge)\b`)
	match := re.FindString(prompt)
	return match
}

// extractEKSNodeType extracts EKS node instance type
func extractEKSNodeType(prompt string) string {
	// Look for "node" or "nodes" followed by instance type
	// Or just use the instance type if strategy is EKS
	re := regexp.MustCompile(`\b(?:node[s]?\s+)?(t2|t3|t4g|m5|m6i|c5|c6i|r5|r6i)\.(micro|nano|small|medium|large|xlarge|2xlarge|4xlarge|8xlarge)\b`)
	match := re.FindString(prompt)

	// Clean up "nodes " prefix if present
	match = strings.TrimPrefix(match, "nodes ")
	match = strings.TrimPrefix(match, "node ")

	return match
}

// extractNodeCounts extracts min/max/desired node counts for EKS
func extractNodeCounts(prompt string) (min, max, desired int) {
	// Pattern: "3 nodes", "5 instances", "between 2 and 5 nodes", "min 1 max 3"

	// Simple pattern: "N nodes" or "N instances"
	re := regexp.MustCompile(`\b(\d+)\s+(?:node[s]?|instance[s]?)\b`)
	if matches := re.FindStringSubmatch(prompt); len(matches) > 1 {
		count, _ := strconv.Atoi(matches[1])
		return count, count, count // Same for all if single number
	}

	// Range pattern: "between X and Y nodes"
	re = regexp.MustCompile(`\bbetween\s+(\d+)\s+and\s+(\d+)\s+(?:node[s]?|instance[s]?)\b`)
	if matches := re.FindStringSubmatch(prompt); len(matches) > 2 {
		minVal, _ := strconv.Atoi(matches[1])
		maxVal, _ := strconv.Atoi(matches[2])
		desired := (minVal + maxVal) / 2
		return minVal, maxVal, desired
	}

	// Min/Max pattern: "min 1 max 3"
	reMin := regexp.MustCompile(`\bmin(?:imum)?\s+(\d+)\b`)
	reMax := regexp.MustCompile(`\bmax(?:imum)?\s+(\d+)\b`)

	if matches := reMin.FindStringSubmatch(prompt); len(matches) > 1 {
		min, _ = strconv.Atoi(matches[1])
	}

	if matches := reMax.FindStringSubmatch(prompt); len(matches) > 1 {
		max, _ = strconv.Atoi(matches[1])
	}

	if min > 0 && max > 0 {
		desired = (min + max) / 2
	} else if min > 0 {
		desired = min
		max = min
	} else if max > 0 {
		desired = max
		min = max
	}

	return min, max, desired
}

// extractLambdaMemory extracts Lambda memory in MB
func extractLambdaMemory(prompt string) int {
	// Pattern: "512MB", "1GB", "2048 MB", "1 GB"
	re := regexp.MustCompile(`\b(\d+)\s*(?:MB|mb|megabytes?)\b`)
	if matches := re.FindStringSubmatch(prompt); len(matches) > 1 {
		mb, _ := strconv.Atoi(matches[1])
		return mb
	}

	// GB pattern
	re = regexp.MustCompile(`\b(\d+)\s*(?:GB|gb|gigabytes?)\b`)
	if matches := re.FindStringSubmatch(prompt); len(matches) > 1 {
		gb, _ := strconv.Atoi(matches[1])
		return gb * 1024 // Convert to MB
	}

	return 0
}

// extractVolumeSize extracts volume size in GB
func extractVolumeSize(prompt string) int {
	// Pattern: "20GB volume", "30 GB disk", "50GB storage"
	re := regexp.MustCompile(`\b(\d+)\s*(?:GB|gb)\s+(?:volume|disk|storage)\b`)
	if matches := re.FindStringSubmatch(prompt); len(matches) > 1 {
		gb, _ := strconv.Atoi(matches[1])
		return gb
	}

	return 0
}

// extractTimeout extracts Lambda timeout in seconds
func extractTimeout(prompt string) int {
	// Pattern: "30 seconds timeout", "timeout 60s", "2 minutes timeout"
	re := regexp.MustCompile(`\b(?:timeout\s+)?(\d+)\s*(?:seconds?|secs?|s)\b`)
	if matches := re.FindStringSubmatch(prompt); len(matches) > 1 {
		sec, _ := strconv.Atoi(matches[1])
		return sec
	}

	// Minutes pattern
	re = regexp.MustCompile(`\b(?:timeout\s+)?(\d+)\s*(?:minutes?|mins?|m)\b`)
	if matches := re.FindStringSubmatch(prompt); len(matches) > 1 {
		min, _ := strconv.Atoi(matches[1])
		return min * 60 // Convert to seconds
	}

	return 0
}

// cleanPrompt removes extracted configuration keywords from prompt
func cleanPrompt(originalPrompt string, config *DeploymentConfig) string {
	cleaned := originalPrompt

	// Remove strategy keywords
	cleaned = regexp.MustCompile(`\b(?:on\s+)?(?:eks|kubernetes|k8s|lambda|serverless|ec2|vm|virtual machine)\b`).ReplaceAllString(cleaned, "")

	// Remove region
	if config.Region != "" {
		cleaned = strings.ReplaceAll(cleaned, config.Region, "")
		cleaned = regexp.MustCompile(`\b(?:in\s+)?(?:region\s+)?\b`).ReplaceAllString(cleaned, "")
	}

	// Remove instance types
	if config.EC2InstanceType != "" {
		cleaned = strings.ReplaceAll(cleaned, config.EC2InstanceType, "")
		cleaned = regexp.MustCompile(`\b(?:using\s+)?(?:instance\s+)?(?:type\s+)?\b`).ReplaceAllString(cleaned, "")
	}

	if config.EKSNodeType != "" {
		cleaned = strings.ReplaceAll(cleaned, config.EKSNodeType, "")
	}

	// Remove node count phrases
	cleaned = regexp.MustCompile(`\b\d+\s+(?:node[s]?|instance[s]?)\b`).ReplaceAllString(cleaned, "")
	cleaned = regexp.MustCompile(`\bbetween\s+\d+\s+and\s+\d+\s+(?:node[s]?|instance[s]?)\b`).ReplaceAllString(cleaned, "")
	cleaned = regexp.MustCompile(`\bmin(?:imum)?\s+\d+\b`).ReplaceAllString(cleaned, "")
	cleaned = regexp.MustCompile(`\bmax(?:imum)?\s+\d+\b`).ReplaceAllString(cleaned, "")

	// Remove memory/storage phrases
	cleaned = regexp.MustCompile(`\b\d+\s*(?:MB|GB|mb|gb)\b`).ReplaceAllString(cleaned, "")
	cleaned = regexp.MustCompile(`\b\d+\s*(?:MB|GB|mb|gb)\s+(?:volume|disk|storage|memory)\b`).ReplaceAllString(cleaned, "")

	// Remove timeout phrases
	cleaned = regexp.MustCompile(`\b(?:timeout\s+)?\d+\s*(?:seconds?|secs?|minutes?|mins?|s|m)\b`).ReplaceAllString(cleaned, "")

	// Clean up extra whitespace
	cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")
	cleaned = strings.TrimSpace(cleaned)

	// If cleaned is too short or empty, return original
	if len(cleaned) < 5 {
		return originalPrompt
	}

	return cleaned
}
