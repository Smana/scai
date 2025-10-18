# IAM Role for EC2 instances with SSM access
resource "aws_iam_role" "instance" {
  name_prefix = "${var.app_name}-ec2-role-"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })

  tags = merge(
    var.tags,
    {
      Name        = "${var.app_name}-ec2-role"
      Environment = var.environment
      ManagedBy   = "SCIA"
    }
  )
}

# Attach AWS Systems Manager policy for instance management
resource "aws_iam_role_policy_attachment" "ssm_managed_instance" {
  role       = aws_iam_role.instance.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

# Attach CloudWatch Agent policy for metrics and logs
resource "aws_iam_role_policy_attachment" "cloudwatch_agent" {
  role       = aws_iam_role.instance.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchAgentServerPolicy"
}

# Instance profile for EC2 instances
resource "aws_iam_instance_profile" "instance" {
  name_prefix = "${var.app_name}-profile-"
  role        = aws_iam_role.instance.name

  tags = merge(
    var.tags,
    {
      Name        = "${var.app_name}-instance-profile"
      Environment = var.environment
      ManagedBy   = "SCIA"
    }
  )
}
