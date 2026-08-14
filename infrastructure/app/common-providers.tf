# 手動編集: 可 - プロバイダー設定を意図して変更する。
# AWS 認証情報は標準の AWS SDK 解決順に任せる。
# Terraform ファイルにはアクセスキーやシークレットキーを書かない。
provider "aws" {
  region                      = var.aws_region
  skip_credentials_validation = var.skip_credentials_validation
  skip_metadata_api_check     = var.skip_metadata_api_check
  skip_requesting_account_id  = var.skip_requesting_account_id

  endpoints {
    ec2            = var.aws_endpoint_url
    ecr            = var.aws_endpoint_url
    ecs            = var.aws_endpoint_url
    elbv2          = var.aws_endpoint_url # Elastic Load Balancing v2
    iam            = var.aws_endpoint_url
    logs           = var.aws_endpoint_url
    rds            = var.aws_endpoint_url
    secretsmanager = var.aws_endpoint_url
    sts            = var.aws_endpoint_url
  }
}
