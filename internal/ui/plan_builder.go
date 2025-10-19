package ui

import (
	"fmt"

	"github.com/Smana/scai/internal/deployer"
	"github.com/Smana/scai/internal/types"
)

// BuildDeploymentPlan creates a deployment plan based on the strategy and configuration
func BuildDeploymentPlan(strategy, region, appName string, analysis *types.Analysis, config *deployer.DeployConfig) *DeploymentPlan {
	plan := &DeploymentPlan{
		Strategy:  strategy,
		Region:    region,
		AppName:   appName,
		Resources: []ResourceConfig{},
	}

	switch strategy {
	case "vm":
		plan.Resources = buildEC2Resources(appName, region, analysis, config)
	case "serverless":
		plan.Resources = buildLambdaResources(appName, region, analysis, config)
	case "kubernetes":
		plan.Resources = buildEKSResources(appName, region, analysis, config)
	default:
		// Fallback to VM
		plan.Resources = buildEC2Resources(appName, region, analysis, config)
	}

	return plan
}

// buildEC2Resources builds resource list for EC2/VM deployment
func buildEC2Resources(appName, region string, analysis *types.Analysis, config *deployer.DeployConfig) []ResourceConfig {
	resources := []ResourceConfig{}

	// VPC
	vpcResource := ResourceConfig{
		Type:       "VPC",
		Name:       "Default VPC",
		Parameters: make(map[string]string),
		Important:  false,
	}
	vpcResource.AddParameter("Type", "Default VPC")
	vpcResource.AddParameter("Region", region)
	resources = append(resources, vpcResource)

	// Security Group
	sgResource := ResourceConfig{
		Type:       "Security Group",
		Name:       fmt.Sprintf("%s-sg", appName),
		Parameters: make(map[string]string),
		Important:  true,
	}
	sgResource.AddParameter("Ingress Ports", fmt.Sprintf("22 (SSH), %d (App)", analysis.Port))
	sgResource.AddParameter("Egress", "All traffic")
	sgResource.AddParameter("CIDR", "0.0.0.0/0")
	resources = append(resources, sgResource)

	// Auto Scaling Group
	asgResource := ResourceConfig{
		Type:       "Auto Scaling Group",
		Name:       fmt.Sprintf("%s-asg", appName),
		Parameters: make(map[string]string),
		Important:  true,
	}
	asgResource.AddParameter("Min/Max/Desired", "1/1/1")
	asgResource.AddParameter("Health Check Type", "EC2")
	asgResource.AddParameter("Health Check Grace Period", "300s")
	resources = append(resources, asgResource)

	// EC2 Instance
	instanceType := config.EC2InstanceType
	if instanceType == "" {
		instanceType = "t3.micro"
	}

	ec2Resource := ResourceConfig{
		Type:       "EC2 Instance",
		Name:       fmt.Sprintf("%s (via ASG)", appName),
		Parameters: make(map[string]string),
		Important:  true,
	}
	ec2Resource.AddParameter("Instance Type", instanceType)
	ec2Resource.AddParameter("AMI", "Amazon Linux 2023 (latest)")
	ec2Resource.AddParameter("Volume Size", fmt.Sprintf("%d GB", config.EC2VolumeSize))
	ec2Resource.AddParameter("Volume Type", "GP3 (encrypted)")
	ec2Resource.AddParameter("Monitoring", "Enabled")
	resources = append(resources, ec2Resource)

	return resources
}

// buildLambdaResources builds resource list for Lambda deployment
func buildLambdaResources(appName, region string, analysis *types.Analysis, config *deployer.DeployConfig) []ResourceConfig {
	resources := []ResourceConfig{}

	// IAM Role
	iamResource := ResourceConfig{
		Type:       "IAM Role",
		Name:       fmt.Sprintf("%s-lambda-role", appName),
		Parameters: make(map[string]string),
		Important:  false,
	}
	iamResource.AddParameter("Service", "lambda.amazonaws.com")
	iamResource.AddParameter("Policies", "AWSLambdaBasicExecutionRole")
	resources = append(resources, iamResource)

	// Lambda Function
	runtime := detectRuntime(analysis.Language, analysis.Framework)
	lambdaResource := ResourceConfig{
		Type:       "Lambda Function",
		Name:       appName,
		Parameters: make(map[string]string),
		Important:  true,
	}
	lambdaResource.AddParameter("Runtime", runtime)
	lambdaResource.AddParameter("Memory", fmt.Sprintf("%d MB", config.LambdaMemory))
	lambdaResource.AddParameter("Timeout", fmt.Sprintf("%d seconds", config.LambdaTimeout))
	if config.LambdaReservedConcurrency > 0 {
		lambdaResource.AddParameter("Reserved Concurrency", fmt.Sprintf("%d", config.LambdaReservedConcurrency))
	} else {
		lambdaResource.AddParameter("Reserved Concurrency", "Unreserved")
	}
	lambdaResource.AddParameter("Tracing", "X-Ray Active")
	resources = append(resources, lambdaResource)

	// CloudWatch Log Group
	logResource := ResourceConfig{
		Type:       "CloudWatch Logs",
		Name:       fmt.Sprintf("/aws/lambda/%s", appName),
		Parameters: make(map[string]string),
		Important:  false,
	}
	logResource.AddParameter("Retention", "7 days")
	resources = append(resources, logResource)

	// API Gateway
	apiResource := ResourceConfig{
		Type:       "API Gateway HTTP API",
		Name:       fmt.Sprintf("%s-api", appName),
		Parameters: make(map[string]string),
		Important:  true,
	}
	apiResource.AddParameter("Protocol", "HTTP")
	apiResource.AddParameter("Routes", "ANY / and ANY /{proxy+}")
	apiResource.AddParameter("CORS", "Enabled (all origins)")
	apiResource.AddParameter("Integration", "Lambda proxy")
	resources = append(resources, apiResource)

	return resources
}

