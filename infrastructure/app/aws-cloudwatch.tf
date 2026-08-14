# 手動編集: 可 - ECS タスクのアプリケーションログ出力先をここに定義する。

# Go API の標準出力ログを CloudWatch Logs に集約する log group。
resource "aws_cloudwatch_log_group" "api" {
  #checkov:skip=CKV_AWS_158: KMS key は鍵管理コストと rotation 設計が必要なため、後続の security フェーズで扱う。
  #checkov:skip=CKV_AWS_338: 開発環境のコスト管理を優先し、8週目では短い保持期間を variable で管理する。
  name              = "/ecs/${local.name_prefix}-api"
  retention_in_days = var.log_retention_days

  tags = merge(
    local.common_tags,
    {
      Name = "/ecs/${local.name_prefix}-api"
      Role = "application-logs"
    }
  )
}
