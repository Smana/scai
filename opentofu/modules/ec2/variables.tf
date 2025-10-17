variable "app_name" {
  description = "Application name"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.micro"
}

variable "app_port" {
  description = "Application port"
  type        = number
  default     = 8080
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
  description = "Programming language (python, javascript, typescript, go)"
  type        = string
}

variable "framework" {
  description = "Framework name (flask, django, express, etc.)"
  type        = string
}

variable "start_command" {
  description = "Command to start the application"
  type        = string
}

variable "env_vars" {
  description = "Environment variables for the application"
  type        = map(string)
  default     = {}
}

variable "python_version" {
  description = "Python version to install"
  type        = string
  default     = "3.11"
}

variable "nodejs_version" {
  description = "Node.js major version"
  type        = string
  default     = "20"
}

variable "go_version" {
  description = "Go version to install"
  type        = string
  default     = "1.21.5"
}

variable "create_iam_role" {
  description = "Whether to create IAM instance profile"
  type        = bool
  default     = true
}
