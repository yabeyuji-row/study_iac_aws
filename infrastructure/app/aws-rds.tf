# 手動編集: 可 - RDS PostgreSQL と DB secret をここに定義する。

resource "random_password" "db" {
  length  = 32
  special = false
}

# RDS を private subnet のみに配置する subnet group。
resource "aws_db_subnet_group" "app" {
  name       = "${local.name_prefix}-db-subnets"
  subnet_ids = values(aws_subnet.private)[*].id

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-db-subnets"
      Role = "database"
    }
  )
}

# PostgreSQL 設定を明示する parameter group。
resource "aws_db_parameter_group" "app" {
  name   = "${local.name_prefix}-postgres"
  family = var.db_parameter_group_family

  parameter {
    name  = "log_min_duration_statement"
    value = "1000"
  }

  parameter {
    name  = "log_connections"
    value = "1"
  }

  parameter {
    name  = "log_disconnections"
    value = "1"
  }

  parameter {
    name  = "rds.force_ssl"
    value = "1"
  }

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-postgres"
      Role = "database"
    }
  )
}

# TODO API の永続データを保存する RDS PostgreSQL instance。
resource "aws_db_instance" "app" {
  #checkov:skip=CKV_AWS_157: 開発環境では RDS Multi-AZ の常時コストを避け、本番設計時に有効化を検討する。
  #checkov:skip=CKV_AWS_293: 開発環境では削除しやすさを優先し、destroy 前の final snapshot 取得で復旧点を残す。
  #checkov:skip=CKV_AWS_353: Performance Insights は追加コストと監視設計が必要なため、後続の DB 運用設計で扱う。
  #checkov:skip=CKV_AWS_118: Enhanced Monitoring は IAM role と追加メトリクス設計が必要なため、後続の DB 運用設計で扱う。
  identifier = "${local.name_prefix}-postgres"

  engine         = "postgres"
  engine_version = var.db_engine_version
  instance_class = var.db_instance_class

  allocated_storage     = var.db_allocated_storage
  max_allocated_storage = var.db_max_allocated_storage
  storage_type          = "gp3"
  storage_encrypted     = true

  db_name  = var.db_name
  username = var.db_username
  password = random_password.db.result
  port     = var.db_port

  db_subnet_group_name   = aws_db_subnet_group.app.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  parameter_group_name   = aws_db_parameter_group.app.name
  publicly_accessible    = false
  multi_az               = false

  backup_retention_period = var.db_backup_retention_days
  backup_window           = "17:00-18:00"
  maintenance_window      = "sun:18:00-sun:19:00"

  enabled_cloudwatch_logs_exports     = ["postgresql", "upgrade"]
  iam_database_authentication_enabled = true

  deletion_protection       = false
  skip_final_snapshot       = false
  final_snapshot_identifier = "${local.name_prefix}-postgres-final"

  auto_minor_version_upgrade = true
  copy_tags_to_snapshot      = true
  apply_immediately          = false

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-postgres"
      Role = "database"
    }
  )
}

# DB 接続情報を保存する Secrets Manager secret。
resource "aws_secretsmanager_secret" "db" {
  #checkov:skip=CKV_AWS_149: KMS CMK は鍵管理コストと rotation 設計が必要なため、security フェーズで扱う。
  #checkov:skip=CKV2_AWS_57: 自動 rotation は Lambda と rotation 手順が必要なため、DB 運用設計後に追加する。
  name                    = "${local.name_prefix}/database"
  description             = "Database connection settings for the TODO API."
  recovery_window_in_days = 7

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-database-secret"
      Role = "database-secret"
    }
  )
}

resource "aws_secretsmanager_secret_version" "db" {
  secret_id = aws_secretsmanager_secret.db.id

  secret_string = jsonencode({
    host     = aws_db_instance.app.address
    port     = var.db_port
    dbname   = var.db_name
    username = var.db_username
    password = random_password.db.result
    sslmode  = var.db_sslmode
  })
}
