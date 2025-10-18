# Lambda API Module

This module deploys an AWS Lambda function with an optional API Gateway HTTP API endpoint for serverless application hosting.

## Features

- **Lambda Function**: Uses terraform-aws-modules/lambda for production-ready serverless deployment
- **API Gateway HTTP API**: Optional HTTP API with automatic integration to Lambda
- **Security**:
  - Least-privilege IAM roles
  - CloudWatch Logs integration
  - AWS X-Ray tracing support
- **CORS Configuration**: Configurable CORS for API Gateway
- **Monitoring**: CloudWatch Logs with configurable retention
- **Flexible Runtime Support**: Python, Node.js, Go (custom runtime)

## Usage

```hcl
module "lambda_app" {
  source = "./opentofu/modules/lambda-api"

  app_name             = "my-api"
  region               = "us-east-1"
  runtime              = "python3.12"
  handler              = "app.handler"
  lambda_package_path  = "${path.module}/lambda.zip"

  # Function configuration
  timeout     = 30
  memory_size = 512

  # Environment variables
  environment_variables = {
    DEBUG = "false"
  }

  # API Gateway
  enable_api_gateway = true
  cors_allow_origins = ["*"]

  # Monitoring
  enable_xray_tracing  = true
  log_retention_days   = 7

  tags = {
    Project = "SCIA"
  }
}
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| aws | ~> 5.0 |
| null | ~> 3.0 |

## Providers

| Name | Version |
|------|---------|
| aws | ~> 5.0 |
| null | ~> 3.0 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| app_name | Application name used for resource naming | `string` | n/a | yes |
| region | AWS region for Lambda deployment | `string` | n/a | yes |
| runtime | Lambda runtime (e.g., python3.12, nodejs20.x) | `string` | n/a | yes |
| handler | Lambda function handler (e.g., app.handler) | `string` | n/a | yes |
| lambda_package_path | Path to the Lambda deployment package (.zip) | `string` | n/a | yes |
| timeout | Lambda function timeout in seconds | `number` | `30` | no |
| memory_size | Lambda function memory size in MB | `number` | `512` | no |
| reserved_concurrent_executions | Reserved concurrent executions (0 = unreserved) | `number` | `0` | no |
| environment_variables | Environment variables for the function | `map(string)` | `{}` | no |
| enable_xray_tracing | Enable AWS X-Ray tracing | `bool` | `true` | no |
| log_retention_days | CloudWatch Logs retention period in days | `number` | `7` | no |
| enable_api_gateway | Create API Gateway HTTP API | `bool` | `true` | no |
| cors_allow_origins | CORS allowed origins for API Gateway | `list(string)` | `["*"]` | no |
| tags | Additional tags to apply to all resources | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| function_name | Lambda function name |
| function_arn | Lambda function ARN |
| function_invoke_arn | Lambda function invoke ARN |
| function_qualified_arn | Lambda function ARN with version |
| function_version | Latest published version |
| role_arn | IAM role ARN for the function |
| role_name | IAM role name |
| log_group_name | CloudWatch Logs group name |
| log_group_arn | CloudWatch Logs group ARN |
| api_id | API Gateway ID (null if disabled) |
| api_arn | API Gateway ARN (null if disabled) |
| api_endpoint | API Gateway endpoint URL (null if disabled) |
| api_invoke_url | API Gateway invoke URL (null if disabled) |
| api_execution_arn | API Gateway execution ARN (null if disabled) |
| test_command_cli | AWS CLI test command |
| test_command_curl | Curl test command (null if API disabled) |

## Lambda Package Requirements

The module expects a pre-built Lambda deployment package (`.zip` file). The package should contain:

### Python Example
```
lambda.zip
├── app.py           # Handler file
├── requirements.txt # (optional)
└── dependencies/    # (installed packages)
```

Handler in `app.py`:
```python
def handler(event, context):
    return {
        "statusCode": 200,
        "body": "Hello from Lambda!"
    }
```

### Node.js Example
```
lambda.zip
├── index.js         # Handler file
├── package.json     # (optional)
└── node_modules/    # (installed packages)
```

Handler in `index.js`:
```javascript
exports.handler = async (event) => {
    return {
        statusCode: 200,
        body: "Hello from Lambda!"
    };
};
```

## API Gateway Integration

When `enable_api_gateway = true`, the module creates:

- **HTTP API Gateway**: Low-latency, cost-effective API
- **Routes**: `ANY /` and `ANY /{proxy+}` for catch-all routing
- **Integration**: Direct Lambda proxy integration with payload format 2.0
- **CORS**: Configured with `cors_allow_origins`
- **Permissions**: Automatic Lambda permission for API Gateway invocation

### Testing the API

```bash
# Get the API URL from Terraform output
terraform output api_invoke_url

# Test with curl
curl https://<api-id>.execute-api.<region>.amazonaws.com/

# Test with AWS CLI (direct Lambda invocation)
aws lambda invoke --function-name my-api --region us-east-1 response.json
cat response.json
```

## Monitoring and Logging

### CloudWatch Logs

All function logs are sent to CloudWatch Logs:

```bash
# View logs
aws logs tail /aws/lambda/my-api --follow

# Search logs
aws logs filter-log-events \
  --log-group-name /aws/lambda/my-api \
  --filter-pattern "ERROR"
```

### X-Ray Tracing

When `enable_xray_tracing = true`, traces are available in the X-Ray console for performance analysis and debugging.

## Security Considerations

- **IAM Permissions**: Function has minimal permissions (logs only by default)
- **API Gateway**: Uses HTTP API (not REST API) for better security and lower cost
- **Environment Variables**: Injected securely, avoid secrets (use Secrets Manager instead)
- **Package**: Ensure deployment package doesn't contain sensitive files (.env, credentials)

## Cost Optimization

- **Memory Size**: Right-size memory (128-10240 MB) for cost vs. performance
- **Timeout**: Set appropriate timeout to avoid unnecessary charges
- **Reserved Concurrency**: Only set if you need guaranteed capacity
- **Log Retention**: Use shorter retention (7 days) for non-production

## Troubleshooting

**Function not found:**
```bash
aws lambda get-function --function-name my-api --region us-east-1
```

**Permission errors:**
```bash
aws lambda get-policy --function-name my-api --region us-east-1
```

**API Gateway 5xx errors:**
- Check Lambda CloudWatch Logs for errors
- Verify handler name matches the actual function
- Ensure package structure is correct

**Cold start issues:**
- Increase memory size (more CPU allocated)
- Consider provisioned concurrency for critical functions
- Optimize package size (remove unnecessary dependencies)

## References

- [AWS Lambda Best Practices](https://docs.aws.amazon.com/lambda/latest/dg/best-practices.html)
- [terraform-aws-modules/lambda](https://registry.terraform.io/modules/terraform-aws-modules/lambda/aws/latest)
- [terraform-aws-modules/apigateway-v2](https://registry.terraform.io/modules/terraform-aws-modules/apigateway-v2/aws/latest)
- [Lambda Runtimes](https://docs.aws.amazon.com/lambda/latest/dg/lambda-runtimes.html)
