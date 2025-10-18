# Security Group for application instances
resource "aws_security_group" "instance" {
  name        = "${var.app_name}-sg"
  description = "Security group for ${var.app_name} application"
  vpc_id      = data.aws_vpc.default.id

  tags = merge(
    var.tags,
    {
      Name        = "${var.app_name}-sg"
      Environment = var.environment
      ManagedBy   = "SCIA"
    }
  )
}

# Application port ingress rule
resource "aws_vpc_security_group_ingress_rule" "app_port" {
  security_group_id = aws_security_group.instance.id
  description       = "Application port"

  from_port   = var.application_port
  to_port     = var.application_port
  ip_protocol = "tcp"
  cidr_ipv4   = "0.0.0.0/0"

  tags = {
    Name = "${var.app_name}-app-port"
  }
}

# SSH ingress rule (use SSM Session Manager instead in production)
resource "aws_vpc_security_group_ingress_rule" "ssh" {
  security_group_id = aws_security_group.instance.id
  description       = "SSH access (use SSM Session Manager instead)"

  from_port   = 22
  to_port     = 22
  ip_protocol = "tcp"
  cidr_ipv4   = "0.0.0.0/0"

  tags = {
    Name = "${var.app_name}-ssh"
  }
}

# Allow all outbound traffic
resource "aws_vpc_security_group_egress_rule" "all_outbound" {
  security_group_id = aws_security_group.instance.id
  description       = "Allow all outbound traffic"

  ip_protocol = "-1"
  cidr_ipv4   = "0.0.0.0/0"

  tags = {
    Name = "${var.app_name}-egress"
  }
}
