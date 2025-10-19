# SCAI OpenTofu Modules

This directory contains reusable OpenTofu/Terraform modules for deploying applications to AWS.

## Module Structure

```
opentofu/
├── modules/
│   ├── ec2-asg/          # VM deployment with Auto Scaling Group
│   ├── eks-cluster/      # Kubernetes cluster deployment (coming soon)
│   └── lambda-api/       # Serverless deployment (coming soon)
├── .tflint.hcl          # TFLint configuration
└── README.md            # This file
```

## Available Modules

### EC2 Auto Scaling Group (`ec2-asg`)

Deploys a single EC2 instance within an Auto Scaling Group for automatic recovery.

**Features:**
- Auto-recovery on instance failure
- IMDSv2 enforced for enhanced security
- SSM access for instance management
- Encrypted EBS volumes
- CloudWatch monitoring

**Usage:**
```hcl
module "app" {
  source = "./opentofu/modules/ec2-asg"

  app_name         = "my-flask-app"
  region           = "us-east-1"
  instance_type    = "t3.medium"
  volume_size      = 30
  application_port = 5000
  user_data        = file("${path.module}/userdata.sh")
}
```

See [modules/ec2-asg/README.md](modules/ec2-asg/README.md) for full documentation.

### EKS Cluster (`eks-cluster`)

**Status:** Coming soon

Deploys an EKS cluster with managed node groups and application deployment.

### Lambda API (`lambda-api`)

**Status:** Coming soon

Deploys a serverless application using AWS Lambda and API Gateway.

## Development

### Prerequisites

- [OpenTofu](https://opentofu.org/) >= 1.0 or [Terraform](https://www.terraform.io/) >= 1.0
- [TFLint](https://github.com/terraform-linters/tflint) >= 0.50
- [pre-commit](https://pre-commit.com/) (optional but recommended)

### Running Checks Locally

```bash
# Format all Terraform files
terraform fmt -recursive .

# Validate configuration
terraform validate

# Run TFLint
tflint --config=.tflint.hcl --recursive

# Run all pre-commit hooks
pre-commit run --all-files
```

### Module Development Guidelines

1. **File Organization**: Follow the standard module structure
   - `main.tf` - Primary resources
   - `variables.tf` - Input variables with validation
   - `outputs.tf` - Module outputs
   - `versions.tf` - Provider version constraints
   - `data.tf` - Data sources
   - `README.md` - Module documentation

2. **Naming Conventions**:
   - Use `snake_case` for all resource names
   - Prefix resources with `${var.app_name}-`
   - Tag all resources with `Name`, `Environment`, `ManagedBy`

3. **Security Best Practices**:
   - Enable encryption by default
   - Use IMDSv2 for EC2 instances
   - Follow least-privilege IAM policies
   - Avoid hardcoded credentials

4. **Documentation**:
   - Document all variables and outputs
   - Provide usage examples
   - Include security considerations
   - List all requirements and dependencies

## CI/CD Integration

These modules are validated in CI using:
- `terraform fmt` - Code formatting
- `terraform validate` - Configuration validation
- `tflint` - Linting and best practices

See `.github/workflows/` for CI configuration.

## Best Practices

This module structure follows recommendations from:
- [Terraform Best Practices](https://www.terraform-best-practices.com/)
- [AWS Prescriptive Guidance](https://docs.aws.amazon.com/prescriptive-guidance/latest/terraform-aws-provider-best-practices/)
- [HashiCorp Standard Module Structure](https://developer.hashicorp.com/terraform/language/modules/develop/structure)

## License

MIT License - See [LICENSE](../LICENSE) for details.
