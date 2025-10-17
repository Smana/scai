terraform {
  required_version = ">= 1.6"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# Data sources
data "aws_ami" "amazon_linux_2023" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-*-x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

data "aws_availability_zones" "available" {
  state = "available"
}

# VPC Module - Simple single-AZ setup for cost efficiency
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "${var.app_name}-vpc"
  cidr = "10.0.0.0/16"

  azs            = [data.aws_availability_zones.available.names[0]]
  public_subnets = ["10.0.1.0/24"]

  enable_nat_gateway = false
  enable_vpn_gateway = false
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "${var.app_name}-vpc"
    Application = var.app_name
    ManagedBy   = "scia"
  }
}

# Security Group for application
module "security_group" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "~> 5.0"

  name        = "${var.app_name}-sg"
  description = "Security group for ${var.app_name} application"
  vpc_id      = module.vpc.vpc_id

  # Ingress rules
  ingress_with_cidr_blocks = [
    {
      from_port   = 22
      to_port     = 22
      protocol    = "tcp"
      description = "SSH access"
      cidr_blocks = "0.0.0.0/0"
    },
    {
      from_port   = var.app_port
      to_port     = var.app_port
      protocol    = "tcp"
      description = "Application port"
      cidr_blocks = "0.0.0.0/0"
    }
  ]

  # Egress rule
  egress_with_cidr_blocks = [
    {
      from_port   = 0
      to_port     = 0
      protocol    = "-1"
      description = "Allow all outbound"
      cidr_blocks = "0.0.0.0/0"
    }
  ]

  tags = {
    Name        = "${var.app_name}-sg"
    Application = var.app_name
    ManagedBy   = "scia"
  }
}

# User data script for bootstrapping
locals {
  user_data = <<-EOT
    #!/bin/bash
    set -e

    # Logging
    exec > >(tee /var/log/user-data.log)
    exec 2>&1

    echo "=== SCIA Deployment Bootstrap Started ==="
    date

    # Update system
    dnf update -y

    # Install git
    dnf install -y git

    # Install language-specific dependencies
    %{if var.language == "python"}
    echo "Installing Python ${var.python_version}..."
    dnf install -y python${var.python_version} python${var.python_version}-pip
    %{endif}

    %{if var.language == "javascript" || var.language == "typescript"}
    echo "Installing Node.js ${var.nodejs_version}..."
    dnf install -y nodejs${var.nodejs_version}
    %{endif}

    %{if var.language == "go"}
    echo "Installing Go ${var.go_version}..."
    wget https://go.dev/dl/go${var.go_version}.linux-amd64.tar.gz
    tar -C /usr/local -xzf go${var.go_version}.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    %{endif}

    # Create application directory
    mkdir -p /opt/app
    cd /opt/app

    # Clone repository
    echo "Cloning repository: ${var.repo_url}"
    git clone ${var.repo_url} .

    %{if var.app_subdir != ""}
    cd ${var.app_subdir}
    %{endif}

    # Install dependencies
    %{if var.language == "python"}
    if [ -f requirements.txt ]; then
      echo "Installing Python dependencies..."
      python${var.python_version} -m pip install -r requirements.txt
    fi
    %{endif}

    %{if var.language == "javascript" || var.language == "typescript"}
    if [ -f package.json ]; then
      echo "Installing Node.js dependencies..."
      npm install --production
    fi
    %{endif}

    %{if var.language == "go"}
    if [ -f go.mod ]; then
      echo "Installing Go dependencies..."
      go mod download
    fi
    %{endif}

    # Replace localhost with 0.0.0.0 in common files
    echo "Replacing localhost with 0.0.0.0..."
    find . -type f -name "*.py" -o -name "*.js" -o -name "*.go" | xargs sed -i 's/localhost/0.0.0.0/g' || true
    find . -type f -name "*.py" -o -name "*.js" -o -name "*.go" | xargs sed -i 's/127\.0\.0\.1/0.0.0.0/g' || true

    # Set environment variables
    %{for key, value in var.env_vars}
    export ${key}="${value}"
    echo 'export ${key}="${value}"' >> /etc/profile
    %{endfor}

    # Create systemd service
    cat > /etc/systemd/system/${var.app_name}.service <<-EOF
[Unit]
Description=${var.app_name} Application
After=network.target

[Service]
Type=simple
User=ec2-user
WorkingDirectory=/opt/app%{if var.app_subdir != ""}/${var.app_subdir}%{endif}
ExecStart=${var.start_command}
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

    # Enable and start service
    systemctl daemon-reload
    systemctl enable ${var.app_name}.service
    systemctl start ${var.app_name}.service

    echo "=== SCIA Deployment Bootstrap Completed ==="
    date
  EOT
}

# Auto Scaling Group Module - Single instance with auto-recovery
module "autoscaling" {
  source  = "terraform-aws-modules/autoscaling/aws"
  version = "~> 7.0"

  name = "${var.app_name}-asg"

  # Single instance configuration
  min_size         = 1
  max_size         = 1
  desired_capacity = 1

  # Health check configuration
  health_check_type         = "EC2"
  health_check_grace_period = 300
  wait_for_capacity_timeout = "5m"

  # VPC configuration
  vpc_zone_identifier = module.vpc.public_subnets

  # Launch template configuration
  launch_template_name        = "${var.app_name}-lt"
  launch_template_description = "Launch template for ${var.app_name}"
  update_default_version      = true

  image_id          = data.aws_ami.amazon_linux_2023.id
  instance_type     = var.instance_type
  user_data         = base64encode(local.user_data)
  enable_monitoring = true

  # Network configuration
  network_interfaces = [
    {
      delete_on_termination = true
      description           = "Primary network interface"
      device_index          = 0
      security_groups       = [module.security_group.security_group_id]
      associate_public_ip_address = true
    }
  ]

  # IAM instance profile (if needed)
  create_iam_instance_profile = var.create_iam_role
  iam_role_name               = "${var.app_name}-role"
  iam_role_description        = "IAM role for ${var.app_name}"
  iam_role_policies = {
    AmazonSSMManagedInstanceCore = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
  }

  # Tags
  tags = {
    Name        = "${var.app_name}-asg"
    Application = var.app_name
    ManagedBy   = "scia"
  }
}

# Get the instance IP (for output)
data "aws_instances" "app" {
  depends_on = [module.autoscaling]

  filter {
    name   = "tag:Name"
    values = ["${var.app_name}-asg"]
  }

  filter {
    name   = "instance-state-name"
    values = ["running"]
  }
}

# Outputs
output "public_ip" {
  description = "Public IP address of the instance"
  value       = length(data.aws_instances.app.public_ips) > 0 ? data.aws_instances.app.public_ips[0] : "pending"
}

output "public_url" {
  description = "Public URL of the application"
  value       = length(data.aws_instances.app.public_ips) > 0 ? "http://${data.aws_instances.app.public_ips[0]}:${var.app_port}" : "pending"
}

output "instance_id" {
  description = "Instance ID"
  value       = length(data.aws_instances.app.ids) > 0 ? data.aws_instances.app.ids[0] : "pending"
}

output "security_group_id" {
  description = "Security group ID"
  value       = module.security_group.security_group_id
}

output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}
