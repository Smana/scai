# API Gateway HTTP API for Lambda function

module "api_gateway" {
  source  = "terraform-aws-modules/apigateway-v2/aws"
  version = "~> 5.0"

  count = var.enable_api_gateway ? 1 : 0

  name          = "${var.app_name}-api"
  description   = "HTTP API Gateway for ${var.app_name}"
  protocol_type = "HTTP"

  # Disable custom domain features
  create_domain_name    = false
  create_certificate    = false
  create_domain_records = false

  # CORS configuration
  cors_configuration = {
    allow_headers = ["*"]
    allow_methods = ["*"]
    allow_origins = var.cors_allow_origins
  }

  # Routes configuration
  routes = {
    "ANY /" = {
      integration = {
        uri                    = module.lambda_function.lambda_function_arn
        payload_format_version = "2.0"
        timeout_milliseconds   = var.timeout * 1000
      }
    }

    "ANY /{proxy+}" = {
      integration = {
        uri                    = module.lambda_function.lambda_function_arn
        payload_format_version = "2.0"
        timeout_milliseconds   = var.timeout * 1000
      }
    }
  }

  tags = merge(
    var.tags,
    {
      Name      = "${var.app_name}-api"
      ManagedBy = "SCIA"
    }
  )
}

# Lambda permission for API Gateway to invoke the function
resource "aws_lambda_permission" "api_gw" {
  count = var.enable_api_gateway ? 1 : 0

  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_function.lambda_function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${module.api_gateway[0].api_execution_arn}/*/*"
}
