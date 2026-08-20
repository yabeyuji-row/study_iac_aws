# 手動編集: 可 - ECS cluster、task definition、service をここに定義する。

# Go API を実行する ECS cluster。
resource "aws_ecs_cluster" "api" {
  name = "${local.name_prefix}-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-cluster"
      Role = "api-runtime"
    }
  )
}

# Fargate で起動する Go API コンテナの実行定義。
resource "aws_ecs_task_definition" "api" {
  family = "${local.name_prefix}-api"
  # この task definition をどの ECS 起動方式に対応させるかを指定する。
  # 選択できる値は EC2、FARGATE、EXTERNAL、MANAGED_INSTANCES。
  # この構成ではサーバー管理を AWS に任せるため FARGATE を使う。
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = var.ecs_task_cpu
  memory                   = var.ecs_task_memory
  execution_role_arn       = aws_iam_role.ecs_task_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([
    {
      name      = "api"
      image     = "${aws_ecr_repository.api.repository_url}:${var.image_tag}"
      essential = true

      secrets = [
        {
          name      = "DATABASE_SECRET_JSON"
          valueFrom = aws_secretsmanager_secret.db.arn
        }
      ]

      portMappings = [
        {
          containerPort = var.app_port
          hostPort      = var.app_port
          protocol      = "tcp"
        }
      ]

      # distroless image には curl/wget がないため、API binary の healthcheck mode を使う。
      healthCheck = {
        command = [
          "CMD",
          "/api",
          "-healthcheck-url",
          "http://127.0.0.1:${var.app_port}${var.container_health_check_path}",
        ]
        interval    = 30
        timeout     = 5
        retries     = 3
        startPeriod = 10
      }

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.api.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = "api"
        }
      }
    }
  ])

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-api-task"
      Role = "api-task-definition"
    }
  )
}

# ALB target group に登録される Fargate service。
# task は通常 private subnet に置く。dev では NAT Gateway を避けるため public subnet
# と public IP を選べるようにしている。
resource "aws_ecs_service" "api" {
  name                               = "${local.name_prefix}-api-service"
  cluster                            = aws_ecs_cluster.api.id
  task_definition                    = aws_ecs_task_definition.api.arn # 利用する ECS タスク定義
  desired_count                      = var.ecs_desired_count
  launch_type                        = "FARGATE" # service の task をどの起動方式で実行するかを指定する。 EC2、FARGATE、EXTERNAL。
  deployment_minimum_healthy_percent = var.ecs_deployment_minimum_healthy_percent
  deployment_maximum_percent         = var.ecs_deployment_maximum_percent
  health_check_grace_period_seconds  = 60

  deployment_circuit_breaker { # ECS デプロイ失敗時の制御
    enable   = true
    rollback = true
  }

  network_configuration {
    subnets = var.ecs_task_subnet_tier == "public" ? (
      values(aws_subnet.public)[*].id
      ) : (
      values(aws_subnet.private)[*].id
    )
    security_groups  = [aws_security_group.ecs.id]
    assign_public_ip = var.ecs_assign_public_ip
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.api.arn
    container_name   = "api"
    container_port   = var.app_port
  }

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-api-service"
      Role = "api-service"
    }
  )

  lifecycle {
    ignore_changes = [
      task_definition,
    ]
  }

  depends_on = [aws_lb_listener.http] # 明示的な依存関係
}
