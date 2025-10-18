# Input variables for Lambda API deployment

variable "app_name" {
  description = "Application name used for resource naming"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9-_]+$", var.app_name))
    error_message = "app_name must contain only lowercase letters, numbers, hyphens, and underscores"
  }
}

variable "region" {
  description = "AWS region for Lambda deployment"
  type        = string
}

variable "runtime" {
  description = "Lambda runtime (e.g., python3.12, nodejs20.x, provided.al2023)"
  type        = string

  validation {
    condition     = contains(["python3.12", "python3.11", "nodejs20.x", "nodejs18.x", "provided.al2023"], var.runtime)
    error_message = "runtime must be a supported Lambda runtime"
  }
}

variable "handler" {
  description = "Lambda function handler (e.g., app.handler, index.handler)"
  type        = string
}

variable "lambda_package_path" {
  description = "Path to the Lambda deployment package (.zip file)"
  type        = string
}

variable "timeout" {
  description = "Lambda function timeout in seconds"
  type        = number
  default     = 30

  validation {
    condition     = var.timeout >= 1 && var.timeout <= 900
    error_message = "timeout must be between 1 and 900 seconds"
  }
}

variable "memory_size" {
  description = "Lambda function memory size in MB"
  type        = number
  default     = 512

  validation {
    condition     = var.memory_size >= 128 && var.memory_size <= 10240
    error_message = "memory_size must be between 128 and 10240 MB"
  }
}

variable "reserved_concurrent_executions" {
  description = "Reserved concurrent executions for the function (0 = unreserved)"
  type        = number
  default     = 0

  validation {
    condition     = var.reserved_concurrent_executions >= 0
    error_message = "reserved_concurrent_executions must be >= 0"
  }
}

variable "environment_variables" {
  description = "Environment variables for the Lambda function"
  type        = map(string)
  default     = {}
}

variable "enable_xray_tracing" {
  description = "Enable AWS X-Ray tracing for the function"
  type        = bool
  default     = true
}

variable "log_retention_days" {
  description = "CloudWatch Logs retention period in days"
  type        = number
  default     = 7

  validation {
    condition     = contains([1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, 3653], var.log_retention_days)
    error_message = "log_retention_days must be a valid CloudWatch Logs retention value"
  }
}

variable "enable_api_gateway" {
  description = "Create API Gateway HTTP API for the function"
  type        = bool
  default     = true
}

variable "cors_allow_origins" {
  description = "CORS allowed origins for API Gateway"
  type        = list(string)
  default     = ["*"]
}

variable "tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}