// buildEKSResources builds resource list for EKS deployment
func buildEKSResources(appName, region string, analysis *types.Analysis, config *deployer.DeployConfig) []ResourceConfig {
	resources := []ResourceConfig{}

	// VPC
	vpcResource := ResourceConfig{
		Type:       "VPC",
		Name:       fmt.Sprintf("%s-vpc", appName),
		Parameters: make(map[string]string),
		Important:  true,
	}
	vpcResource.AddParameter("CIDR Block", "10.0.0.0/16")
	vpcResource.AddParameter("Availability Zones", "2")
	vpcResource.AddParameter("Private Subnets", "10.0.1.0/24, 10.0.2.0/24")
	vpcResource.AddParameter("Public Subnets", "10.0.101.0/24, 10.0.102.0/24")
	vpcResource.AddParameter("NAT Gateway", "Single (in public subnet)")
	resources = append(resources, vpcResource)

	// EKS Cluster
	eksResource := ResourceConfig{
		Type:       "EKS Cluster",
		Name:       fmt.Sprintf("%s-eks", appName),
		Parameters: make(map[string]string),
		Important:  true,
	}
	eksResource.AddParameter("Kubernetes Version", "1.31")
	eksResource.AddParameter("Endpoint Access", "Public")
	eksResource.AddParameter("Cluster Logging", "API, Audit, Authenticator")
	eksResource.AddParameter("Encryption", "Secrets encrypted with KMS")
	eksResource.AddParameter("Pod Identity", "Enabled")
	resources = append(resources, eksResource)

	// EKS Node Group
	nodeResource := ResourceConfig{
		Type:       "EKS Managed Node Group",
		Name:       fmt.Sprintf("%s-node-group", appName),
		Parameters: make(map[string]string),
		Important:  true,
	}
	nodeResource.AddParameter("Instance Type", config.EKSNodeType)
	nodeResource.AddParameter("Min Nodes", fmt.Sprintf("%d", config.EKSMinNodes))
	nodeResource.AddParameter("Max Nodes", fmt.Sprintf("%d", config.EKSMaxNodes))
	nodeResource.AddParameter("Desired Nodes", fmt.Sprintf("%d", config.EKSDesiredNodes))
	nodeResource.AddParameter("Volume Size", fmt.Sprintf("%d GB", config.EKSNodeVolumeSize))
	nodeResource.AddParameter("Volume Type", "GP3 (encrypted)")
	nodeResource.AddParameter("Capacity Type", "ON_DEMAND")
	resources = append(resources, nodeResource)

	// Kubernetes Deployment
	deployResource := ResourceConfig{
		Type:       "Kubernetes Deployment",
		Name:       fmt.Sprintf("%s-deployment", appName),
		Parameters: make(map[string]string),
		Important:  true,
	}
	deployResource.AddParameter("Replicas", "2")
	deployResource.AddParameter("Container Image", detectContainerImage(analysis.Language, analysis.Framework))
	deployResource.AddParameter("Container Port", fmt.Sprintf("%d", analysis.Port))
	deployResource.AddParameter("CPU Request", "100m")
	deployResource.AddParameter("Memory Request", "128Mi")
	deployResource.AddParameter("CPU Limit", "500m")
	deployResource.AddParameter("Memory Limit", "512Mi")
	resources = append(resources, deployResource)

	// Kubernetes Service
	svcResource := ResourceConfig{
		Type:       "Load Balancer Service",
		Name:       fmt.Sprintf("%s-service", appName),
		Parameters: make(map[string]string),
		Important:  true,
	}
	svcResource.AddParameter("Type", "LoadBalancer")
	svcResource.AddParameter("Port Mapping", fmt.Sprintf("80 → %d", analysis.Port))
	svcResource.AddParameter("Protocol", "TCP")
	svcResource.AddParameter("AWS Load Balancer", "Classic ELB (auto-created)")
	resources = append(resources, svcResource)

	return resources
}

// detectRuntime determines the Lambda runtime from language and framework
func detectRuntime(language, framework string) string {
	switch language {
	case "python":
		return "python3.12"
	case "javascript", "typescript":
		return "nodejs20.x"
	case "go":
		return "provided.al2023"
	default:
		return "python3.12"
	}
}

// detectContainerImage determines the container image for EKS
func detectContainerImage(language, framework string) string {
	switch language {
	case "python":
		return "python:3.12-slim"
	case "javascript", "typescript":
		return "node:20-alpine"
	case "go":
		return "golang:1.23-alpine"
	default:
		return "nginx:alpine"
	}
}
