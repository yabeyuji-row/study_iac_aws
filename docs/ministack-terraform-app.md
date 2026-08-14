# MiniStack Terraform App 手順

## 目的

MiniStack では、VPC、subnet、security group、IAM role などの API 作成確認に留める。
実ネットワーク制御、IAM 権限評価、RDS の private subnet 内到達性は AWS 実環境で確認する前提にする。

## Provider override の考え方

MiniStack に向ける場合は、通常の AWS endpoint ではなく MiniStack endpoint を Terraform provider に指定する。
有効な `.tf` として常設すると通常の AWS plan と混ざるため、ローカル検証時だけ override 用ファイルを一時的に置く。

例:

```hcl
provider "aws" {
  region                      = var.aws_region
  skip_credentials_validation = var.skip_credentials_validation
  skip_metadata_api_check     = var.skip_metadata_api_check
  skip_requesting_account_id  = var.skip_requesting_account_id

  endpoints {
    ec2 = var.aws_endpoint_url
    iam = var.aws_endpoint_url
    sts = var.aws_endpoint_url
  }
}
```

`envs/dev.tfvars` では、ローカル確認用に次の値を指定する。

```hcl
skip_credentials_validation = true
skip_metadata_api_check     = true
skip_requesting_account_id  = true
aws_endpoint_url            = "http://localhost:4566"
```

## 確認手順

```sh
make ministack-up
make ministack-health
terraform -chdir=infrastructure/app init -backend=false
terraform -chdir=infrastructure/app validate
terraform -chdir=infrastructure/app plan -input=false -refresh=false -var-file=envs/dev.tfvars
```

MiniStack 側に向いていることは、`aws_endpoint_url` が `http://localhost:4566` になっていることと、MiniStack health check が成功していることで確認する。

## 注意

- override ファイルに実 AWS 認証情報を書かない。
- `terraform apply` はこの週の必須作業にしない。
- AWS apply が必要な場合は、対象 resource と概算コストを先に報告して承認を得る。
