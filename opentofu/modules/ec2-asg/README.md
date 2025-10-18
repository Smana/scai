# EC2 Auto Scaling Group Module

This module deploys a single EC2 instance within an Auto Scaling Group for automatic recovery and high availability.

## Features

- **Auto-Recovery**: Single instance with ASG for automatic replacement on failure
- **Security**: IMDSv2 required, encrypted EBS volumes, SSM access instead of SSH
- **Monitoring**: CloudWatch detailed monitoring and SSM integration
- **Networking**: Uses default VPC with configurable security groups
- **IAM**: Least-privilege IAM roles with SSM and CloudWatch permissions

## Usage

```hcl
module "app_deployment" {
  source = "./modules/ec2-asg"

  app_name         = "my-app"
  region           = "us-east-1"
  instance_type    = "t3.medium"
  volume_size      = 30
  application_port = 5000
  user_data        = file("${path.module}/userdata.sh")

  tags = {
    Project = "SCIA"
    Owner   = "Platform Team"
  }
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| app_name | Application name for resource naming | `string` | n/a | yes |
| region | AWS region | `string` | n/a | yes |
| instance_type | EC2 instance type | `string` | `"t3.micro"` | no |
| volume_size | Root volume size in GB | `number` | `20` | no |
| application_port | Application port number | `number` | `8080` | no |
| user_data | User data script | `string` | n/a | yes |
| environment | Environment name | `string` | `"production"` | no |
| enable_monitoring | Enable detailed monitoring | `bool` | `true` | no |
| health_check_grace_period | Health check grace period in seconds | `number` | `300` | no |
| tags | Additional tags | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| autoscaling_group_id | The Auto Scaling Group ID |
| autoscaling_group_name | The Auto Scaling Group name |
| security_group_id | The security group ID |
| iam_role_arn | IAM role ARN |
| application_port | Application port number |

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| aws | ~> 5.0 |

## Providers

| Name | Version |
|------|---------|
| aws | ~> 5.0 |

## Modules

| Name | Source | Version |
|------|--------|---------|
| autoscaling | terraform-aws-modules/autoscaling/aws | ~> 8.0 |
| security_group | terraform-aws-modules/security-group/aws | ~> 5.3 |

## Resources

| Name | Type |
|------|------|
| aws_iam_role.instance | resource |
| aws_iam_role_policy_attachment.ssm_managed_instance | resource |
| aws_iam_role_policy_attachment.cloudwatch_agent | resource |
| aws_iam_instance_profile.instance | resource |
| aws_ami.amazon_linux_2023 | data source |
| aws_vpc.default | data source |
| aws_subnets.default | data source |

## Security Considerations

- IMDSv2 is enforced (no IMDSv1)
- EBS volumes are encrypted by default
- Uses AWS Systems Manager for instance access (prefer over SSH)
- Security group allows SSH (port 22) but SSM is recommended
- IAM roles follow least-privilege principle

## Notes

- This module uses the default VPC. For production, consider using a custom VPC
- Single instance configuration is suitable for development/small workloads
- For high-availability production workloads, consider increasing min_size/max_size
- Health check grace period should match application startup time
