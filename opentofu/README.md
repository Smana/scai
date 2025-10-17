# SCIA OpenTofu Modules

This directory contains OpenTofu modules for deploying applications using terraform-aws-modules.

## Structure

```
opentofu/
├── modules/
│   ├── ec2/       # EC2 deployment with Auto Scaling Group
│   ├── eks/       # Kubernetes deployment on AWS EKS
│   └── lambda/    # Serverless deployment with AWS Lambda
└── README.md
```

## Modules Overview

### EC2 Module (`modules/ec2/`)

Deploys applications on EC2 instances using Auto Scaling Groups for reliability.

**Features**:
- Uses `terraform-aws-modules/autoscaling/aws` for single-instance ASG
- Uses `terraform-aws-modules/vpc/aws` for networking
- Uses `terraform-aws-modules/security-group/aws` for security
- Automatic instance recovery on failure
- User-data script for application bootstrapping
- Support for Python, Node.js, and Go applications

**Usage**:
```hcl
module "app_deployment" {
  source = "./modules/ec2"

  app_name      = "my-app"
  aws_region    = "us-east-1"
  instance_type = "t3.micro"
  app_port      = 5000
  repo_url      = "https://github.com/user/repo"
  language      = "python"
  framework     = "flask"
  start_command = "python app.py"
}
```

### EKS Module (`modules/eks/`)

Deploys containerized applications on AWS EKS with managed node groups.

**Features**:
- Uses `terraform-aws-modules/eks/aws` for cluster management
- Uses `terraform-aws-modules/vpc/aws` for networking
- Managed node groups with autoscaling
- Kubernetes Deployment with health checks
- LoadBalancer Service for external access
- Init container for Git repository cloning

**Usage**:
```hcl
module "app_deployment" {
  source = "./modules/eks"

  app_name           = "my-app"
  aws_region         = "us-east-1"
  kubernetes_version = "1.28"
  node_instance_type = "t3.small"
  node_desired_size  = 2
  replicas           = 2
  repo_url           = "https://github.com/user/repo"
  language           = "python"
  framework          = "flask"
  start_command      = "gunicorn app:app"
  app_port           = 5000
}
```

### Lambda Module (`modules/lambda/`)

Deploys serverless applications using AWS Lambda with API Gateway.

**Features**:
- Uses `terraform-aws-modules/lambda/aws` for function management
- Automatic code packaging for Python and Node.js
- API Gateway HTTP API integration
- CloudWatch Logs integration
- Support for ASGI/WSGI frameworks via adapters

**Usage**:
```hcl
module "app_deployment" {
  source = "./modules/lambda"

  app_name  = "my-app"
  aws_region = "us-east-1"
  runtime   = "python3.11"
  timeout   = 30
  memory_size = 512
  repo_url  = "https://github.com/user/repo"
  language  = "python"
  framework = "fastapi"
}
```

## Requirements

- OpenTofu >= 1.6 (or Terraform >= 1.6)
- AWS credentials configured
- AWS provider ~> 5.0

## OpenTofu vs Terraform

These modules are compatible with both OpenTofu and Terraform. SCIA uses **OpenTofu** by default as it's fully open source.

To use Terraform instead, configure:
```yaml
# ~/.scia.yaml
terraform:
  bin: terraform  # or tofu
```

## Module Development

### Adding a New Module

1. Create directory: `opentofu/modules/<module-name>/`
2. Create `main.tf` with terraform-aws-modules integration
3. Create `variables.tf` with input variables
4. Document usage in this README
5. Update SCIA code to reference new module

### Best Practices

- **Use terraform-aws-modules**: Always use community modules instead of raw resources
- **Enable auto-recovery**: Especially for single-instance deployments
- **Tag everything**: Use consistent tagging (Application, ManagedBy, Name)
- **Security first**: Minimal security group rules, non-root users, HTTPS
- **Cost-aware**: Use single NAT gateway, t3.micro instances where appropriate

## Testing

Test modules locally:

```bash
cd opentofu/modules/ec2
tofu init
tofu plan -var="app_name=test" -var="repo_url=https://github.com/test/repo" -var="language=python" -var="framework=flask" -var="start_command=python app.py"
```

## Contributing

When modifying modules:
1. Test with sample applications
2. Ensure backward compatibility
3. Update documentation
4. Follow OpenTofu/Terraform best practices
