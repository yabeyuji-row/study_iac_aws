# 手動編集: 可 - ブートストラップリソースをここに定義する。
# 概要:
# Terraform リモートステート用の S3 バケットと、その安全設定を定義する。
#
# 詳細:
# ブートストラップはアプリケーション本体のインフラとは別のライフサイクルで管理する。
# このファイルではステート保存先そのものを作るため、バックエンド設定はまだ使わない。
# terraform apply すると AWS 上に S3 バケットが作成されるため、実行前に明示承認とコスト確認が必要。

# Terraform ステートファイルを保存する S3 バケット本体。
resource "aws_s3_bucket" "terraform_state" {
  # S3 バケット名はグローバルに一意である必要があるため、環境ごとに外から指定する。
  bucket = var.state_bucket_name

  # 所有者、環境、用途などの管理情報を付ける。
  tags = var.tags
}

# ステートの過去版を保持するため、バケットのバージョニングを有効化する。
resource "aws_s3_bucket_versioning" "terraform_state" {
  # 上で作成したバケットに対してバージョニング設定を紐付ける。
  bucket = aws_s3_bucket.terraform_state.id

  versioning_configuration {
    # 誤更新や削除から復旧できるように、ステートオブジェクトのバージョンを残す。
    status = "Enabled"
  }
}

# ステートファイルを S3 側で暗号化して保存する。
resource "aws_s3_bucket_server_side_encryption_configuration" "terraform_state" {
  # 上で作成したバケットに対して暗号化設定を紐付ける。
  bucket = aws_s3_bucket.terraform_state.id

  rule {
    apply_server_side_encryption_by_default {
      # ブートストラップでは追加の KMS キー管理を避け、S3 管理の暗号化を使う。
      sse_algorithm = "AES256"
    }
  }
}

# ステートバケットが誤って公開されないよう、S3 パブリックアクセスブロックを有効化する。
resource "aws_s3_bucket_public_access_block" "terraform_state" {
  # 上で作成したバケットに対して公開ブロック設定を紐付ける。
  bucket = aws_s3_bucket.terraform_state.id

  # パブリック ACL の新規設定を拒否する。
  block_public_acls = true
  # パブリックバケットポリシーの新規設定を拒否する。
  block_public_policy = true
  # 既存または外部由来のパブリック ACL を評価上無視する。
  ignore_public_acls = true
  # パブリックポリシーによる公開アクセスを制限する。
  restrict_public_buckets = true
}
