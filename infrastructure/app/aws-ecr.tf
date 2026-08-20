# 手動編集: 可 - API コンテナイメージを置く ECR をここに定義する。

# Go API の Docker image を保存する repository。
# 同じ tag の上書きを防ぐため image_tag_mutability は IMMUTABLE にする。
resource "aws_ecr_repository" "api" {
  #checkov:skip=CKV_AWS_136: KMS key は鍵管理コストと運用設計が必要なため、8週目では AWS 管理の AES256 暗号化に留める。
  name                 = "${local.name_prefix}-${var.ecr_repository_name}"
  image_tag_mutability = "IMMUTABLE"
  force_delete         = true

  # push された image を ECR 側でスキャンする。
  image_scanning_configuration {
    scan_on_push = true
  }

  # AWS 管理の暗号化キーで image を暗号化する。
  encryption_configuration {
    encryption_type = "AES256"
  }

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-${var.ecr_repository_name}"
      Role = "container-registry"
    }
  )
}

# tag が外れた古い image は溜まりやすいため、指定日数後に削除する。
resource "aws_ecr_lifecycle_policy" "api" {
  repository = aws_ecr_repository.api.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Expire untagged images after ${var.ecr_lifecycle_untagged_days} days"
        selection = {
          tagStatus   = "untagged"
          countType   = "sinceImagePushed"
          countUnit   = "days"
          countNumber = var.ecr_lifecycle_untagged_days
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}
