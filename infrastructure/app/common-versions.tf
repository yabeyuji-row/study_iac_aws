# 手動編集: 可 - Terraform とプロバイダーのバージョン制約を意図して変更する。
# アプリケーションインフラ用の Terraform CLI とプロバイダーのバージョン制約。
terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.7"
    }
  }
}
