# VPC configuration for EKS cluster

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 6.0"

  name = "${var.app_name}-vpc"
  cidr = "10.0.0.0/16"

  # Use 2 AZs for high availability
  azs             = slice(data.aws_availability_zones.available.names, 0, 2)
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24"]

  # NAT Gateway for private subnet internet access
  enable_nat_gateway   = true
  single_nat_gateway   = true # Cost-effective for non-prod
  enable_dns_hostnames = true
  enable_dns_support   = true

  # EKS requires specific tags on subnets for load balancer provisioning
  public_subnet_tags = {
    "kubernetes.io/role/elb" = "1"
  }

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = "1"
  }

  tags = merge(
    var.tags,
    {
      Name        = "${var.app_name}-vpc"
      Environment = var.environment
      ManagedBy   = "SCIA"
    }
  )
}
