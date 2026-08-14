# 手動編集: 可 - セキュリティグループとルールをここに定義する。

# ALB 用のセキュリティグループ。
# インターネットから HTTP だけを受け、ECS への通信だけを許可する。
resource "aws_security_group" "alb" {
  name        = "${local.name_prefix}-alb-sg"
  description = "Security group for the public ALB."
  vpc_id      = aws_vpc.main.id

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-alb-sg"
      Role = "alb"
    }
  )
}

# インターネットから ALB への HTTP ingress。
resource "aws_vpc_security_group_ingress_rule" "alb_http" {
  #checkov:skip=CKV_AWS_260: public ALB の HTTP 入口として 7週目の要件で意図的に許可する。
  security_group_id = aws_security_group.alb.id
  description       = "Allow HTTP from the internet to the ALB."

  cidr_ipv4   = "0.0.0.0/0"
  ip_protocol = "tcp"
  from_port   = 80
  to_port     = 80
}

# ALB から ECS タスクへの app port egress。
resource "aws_vpc_security_group_egress_rule" "alb_to_ecs" {
  security_group_id            = aws_security_group.alb.id
  referenced_security_group_id = aws_security_group.ecs.id
  description                  = "Allow ALB traffic to ECS tasks."

  ip_protocol = "tcp"
  from_port   = var.app_port
  to_port     = var.app_port
}

# ECS タスク用のセキュリティグループ。
# ALB からの app port だけを受け、RDS への PostgreSQL 通信だけを許可する。
resource "aws_security_group" "ecs" {
  name        = "${local.name_prefix}-ecs-sg"
  description = "Security group for ECS tasks."
  vpc_id      = aws_vpc.main.id

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-ecs-sg"
      Role = "ecs"
    }
  )
}

# ALB から ECS タスクへの app port ingress。
resource "aws_vpc_security_group_ingress_rule" "ecs_from_alb" {
  security_group_id            = aws_security_group.ecs.id
  referenced_security_group_id = aws_security_group.alb.id
  description                  = "Allow application traffic from the ALB."

  ip_protocol = "tcp"
  from_port   = var.app_port
  to_port     = var.app_port
}

# ECS タスクから RDS への PostgreSQL egress。
resource "aws_vpc_security_group_egress_rule" "ecs_to_rds" {
  security_group_id            = aws_security_group.ecs.id
  referenced_security_group_id = aws_security_group.rds.id
  description                  = "Allow PostgreSQL traffic from ECS tasks to RDS."

  ip_protocol = "tcp"
  from_port   = var.db_port
  to_port     = var.db_port
}

# RDS PostgreSQL 用のセキュリティグループ。
# ECS タスクからの PostgreSQL 通信だけを受ける。
resource "aws_security_group" "rds" {
  name        = "${local.name_prefix}-rds-sg"
  description = "Security group for RDS PostgreSQL."
  vpc_id      = aws_vpc.main.id

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-rds-sg"
      Role = "rds"
    }
  )
}

# ECS タスクから RDS への PostgreSQL ingress。
resource "aws_vpc_security_group_ingress_rule" "rds_from_ecs" {
  security_group_id            = aws_security_group.rds.id
  referenced_security_group_id = aws_security_group.ecs.id
  description                  = "Allow PostgreSQL traffic from ECS tasks."

  ip_protocol = "tcp"
  from_port   = var.db_port
  to_port     = var.db_port
}
