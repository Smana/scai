variable "app_name" {
  description = "Application name"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "kubernetes_version" {
  description = "Kubernetes version"
  type        = string
  default     = "1.28"
}

variable "node_instance_type" {
  description = "EC2 instance type for EKS nodes"
  type        = string
  default     = "t3.small"
}

variable "node_min_size" {
  description = "Minimum number of nodes"
  type        = number
  default     = 1
}

variable "node_max_size" {
  description = "Maximum number of nodes"
  type        = number
  default     = 3
}

variable "node_desired_size" {
  description = "Desired number of nodes"
  type        = number
  default     = 2
}

variable "single_nat_gateway" {
  description = "Use single NAT gateway for cost savings"
  type        = bool
  default     = true
}

variable "namespace" {
  description = "Kubernetes namespace"
  type        = string
  default     = "default"
}

variable "replicas" {
  description = "Number of pod replicas"
  type        = number
  default     = 2
}

variable "repo_url" {
  description = "Git repository URL"
  type        = string
}

variable "app_subdir" {
  description = "Subdirectory within repository (if any)"
  type        = string
  default     = ""
}

variable "language" {
  description = "Programming language"
  type        = string
}

variable "framework" {
  description = "Framework name"
  type        = string
}

variable "start_command" {
  description = "Command to start the application"
  type        = string
}

variable "app_port" {
  description = "Application port"
  type        = number
  default     = 8080
}

variable "container_image" {
  description = "Base container image"
  type        = string
  default     = "python:3.11-slim"
}

variable "env_vars" {
  description = "Environment variables"
  type        = map(string)
  default     = {}
}

variable "cpu_request" {
  description = "CPU request"
  type        = string
  default     = "250m"
}

variable "memory_request" {
  description = "Memory request"
  type        = string
  default     = "512Mi"
}

variable "cpu_limit" {
  description = "CPU limit"
  type        = string
  default     = "500m"
}

variable "memory_limit" {
  description = "Memory limit"
  type        = string
  default     = "1Gi"
}

variable "health_path" {
  description = "Health check path"
  type        = string
  default     = "/"
}
