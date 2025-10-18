# Input variables for EKS cluster deployment

variable "app_name" {
  description = "Application name used for resource naming"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9-]+$", var.app_name))
    error_message = "app_name must contain only lowercase letters, numbers, and hyphens"
  }
}

variable "region" {
  description = "AWS region for EKS cluster deployment"
  type        = string
}

variable "cluster_version" {
  description = "Kubernetes version for EKS cluster"
  type        = string
  default     = "1.31"

  validation {
    condition     = can(regex("^\\d+\\.\\d+$", var.cluster_version))
    error_message = "cluster_version must be in format X.Y (e.g., 1.31)"
  }
}

variable "node_instance_type" {
  description = "EC2 instance type for EKS worker nodes"
  type        = string
  default     = "t3.medium"

  validation {
    condition     = can(regex("^[a-z][0-9][a-z]?\\.", var.node_instance_type))
    error_message = "node_instance_type must be a valid EC2 instance type"
  }
}

variable "node_min_size" {
  description = "Minimum number of nodes in the node group"
  type        = number
  default     = 1

  validation {
    condition     = var.node_min_size >= 1 && var.node_min_size <= 100
    error_message = "node_min_size must be between 1 and 100"
  }
}

variable "node_max_size" {
  description = "Maximum number of nodes in the node group"
  type        = number
  default     = 3

  validation {
    condition     = var.node_max_size >= 1 && var.node_max_size <= 100
    error_message = "node_max_size must be between 1 and 100"
  }
}

variable "node_desired_size" {
  description = "Desired number of nodes in the node group"
  type        = number
  default     = 2

  validation {
    condition     = var.node_desired_size >= 1 && var.node_desired_size <= 100
    error_message = "node_desired_size must be between 1 and 100"
  }
}

variable "node_volume_size" {
  description = "Root volume size in GB for worker nodes"
  type        = number
  default     = 30

  validation {
    condition     = var.node_volume_size >= 20 && var.node_volume_size <= 1000
    error_message = "node_volume_size must be between 20 and 1000 GB"
  }
}

variable "application_port" {
  description = "Port number for the application container"
  type        = number
  default     = 8080

  validation {
    condition     = var.application_port >= 1 && var.application_port <= 65535
    error_message = "application_port must be between 1 and 65535"
  }
}

variable "container_image" {
  description = "Container image for the application deployment"
  type        = string
}

variable "replicas" {
  description = "Number of pod replicas for the application"
  type        = number
  default     = 2

  validation {
    condition     = var.replicas >= 1 && var.replicas <= 10
    error_message = "replicas must be between 1 and 10"
  }
}

variable "environment" {
  description = "Environment name (e.g., production, staging, development)"
  type        = string
  default     = "production"

  validation {
    condition     = contains(["production", "staging", "development"], var.environment)
    error_message = "environment must be one of: production, staging, development"
  }
}

variable "enable_monitoring" {
  description = "Enable CloudWatch Container Insights for the cluster"
  type        = bool
  default     = true
}

variable "tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}
