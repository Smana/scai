# EKS cluster using terraform-aws-modules/eks

module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 21.0"

  name               = "${var.app_name}-eks"
  kubernetes_version = var.cluster_version

  # Cluster endpoint access configuration
  endpoint_public_access = true

  # Enable cluster creator admin permissions
  enable_cluster_creator_admin_permissions = true

  # VPC and subnet configuration
  vpc_id                   = module.vpc.vpc_id
  subnet_ids               = module.vpc.private_subnets
  control_plane_subnet_ids = module.vpc.private_subnets

  # CloudWatch logging for audit and diagnostics
  enabled_log_types = [
    "api",
    "audit",
    "authenticator",
    "controllerManager",
    "scheduler"
  ]

  # EKS Managed Node Group
  eks_managed_node_groups = {
    default = {
      name = "${var.app_name}-node-group"

      instance_types = [var.node_instance_type]
      capacity_type  = "ON_DEMAND"

      min_size     = var.node_min_size
      max_size     = var.node_max_size
      desired_size = var.node_desired_size

      # EBS volume configuration
      block_device_mappings = {
        xvda = {
          device_name = "/dev/xvda"
          ebs = {
            volume_size           = var.node_volume_size
            volume_type           = "gp3"
            delete_on_termination = true
            encrypted             = true
          }
        }
      }

      # Instance metadata options (IMDSv2)
      metadata_options = {
        http_endpoint               = "enabled"
        http_tokens                 = "required"
        http_put_response_hop_limit = 1
        instance_metadata_tags      = "enabled"
      }

      tags = merge(
        var.tags,
        {
          Name        = "${var.app_name}-node"
          Environment = var.environment
          ManagedBy   = "SCIA"
        }
      )
    }
  }

  # Cluster add-ons
  addons = {
    # VPC CNI for pod networking
    vpc-cni = {
      most_recent = true
    }
    # CoreDNS for service discovery
    coredns = {
      most_recent = true
    }
    # kube-proxy for service load balancing
    kube-proxy = {
      most_recent = true
    }
    # EBS CSI driver for persistent volumes
    aws-ebs-csi-driver = {
      most_recent = true
    }
  }

  tags = merge(
    var.tags,
    {
      Name        = "${var.app_name}-eks"
      Environment = var.environment
      ManagedBy   = "SCIA"
    }
  )
}
