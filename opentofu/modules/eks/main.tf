terraform {
  required_version = ">= 1.6"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# Data sources
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_caller_identity" "current" {}

# VPC Module for EKS
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "${var.app_name}-eks-vpc"
  cidr = "10.0.0.0/16"

  azs             = slice(data.aws_availability_zones.available.names, 0, 3)
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]

  enable_nat_gateway   = true
  single_nat_gateway   = var.single_nat_gateway
  enable_dns_hostnames = true
  enable_dns_support   = true

  # Kubernetes tags
  public_subnet_tags = {
    "kubernetes.io/role/elb" = "1"
  }

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = "1"
  }

  tags = {
    Name        = "${var.app_name}-eks-vpc"
    Application = var.app_name
    ManagedBy   = "scia"
  }
}

# EKS Cluster Module
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.0"

  cluster_name    = "${var.app_name}-eks"
  cluster_version = var.kubernetes_version

  # Networking
  vpc_id                   = module.vpc.vpc_id
  subnet_ids               = module.vpc.private_subnets
  control_plane_subnet_ids = module.vpc.public_subnets

  # Cluster endpoint access
  cluster_endpoint_public_access  = true
  cluster_endpoint_private_access = true

  # Cluster addons
  cluster_addons = {
    coredns = {
      most_recent = true
    }
    kube-proxy = {
      most_recent = true
    }
    vpc-cni = {
      most_recent = true
    }
  }

  # Managed node groups
  eks_managed_node_groups = {
    default = {
      name = "${var.app_name}-nodes"

      min_size     = var.node_min_size
      max_size     = var.node_max_size
      desired_size = var.node_desired_size

      instance_types = [var.node_instance_type]
      capacity_type  = "ON_DEMAND"

      # Node group configuration
      labels = {
        Environment = "production"
        Application = var.app_name
      }

      tags = {
        Name        = "${var.app_name}-node-group"
        Application = var.app_name
        ManagedBy   = "scia"
      }
    }
  }

  # Enable IRSA (IAM Roles for Service Accounts)
  enable_irsa = true

  tags = {
    Name        = "${var.app_name}-eks"
    Application = var.app_name
    ManagedBy   = "scia"
  }
}

# Configure Kubernetes provider
provider "kubernetes" {
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)

  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args = [
      "eks",
      "get-token",
      "--cluster-name",
      module.eks.cluster_name,
      "--region",
      var.aws_region
    ]
  }
}

# Kubernetes resources
resource "kubernetes_namespace" "app" {
  metadata {
    name = var.namespace

    labels = {
      app       = var.app_name
      managedBy = "scia"
    }
  }

  depends_on = [module.eks]
}

resource "kubernetes_deployment" "app" {
  metadata {
    name      = var.app_name
    namespace = kubernetes_namespace.app.metadata[0].name

    labels = {
      app = var.app_name
    }
  }

  spec {
    replicas = var.replicas

    selector {
      match_labels = {
        app = var.app_name
      }
    }

    template {
      metadata {
        labels = {
          app = var.app_name
        }
      }

      spec {
        # Init container to clone repository
        init_container {
          name  = "git-clone"
          image = "alpine/git:latest"

          command = ["sh", "-c"]
          args = [
            "git clone ${var.repo_url} /app${var.app_subdir != "" ? " && cd /app/${var.app_subdir}" : ""}"
          ]

          volume_mount {
            name       = "app-code"
            mount_path = "/app"
          }
        }

        # Main application container
        container {
          name  = "app"
          image = var.container_image

          command = ["sh", "-c"]
          args = [
            <<-EOT
            cd /app${var.app_subdir != "" ? "/${var.app_subdir}" : ""} && \
            ${var.language == "python" ? "pip install -r requirements.txt && " : ""}
            ${var.language == "javascript" || var.language == "typescript" ? "npm install && " : ""}
            ${var.start_command}
            EOT
          ]

          port {
            container_port = var.app_port
            protocol       = "TCP"
          }

          # Environment variables
          dynamic "env" {
            for_each = var.env_vars
            content {
              name  = env.key
              value = env.value
            }
          }

          # Resource limits
          resources {
            requests = {
              cpu    = var.cpu_request
              memory = var.memory_request
            }
            limits = {
              cpu    = var.cpu_limit
              memory = var.memory_limit
            }
          }

          # Health checks
          liveness_probe {
            http_get {
              path = var.health_path
              port = var.app_port
            }
            initial_delay_seconds = 30
            period_seconds        = 10
          }

          readiness_probe {
            http_get {
              path = var.health_path
              port = var.app_port
            }
            initial_delay_seconds = 10
            period_seconds        = 5
          }

          volume_mount {
            name       = "app-code"
            mount_path = "/app"
          }
        }

        volume {
          name = "app-code"
          empty_dir {}
        }
      }
    }
  }
}

# Service (LoadBalancer)
resource "kubernetes_service" "app" {
  metadata {
    name      = var.app_name
    namespace = kubernetes_namespace.app.metadata[0].name
  }

  spec {
    selector = {
      app = var.app_name
    }

    port {
      port        = var.app_port
      target_port = var.app_port
      protocol    = "TCP"
    }

    type = "LoadBalancer"
  }
}

# Outputs
output "cluster_endpoint" {
  description = "EKS cluster endpoint"
  value       = module.eks.cluster_endpoint
}

output "cluster_name" {
  description = "EKS cluster name"
  value       = module.eks.cluster_name
}

output "cluster_security_group_id" {
  description = "Security group ID attached to the EKS cluster"
  value       = module.eks.cluster_security_group_id
}

output "cluster_certificate_authority_data" {
  description = "Base64 encoded certificate data"
  value       = module.eks.cluster_certificate_authority_data
  sensitive   = true
}

output "namespace" {
  description = "Kubernetes namespace"
  value       = kubernetes_namespace.app.metadata[0].name
}

output "service_name" {
  description = "Kubernetes service name"
  value       = kubernetes_service.app.metadata[0].name
}

output "load_balancer_hostname" {
  description = "Load balancer hostname"
  value       = try(kubernetes_service.app.status[0].load_balancer[0].ingress[0].hostname, "pending")
}

output "public_url" {
  description = "Public URL of the application"
  value       = "http://${try(kubernetes_service.app.status[0].load_balancer[0].ingress[0].hostname, "pending")}:${var.app_port}"
}
