# 手動編集: 可 - 入力変数と検証を意図して変更する。
# 概要:
# ブートストラップ Terraform に外部から渡す入力値を定義する。
#
# 詳細:
# バケット名やリージョンなど、環境や AWS アカウントによって変わる値を
# ハードコードせず変数として受け取る。秘密情報はここに書かない。

# Terraform ステートを保存する S3 バケット名。
variable "state_bucket_name" {
  # 利用者に、この値がリモートステート用バケット名であることを説明する。
  description = "S3 bucket name for Terraform remote state."
  # S3 バケット名は文字列として指定する。
  type = string

  validation {
    # S3 バケット名の AWS 制約に合わせて検証する。
    # - 3文字以上63文字以下であること。
    # - 小文字英数字、ハイフン、ドットだけを使うこと。
    # - 先頭と末尾、ドット区切りの各要素の先頭と末尾は英数字であること。
    # - 連続したドットを含まないこと。
    # - IP アドレスのような形式ではないこと。
    condition = (
      length(var.state_bucket_name) >= 3 &&
      length(var.state_bucket_name) <= 63 &&
      can(regex("^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)*$", var.state_bucket_name)) &&
      !can(regex("\\.\\.", var.state_bucket_name)) &&
      !can(regex("^[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+$", var.state_bucket_name))
    )
    error_message = "state_bucket_name must be a valid S3 bucket name: 3-63 characters, lowercase letters, numbers, dots, and hyphens only; start and end with a letter or number; no consecutive dots; not an IP address."
  }
}

# ブートストラップリソースを作成する AWS リージョン。
variable "aws_region" {
  # プロバイダー設定で利用するリージョンであることを説明する。
  description = "AWS region for bootstrap resources."
  # AWS リージョン名は文字列として指定する。
  type = string
  # この学習プロジェクトの既定リージョンは東京リージョンにする。
  default = "ap-northeast-1"
}

# ブートストラップリソースに付与する共通タグ。
variable "tags" {
  # コスト管理や所有者識別に使うタグであることを説明する。
  description = "Tags applied to bootstrap resources."
  # 任意個数のキーと値のタグを受け取る。
  type = map(string)
  # タグ指定なしでも検証できるように空の map を既定値にする。
  default = {}
}
