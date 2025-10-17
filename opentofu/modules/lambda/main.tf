terraform {
  required_version = ">= 1.6"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# Lambda Module from terraform-aws-modules
module "lambda_function" {
  source  = "terraform-aws-modules/lambda/aws"
  version = "~> 7.0"

  function_name = var.app_name
  description   = "Serverless deployment for ${var.app_name}"
  handler       = var.handler
  runtime       = var.runtime

  # Source code configuration
  create_package = true
  source_path    = var.source_path != "" ? var.source_path : null

  # If source_path not provided, use local provisioner to clone and package
  ignore_source_code_hash = var.source_path == ""

  # Lambda configuration
  timeout     = var.timeout
  memory_size = var.memory_size

  # Environment variables
  environment_variables = merge(
    var.env_vars,
    {
      IS_LAMBDA = "true"
    }
  )

  # Logging
  cloudwatch_logs_retention_in_days = 7

  # VPC configuration (optional)
  attach_network_policy = var.vpc_subnet_ids != null
  vpc_subnet_ids        = var.vpc_subnet_ids
  vpc_security_group_ids = var.vpc_security_group_ids

  # Permissions
  attach_policy_statements = true
  policy_statements = {
    logs = {
      effect = "Allow"
      actions = [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ]
      resources = ["arn:aws:logs:*:*:*"]
    }
  }

  # Tags
  tags = {
    Name        = var.app_name
    Application = var.app_name
    ManagedBy   = "scia"
  }
}

# API Gateway HTTP API
resource "aws_apigatewayv2_api" "lambda" {
  name          = "${var.app_name}-api"
  protocol_type = "HTTP"
  description   = "API Gateway for ${var.app_name}"

  cors_configuration {
    allow_origins = ["*"]
    allow_methods = ["*"]
    allow_headers = ["*"]
  }

  tags = {
    Name        = "${var.app_name}-api"
    Application = var.app_name
    ManagedBy   = "scia"
  }
}

# API Gateway Stage
resource "aws_apigatewayv2_stage" "lambda" {
  api_id      = aws_apigatewayv2_api.lambda.id
  name        = "$default"
  auto_deploy = true

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_gateway.arn
    format = jsonencode({
      requestId      = "$context.requestId"
      ip             = "$context.identity.sourceIp"
      requestTime    = "$context.requestTime"
      httpMethod     = "$context.httpMethod"
      routeKey       = "$context.routeKey"
      status         = "$context.status"
      protocol       = "$context.protocol"
      responseLength = "$context.responseLength"
    })
  }

  tags = {
    Name        = "${var.app_name}-stage"
    Application = var.app_name
    ManagedBy   = "scia"
  }
}

# CloudWatch Log Group for API Gateway
resource "aws_cloudwatch_log_group" "api_gateway" {
  name              = "/aws/apigateway/${var.app_name}"
  retention_in_days = 7

  tags = {
    Name        = "${var.app_name}-api-logs"
    Application = var.app_name
    ManagedBy   = "scia"
  }
}

# API Gateway Integration with Lambda
resource "aws_apigatewayv2_integration" "lambda" {
  api_id = aws_apigatewayv2_api.lambda.id

  integration_uri    = module.lambda_function.lambda_function_invoke_arn
  integration_type   = "AWS_PROXY"
  integration_method = "POST"

  payload_format_version = "2.0"
}

# API Gateway Route (catch-all)
resource "aws_apigatewayv2_route" "lambda" {
  api_id    = aws_apigatewayv2_api.lambda.id
  route_key = "$default"
  target    = "integrations/${aws_apigatewayv2_integration.lambda.id}"
}

# Lambda Permission for API Gateway
resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_function.lambda_function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.lambda.execution_arn}/*/*"
}

# Null resource for custom packaging (when source_path not provided)
resource "null_resource" "package_lambda" {
  count = var.source_path == "" ? 1 : 0

  provisioner "local-exec" {
    command = <<-EOT
      #!/bin/bash
      set -e

      # Create temp directory
      TMPDIR=$(mktemp -d)
      cd $TMPDIR

      # Clone repository
      git clone ${var.repo_url} app
      cd app
      ${var.app_subdir != "" ? "cd ${var.app_subdir}" : ""}

      # Install dependencies and create handler based on language
      ${var.language == "python" ? <<-PYTHON
      # Python Lambda handler
      pip install -r requirements.txt -t . 2>/dev/null || true

      # Create Lambda handler wrapper
      cat > lambda_handler.py << 'HANDLER'
import os
os.environ['IS_LAMBDA'] = 'true'

# Import the application
try:
    from ${var.main_module} import app
except ImportError:
    import ${var.main_module}
    app = ${var.main_module}.app

# AWS Lambda handler using Mangum (ASGI/WSGI adapter)
def handler(event, context):
    from mangum import Mangum
    handler_func = Mangum(app, lifespan="off")
    return handler_func(event, context)
HANDLER

      # Install mangum adapter
      pip install mangum -t . 2>/dev/null || true
      PYTHON
      : ""}

      ${var.language == "javascript" || var.language == "typescript" ? <<-NODEJS
      # Node.js Lambda handler
      npm install 2>/dev/null || true

      # Create Lambda handler wrapper
      cat > lambda_handler.js << 'HANDLER'
const serverless = require('serverless-http');
const app = require('./${var.main_module}');

module.exports.handler = serverless(app);
HANDLER

      # Install serverless-http
      npm install serverless-http 2>/dev/null || true
      NODEJS
      : ""}

      # Package everything
      zip -r ${var.app_name}_lambda.zip . -x "*.git*" "*node_modules/.cache*" "*.pyc" "__pycache__/*"

      # Move to expected location
      mkdir -p ${path.module}/builds
      mv ${var.app_name}_lambda.zip ${path.module}/builds/

      echo "Lambda package created: ${path.module}/builds/${var.app_name}_lambda.zip"
    EOT
  }

  triggers = {
    repo_url  = var.repo_url
    timestamp = timestamp()
  }
}

# Outputs
output "lambda_function_arn" {
  description = "Lambda function ARN"
  value       = module.lambda_function.lambda_function_arn
}

output "lambda_function_name" {
  description = "Lambda function name"
  value       = module.lambda_function.lambda_function_name
}

output "lambda_function_invoke_arn" {
  description = "Lambda function invoke ARN"
  value       = module.lambda_function.lambda_function_invoke_arn
}

output "api_gateway_endpoint" {
  description = "API Gateway endpoint URL"
  value       = aws_apigatewayv2_api.lambda.api_endpoint
}

output "public_url" {
  description = "Public URL of the application"
  value       = aws_apigatewayv2_api.lambda.api_endpoint
}

output "invoke_url" {
  description = "Full invoke URL with stage"
  value       = "${aws_apigatewayv2_api.lambda.api_endpoint}/"
}

output "cloudwatch_log_group" {
  description = "CloudWatch log group name"
  value       = module.lambda_function.lambda_cloudwatch_log_group_name
}
