output "autoscaling_group_id" {
  description = "The Auto Scaling Group ID"
  value       = module.autoscaling.autoscaling_group_id
}

output "autoscaling_group_name" {
  description = "The Auto Scaling Group name"
  value       = module.autoscaling.autoscaling_group_name
}

output "autoscaling_group_arn" {
  description = "The ARN for this Auto Scaling Group"
  value       = module.autoscaling.autoscaling_group_arn
}

output "launch_template_id" {
  description = "The ID of the launch template"
  value       = module.autoscaling.launch_template_id
}

output "launch_template_latest_version" {
  description = "The latest version of the launch template"
  value       = module.autoscaling.launch_template_latest_version
}

output "security_group_id" {
  description = "The ID of the security group"
  value       = aws_security_group.instance.id
}

output "security_group_name" {
  description = "The name of the security group"
  value       = aws_security_group.instance.name
}

output "iam_role_arn" {
  description = "ARN of the IAM role"
  value       = aws_iam_role.instance.arn
}

output "iam_role_name" {
  description = "Name of the IAM role"
  value       = aws_iam_role.instance.name
}

output "iam_instance_profile_arn" {
  description = "ARN of the IAM instance profile"
  value       = aws_iam_instance_profile.instance.arn
}

output "ami_id" {
  description = "ID of the AMI used for instances"
  value       = data.aws_ami.amazon_linux_2023.id
}

output "application_port" {
  description = "Port number where the application is listening"
  value       = var.application_port
}
