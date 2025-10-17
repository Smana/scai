variable "app_name" {
  description = "Application name"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "runtime" {
  description = "Lambda runtime"
  type        = string
  default     = "python3.11"
}

variable "handler" {
  description = "Lambda handler"
  type        = string
  default     = "lambda_handler.handler"
}

variable "timeout" {
  description = "Lambda timeout in seconds"
  type        = number
  default     = 30
}

variable "memory_size" {
  description = "Lambda memory size in MB"
  type        = number
  default     = 512
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
  description = "Programming language (python, javascript, typescript)"
  type        = string
}

variable "framework" {
  description = "Framework name"
  type        = string
}

variable "main_module" {
  description = "Main module name (e.g., 'app' for Python, 'index' for Node.js)"
  type        = string
  default     = "app"
}

variable "source_path" {
  description = "Local path to source code (if already packaged)"
  type        = string
  default     = ""
}

variable "env_vars" {
  description = "Environment variables"
  type        = map(string)
  default     = {}
}

variable "vpc_subnet_ids" {
  description = "VPC subnet IDs for Lambda (optional)"
  type        = list(string)
  default     = null
}

variable "vpc_security_group_ids" {
  description = "VPC security group IDs for Lambda (optional)"
  type        = list(string)
  default     = null
}
