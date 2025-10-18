variable "app_name" {
  description = "Application name used for resource naming and tagging"
  type        = string

  validation {
    condition     = can(regex("^[a-zA-Z0-9_-]+$", var.app_name))
    error_message = "Application name must contain only alphanumeric characters, hyphens, and underscores."
  }
}

variable "region" {
  description = "AWS region where resources will be created"
  type        = string
}

variable "instance_type" {
  description = "EC2 instance type for the application"
  type        = string
  default     = "t3.micro"

  validation {
    condition     = can(regex("^[a-z][0-9][a-z]?\\.", var.instance_type))
    error_message = "Instance type must be a valid EC2 instance type (e.g., t3.micro, t3.medium)."
  }
}

variable "volume_size" {
  description = "Root volume size in GB"
  type        = number
  default     = 20

  validation {
    condition     = var.volume_size >= 8 && var.volume_size <= 1000
    error_message = "Volume size must be between 8 and 1000 GB."
  }
}

variable "application_port" {
  description = "Port on which the application listens"
  type        = number
  default     = 8080

  validation {
    condition     = var.application_port > 0 && var.application_port < 65536
    error_message = "Application port must be between 1 and 65535."
  }
}

variable "user_data" {
  description = "User data script to run on instance launch"
  type        = string
}

variable "environment" {
  description = "Environment name (e.g., production, staging, development)"
  type        = string
  default     = "production"
}

variable "enable_monitoring" {
  description = "Enable detailed CloudWatch monitoring"
  type        = bool
  default     = true
}

variable "health_check_grace_period" {
  description = "Time (in seconds) after instance comes into service before checking health"
  type        = number
  default     = 300
}

variable "tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}
