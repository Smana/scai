# Auto Scaling Group with single instance for auto-recovery
module "autoscaling" {
  source  = "terraform-aws-modules/autoscaling/aws"
  version = "~> 9.0"

  name = "${var.app_name}-asg"

  # Single instance configuration with auto-recovery
  min_size         = 1
  max_size         = 1
  desired_capacity = 1

  # Health check configuration
  health_check_type         = "EC2"
  health_check_grace_period = var.health_check_grace_period
  wait_for_capacity_timeout = "10m"

  # VPC and networking
  vpc_zone_identifier = data.aws_subnets.default.ids

  # Launch template configuration
  image_id                 = data.aws_ami.amazon_linux_2023.id
  instance_type            = var.instance_type
  iam_instance_profile_arn = aws_iam_instance_profile.instance.arn

  # Security groups
  security_groups = [aws_security_group.instance.id]

  # Root volume configuration
  block_device_mappings = [
    {
      device_name = "/dev/xvda"
      ebs = {
        volume_size           = var.volume_size
        volume_type           = "gp3"
        iops                  = 3000
        throughput            = 125
        delete_on_termination = true
        encrypted             = true
      }
    }
  ]

  # User data script (base64 encoded)
  user_data = base64encode(var.user_data)

  # Enable detailed CloudWatch monitoring
  enable_monitoring = var.enable_monitoring

  # Instance metadata options (IMDSv2)
  metadata_options = {
    http_endpoint               = "enabled"
    http_tokens                 = "required" # Require IMDSv2
    http_put_response_hop_limit = 1
    instance_metadata_tags      = "enabled"
  }

  # Tagging
  tags = merge(
    var.tags,
    {
      Name        = var.app_name
      Environment = var.environment
      ManagedBy   = "SCIA"
    }
  )

  # Instance refresh on launch template changes
  instance_refresh = {
    strategy = "Rolling"
    preferences = {
      min_healthy_percentage = 0 # Allow complete replacement for single instance
    }
  }
}
