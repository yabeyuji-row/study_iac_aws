# 手動編集: 可 - Terraform 出力をここに定義する。
# 概要:
# ブートストラップ適用後に参照したい値を出力する。
#
# 詳細:
# ここで出す値は、後続のアプリケーションインフラが
# S3 バックエンドやリモートステート設定を組み立てるときの確認材料になる。
# 秘密情報は出力しない。

# 作成された Terraform ステートバケット名。
output "state_bucket_name" {
  # 出力の用途を Terraform 利用者に説明する。
  description = "S3 bucket name for Terraform remote state."
  # aws_s3_bucket リソースの実際のバケット名を返す。
  value = aws_s3_bucket.terraform_state.bucket
}

# Terraform ステートバケットの AWS リージョン。
output "state_bucket_region" {
  # バックエンド設定時にリージョンを確認できるように説明する。
  description = "AWS region for the Terraform state bucket."
  # プロバイダーに渡したリージョンと同じ値を返す。
  value = var.aws_region
}
