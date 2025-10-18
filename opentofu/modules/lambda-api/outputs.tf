# Output values from Lambda API deployment

# Lambda function information
output "function_name" {
  description = "Lambda function name"
  value       = module.lambda_function.lambda_function_name
}

output "function_arn" {
  description = "Lambda function ARN"
  value       = module.lambda_function.lambda_function_arn
}

output "function_invoke_arn" {
  description = "Lambda function invoke ARN for API Gateway integration"
  value       = module.lambda_function.lambda_function_invoke_arn
}

output "function_qualified_arn" {
  description = "Lambda function ARN with version qualifier"
  value       = module.lambda_function.lambda_function_qualified_arn
}

output "function_version" {
  description = "Latest published version of the Lambda function"
  value       = module.lambda_function.lambda_function_version
}

output "role_arn" {
  description = "IAM role ARN for the Lambda function"
  value       = module.lambda_function.lambda_role_arn
}

output "role_name" {
  description = "IAM role name for the Lambda function"
  value       = module.lambda_function.lambda_role_name
}

# CloudWatch Logs
output "log_group_name" {
  description = "CloudWatch Logs group name for the function"
  value       = module.lambda_function.lambda_cloudwatch_log_group_name
}

output "log_group_arn" {
  description = "CloudWatch Logs group ARN"
  value       = module.lambda_function.lambda_cloudwatch_log_group_arn
}

# API Gateway information (conditional)
output "api_id" {
  description = "API Gateway ID (null if API Gateway is disabled)"
  value       = var.enable_api_gateway ? module.api_gateway[0].api_id : null
}

output "api_arn" {
  description = "API Gateway ARN (null if API Gateway is disabled)"
  value       = var.enable_api_gateway ? module.api_gateway[0].api_arn : null
}

output "api_endpoint" {
  description = "API Gateway endpoint URL (null if API Gateway is disabled)"
  value       = var.enable_api_gateway ? module.api_gateway[0].api_endpoint : null
}

output "api_invoke_url" {
  description = "API Gateway invoke URL (null if API Gateway is disabled)"
  value       = var.enable_api_gateway ? "${module.api_gateway[0].api_endpoint}/" : null
}

output "api_execution_arn" {
  description = "API Gateway execution ARN for permissions (null if API Gateway is disabled)"
  value       = var.enable_api_gateway ? module.api_gateway[0].api_execution_arn : null
}

# Testing commands
output "test_command_cli" {
  description = "AWS CLI command to test the Lambda function directly"
  value       = "aws lambda invoke --function-name ${module.lambda_function.lambda_function_name} --region ${var.region} response.json"
}

output "test_command_curl" {
  description = "Curl command to test the API endpoint (null if API Gateway is disabled)"
  value       = var.enable_api_gateway ? "curl ${module.api_gateway[0].api_endpoint}/" : null
}
