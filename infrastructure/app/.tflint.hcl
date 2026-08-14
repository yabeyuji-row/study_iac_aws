# 手動編集: 可 - Terraform lint のルールセットを意図して変更する。
plugin "aws" {
  enabled = true
  version = "0.35.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}
