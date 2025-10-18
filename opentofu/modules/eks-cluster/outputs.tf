# Output values from EKS cluster deployment

# Cluster information
output "cluster_id" {
  description = "EKS cluster ID"
  value       = module.eks.cluster_id
}

output "cluster_name" {
  description = "EKS cluster name"
  value       = module.eks.cluster_name
}

output "cluster_arn" {
  description = "EKS cluster ARN"
  value       = module.eks.cluster_arn
}

output "cluster_endpoint" {
  description = "EKS cluster endpoint URL"
  value       = module.eks.cluster_endpoint
}

output "cluster_version" {
  description = "Kubernetes version running on the cluster"
  value       = module.eks.cluster_version
}

output "cluster_certificate_authority_data" {
  description = "Base64 encoded certificate data for cluster authentication"
  value       = module.eks.cluster_certificate_authority_data
  sensitive   = true
}

output "cluster_security_group_id" {
  description = "Security group ID attached to the EKS cluster"
  value       = module.eks.cluster_security_group_id
}

# Node group information
output "node_group_id" {
  description = "EKS managed node group ID"
  value       = module.eks.eks_managed_node_groups["default"].node_group_id
}

output "node_group_arn" {
  description = "EKS managed node group ARN"
  value       = module.eks.eks_managed_node_groups["default"].node_group_arn
}

output "node_security_group_id" {
  description = "Security group ID for worker nodes"
  value       = module.eks.node_security_group_id
}

# VPC information
output "vpc_id" {
  description = "VPC ID where the cluster is deployed"
  value       = module.vpc.vpc_id
}

output "vpc_cidr_block" {
  description = "CIDR block for the VPC"
  value       = module.vpc.vpc_cidr_block
}

output "private_subnets" {
  description = "List of private subnet IDs"
  value       = module.vpc.private_subnets
}

output "public_subnets" {
  description = "List of public subnet IDs"
  value       = module.vpc.public_subnets
}

# Application service information
output "service_url" {
  description = "LoadBalancer hostname for the application service"
  value       = try(kubernetes_service.app.status[0].load_balancer[0].ingress[0].hostname, "pending")
}

output "service_name" {
  description = "Kubernetes service name"
  value       = kubernetes_service.app.metadata[0].name
}

output "deployment_name" {
  description = "Kubernetes deployment name"
  value       = kubernetes_deployment.app.metadata[0].name
}

# Configuration commands
output "kubeconfig_command" {
  description = "Command to configure kubectl to connect to the cluster"
  value       = "aws eks update-kubeconfig --region ${var.region} --name ${module.eks.cluster_name}"
}

output "kubectl_get_pods_command" {
  description = "Command to list pods in the deployment"
  value       = "kubectl get pods -l app=${var.app_name}"
}
