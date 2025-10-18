# Lambda function using terraform-aws-modules/lambda

module "lambda_function" {
  source  = "terraform-aws-modules/lambda/aws"
  version = "~> 8.0"

  function_name = var.app_name
  description   = "Lambda function for ${var.app_name} deployed by SCIA"
  handler       = var.handler
  runtime       = var.runtime

  # Package configuration
  create_package         = false
  local_existing_package = var.lambda_package_path

  # Function configuration
  timeout     = var.timeout
  memory_size = var.memory_size

  # Reserved concurrency
  reserved_concurrent_executions = var.reserved_concurrent_executions > 0 ? var.reserved_concurrent_executions : null

  # Environment variables
  environment_variables = merge(
    var.environment_variables,
    {
      APP_NAME = var.app_name
      REGION   = var.region
    }
  )

  # CloudWatch Logs configuration
  cloudwatch_logs_retention_in_days = var.log_retention_days
  attach_cloudwatch_logs_policy     = true

  # X-Ray tracing
  tracing_mode = var.enable_xray_tracing ? "Active" : "PassThrough"

  # Permissions
  attach_network_policy = false
  attach_policy_json    = true
  policy_json = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:${var.region}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${var.app_name}:*"
      }
    ]
  })

  # Tags
  tags = merge(
    var.tags,
    {
      Name      = var.app_name
      ManagedBy = "SCIA"
    }
  )
}
