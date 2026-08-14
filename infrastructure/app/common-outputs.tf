# 手動編集: 可 - Terraform 出力をここに定義する。

output "vpc_id" {
  description = "Application VPC ID."
  value       = aws_vpc.main.id
}

output "public_subnet_ids" {
  description = "Public subnet IDs keyed by availability zone."
  value = {
    for az, subnet in aws_subnet.public :
    az => subnet.id
  }
}

output "private_subnet_ids" {
  description = "Private subnet IDs keyed by availability zone."
  value = {
    for az, subnet in aws_subnet.private :
    az => subnet.id
  }
}

output "alb_security_group_id" {
  description = "ALB security group ID."
  value       = aws_security_group.alb.id
}

output "ecs_security_group_id" {
  description = "ECS task security group ID."
  value       = aws_security_group.ecs.id
}

output "rds_security_group_id" {
  description = "RDS security group ID."
  value       = aws_security_group.rds.id
}

output "ecs_task_execution_role_arn" {
  description = "ECS task execution role ARN."
  value       = aws_iam_role.ecs_task_execution.arn
}

output "ecs_task_role_arn" {
  description = "ECS task role ARN."
  value       = aws_iam_role.ecs_task.arn
}

output "ecr_repository_url" {
  description = "ECR repository URL for the API image."
  value       = aws_ecr_repository.api.repository_url
}

output "cloudwatch_log_group_name" {
  description = "CloudWatch Logs log group name for the API."
  value       = aws_cloudwatch_log_group.api.name
}

output "cloudwatch_dashboard_name" {
  description = "CloudWatch dashboard name for operations."
  value       = aws_cloudwatch_dashboard.operations.dashboard_name
}

output "alb_dns_name" {
  description = "Public DNS name of the ALB."
  value       = aws_lb.app.dns_name
}

output "alb_target_group_arn" {
  description = "ALB target group ARN for the API."
  value       = aws_lb_target_group.api.arn
}

output "ecs_cluster_name" {
  description = "ECS cluster name."
  value       = aws_ecs_cluster.api.name
}

output "ecs_service_name" {
  description = "ECS service name."
  value       = aws_ecs_service.api.name
}

output "rds_instance_address" {
  description = "RDS PostgreSQL endpoint address."
  value       = aws_db_instance.app.address
}

output "rds_subnet_group_name" {
  description = "RDS subnet group name."
  value       = aws_db_subnet_group.app.name
}

output "database_secret_arn" {
  description = "Secrets Manager secret ARN for database connection settings."
  value       = aws_secretsmanager_secret.db.arn
}

output "github_plan_role_arn" {
  description = "IAM role ARN for GitHub Actions Terraform plan."
  value       = aws_iam_role.github_plan.arn
}

output "github_deploy_role_arn" {
  description = "IAM role ARN for GitHub Actions image push and ECS deploy."
  value       = aws_iam_role.github_deploy.arn
}
