# EKS Cluster Module

This module deploys an Amazon EKS (Elastic Kubernetes Service) cluster with a managed node group and a containerized application deployment using Kubernetes resources.

## Features

- **Fully Managed EKS Cluster**: Uses terraform-aws-modules/eks for production-ready cluster provisioning
- **Dedicated VPC**: Isolated network with public and private subnets across 2 availability zones
- **Managed Node Group**: Auto-scaling worker nodes with configurable instance types and sizes
- **Security Hardening**:
  - IMDSv2 enforced on all nodes
  - Encrypted EBS volumes (GP3)
  - Private subnet placement for worker nodes
  - Security groups with least-privilege access
  - Pod security contexts (non-root user)
- **Kubernetes Deployment**: Automated application deployment with LoadBalancer service
- **Health Checks**: Liveness and readiness probes for application reliability
- **Monitoring**: CloudWatch Container Insights support and cluster logging
- **Add-ons**: vpc-cni, coredns, kube-proxy, aws-ebs-csi-driver

## Usage

```hcl
module "eks_app" {
  source = "./opentofu/modules/eks-cluster"

  app_name         = "my-web-app"
  region           = "us-east-1"
  container_image  = "nginx:latest"
  application_port = 8080

  # Node group configuration
  node_instance_type = "t3.medium"
  node_min_size      = 1
  node_max_size      = 3
  node_desired_size  = 2
  node_volume_size   = 30

  # Kubernetes deployment configuration
  replicas    = 2
  environment = "production"

  tags = {
    Project = "SCAI"
  }
}
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| aws | ~> 5.0 |
| kubernetes | ~> 2.20 |

## Providers

| Name | Version |
|------|---------|
| aws | ~> 5.0 |
| kubernetes | ~> 2.20 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| app_name | Application name used for resource naming | `string` | n/a | yes |
| region | AWS region for EKS cluster deployment | `string` | n/a | yes |
| container_image | Container image for the application deployment | `string` | n/a | yes |
| cluster_version | Kubernetes version for EKS cluster | `string` | `"1.31"` | no |
| node_instance_type | EC2 instance type for EKS worker nodes | `string` | `"t3.medium"` | no |
| node_min_size | Minimum number of nodes in the node group | `number` | `1` | no |
| node_max_size | Maximum number of nodes in the node group | `number` | `3` | no |
| node_desired_size | Desired number of nodes in the node group | `number` | `2` | no |
| node_volume_size | Root volume size in GB for worker nodes | `number` | `30` | no |
| application_port | Port number for the application container | `number` | `8080` | no |
| replicas | Number of pod replicas for the application | `number` | `2` | no |
| environment | Environment name (production, staging, development) | `string` | `"production"` | no |
| enable_monitoring | Enable CloudWatch Container Insights for the cluster | `bool` | `true` | no |
| tags | Additional tags to apply to all resources | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| cluster_id | EKS cluster ID |
| cluster_name | EKS cluster name |
| cluster_arn | EKS cluster ARN |
| cluster_endpoint | EKS cluster endpoint URL |
| cluster_version | Kubernetes version running on the cluster |
| cluster_certificate_authority_data | Base64 encoded certificate data (sensitive) |
| cluster_security_group_id | Security group ID attached to the cluster |
| node_group_id | EKS managed node group ID |
| node_group_arn | EKS managed node group ARN |
| node_security_group_id | Security group ID for worker nodes |
| vpc_id | VPC ID where the cluster is deployed |
| vpc_cidr_block | CIDR block for the VPC |
| private_subnets | List of private subnet IDs |
| public_subnets | List of public subnet IDs |
| service_url | LoadBalancer hostname for the application |
| service_name | Kubernetes service name |
| deployment_name | Kubernetes deployment name |
| kubeconfig_command | Command to configure kubectl |
| kubectl_get_pods_command | Command to list pods |

## Post-Deployment

After deployment, configure kubectl to access your cluster:

```bash
# Get the kubeconfig command from Terraform output
terraform output kubeconfig_command

# Run the command (example)
aws eks update-kubeconfig --region us-east-1 --name my-web-app-eks

# Verify cluster access
kubectl get nodes
kubectl get pods -l app=my-web-app
kubectl get svc
```

## Security Considerations

- **Network Isolation**: Worker nodes are deployed in private subnets with NAT gateway for outbound access
- **IMDSv2**: Instance metadata service v2 is enforced on all nodes to prevent SSRF attacks
- **Encryption**: All EBS volumes are encrypted at rest
- **Pod Security**: Containers run as non-root user (UID 1000)
- **Health Probes**: Liveness and readiness probes ensure only healthy pods receive traffic
- **Logging**: Cluster control plane logs are sent to CloudWatch for audit and diagnostics

## Cost Optimization

- Single NAT Gateway: Uses one NAT gateway to reduce costs (not recommended for production HA)
- On-Demand instances: Uses on-demand capacity (consider Spot for cost savings)
- GP3 volumes: Uses GP3 for better price/performance than GP2

## Troubleshooting

**Pods not starting:**
```bash
kubectl describe pod <pod-name>
kubectl logs <pod-name>
```

**LoadBalancer not provisioning:**
```bash
kubectl describe svc <service-name>
# Check subnet tags for kubernetes.io/role/elb
```

**Node group scaling issues:**
```bash
aws eks describe-nodegroup --cluster-name <cluster-name> --nodegroup-name <nodegroup-name>
```

## References

- [EKS Best Practices](https://aws.github.io/aws-eks-best-practices/)
- [terraform-aws-modules/eks](https://registry.terraform.io/modules/terraform-aws-modules/eks/aws/latest)
- [terraform-aws-modules/vpc](https://registry.terraform.io/modules/terraform-aws-modules/vpc/aws/latest)
- [Kubernetes Best Practices](https://kubernetes.io/docs/concepts/configuration/overview/)
