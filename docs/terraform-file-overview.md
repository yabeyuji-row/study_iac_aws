# Terraform ファイル概要

このドキュメントは、`infrastructure/` 配下にある Terraform ファイルの役割をまとめる。

このプロジェクトでは Terraform state 用の土台と、アプリケーション用 AWS インフラを別々の configuration として管理する。

```text
infrastructure/
  bootstrap/   # Terraform state 保存先など、最初に作る土台
  app/         # VPC、subnet、ECS、RDS などアプリ本体のインフラ
```

各 Terraform 関連ファイルの冒頭には、手で編集するファイルかどうかを判別するコメントを置く。

```hcl
# 手動編集: 可 - ...
```

`手動編集: 可` は、人間が意図を持って編集するファイルである。

```hcl
# 手動編集: 不可 - ...
```

`手動編集: 不可` は、Terraform などのツールが自動更新するファイルである。通常は手で編集しない。

## `infrastructure/bootstrap/`

`bootstrap` は、アプリケーション用 Terraform が利用する remote state 保存先を作るための configuration である。

state 保存先そのものを作る段階なので、ここでは S3 backend 設定を使わず、local state で実行する想定にする。

app infrastructure で S3 backend を使う場合も、この段階では有効な `.tf` ファイルには書かない。`infrastructure/app/backend.tf.example` に `terraform` ブロックの例として残す。

| ファイル | 編集 | 概要 |
| --- | --- | --- |
| `infrastructure/bootstrap/versions.tf` | 可 | `terraform` ブロックで Terraform CLI の最小バージョンと、利用する provider の取得元・バージョン制約を定義する。 |
| `infrastructure/bootstrap/providers.tf` | 可 | `provider` ブロックで AWS provider の基本設定を定義する。認証情報は AWS SDK の標準解決順に任せ、access key や secret key は書かない。 |
| `infrastructure/bootstrap/variables.tf` | 可 | `variable` ブロックで bootstrap 実行時に外から渡す値を定義する。現在は state bucket 名、AWS region、共通 tags を扱う。 |
| `infrastructure/bootstrap/main.tf` | 可 | `resource` ブロックで Terraform remote state 用 S3 bucket と、versioning、server-side encryption、public access block を定義する。 |
| `infrastructure/bootstrap/outputs.tf` | 可 | `output` ブロックで bootstrap 適用後に確認したい値を出力する。現在は state bucket 名と region を出力する。 |
| `infrastructure/bootstrap/.terraform.lock.hcl` | 不可 | `terraform init` が provider のバージョンと checksum を記録する lock file。通常は手で編集しない。 |
| `infrastructure/bootstrap/tests/state_bucket_name.tftest.hcl` | 可 | `terraform test` 用のテスト。bucket 名が variable どおりになることと、不正な bucket 名が validation で失敗することを確認する。 |

## `infrastructure/app/`

`app` は、アプリケーションを AWS 上で動かすための本体インフラを管理する configuration である。

将来的には次の構成を目指す。

```text
Client -> Application Load Balancer -> ECS Fargate -> Go REST API -> RDS PostgreSQL
```

| ファイル | 編集 | 概要 |
| --- | --- | --- |
| `infrastructure/app/common-versions.tf` | 可 | `terraform` ブロックで app infrastructure に使う Terraform CLI と AWS provider のバージョン制約を定義する。 |
| `infrastructure/app/backend.tf.example` | 可 | S3 backend を有効化するときの `terraform` ブロック例を置く。`.tf.example` は Terraform に読み込まれないため、この段階では backend を有効化しない。 |
| `infrastructure/app/common-providers.tf` | 可 | `provider` ブロックで AWS provider の region、ローカル plan/MiniStack 向けの検証 skip、任意の endpoint URL を variable から設定する。認証情報は Terraform ファイルに書かない。 |
| `infrastructure/app/common-variables.tf` | 可 | `variable` ブロックで環境ごとに変わる入力値を定義する。現在は environment、AWS region、provider 検証 skip、endpoint URL、project 名、VPC CIDR、AZ 数、AZ 名一覧、app port、DB port、ECR repository 名、image tag、ログ保持期間、health check path、ECS desired count、task CPU/memory、共通 tags を扱う。 |
| `infrastructure/app/common-locals.tf` | 可 | `locals` ブロックで入力値から派生する共通値を一元管理する。現在は name prefix、共通 tags、利用 AZ、public/private subnet CIDR、subnet 定義 map を組み立てる。 |
| `infrastructure/app/aws-subnet.tf` | 可 | `resource` ブロックで VPC、subnet、internet gateway、route table、route table association などネットワーク resource を定義する。 |
| `infrastructure/app/aws-security-groups.tf` | 可 | `resource` ブロックで ALB、ECS、RDS の security group と ingress/egress rule を定義する。 |
| `infrastructure/app/aws-iam.tf` | 可 | `resource` ブロックで ECS task execution role と task role を定義する。7週目時点では task role に追加権限を付けない。 |
| `infrastructure/app/aws-ecr.tf` | 可 | `resource` ブロックで API image を保存する ECR repository と lifecycle policy を定義する。 |
| `infrastructure/app/aws-cloudwatch.tf` | 可 | `resource` ブロックで ECS task の標準出力を集約する CloudWatch Logs log group を定義する。 |
| `infrastructure/app/aws-observability.tf` | 可 | `resource` ブロックで CloudWatch Logs metric filter、alarm、dashboard を定義する。 |
| `infrastructure/app/aws-alb.tf` | 可 | `resource` ブロックで public ALB、target group、HTTP listener、health check を定義する。 |
| `infrastructure/app/aws-ecs.tf` | 可 | `resource` ブロックで ECS cluster、Fargate task definition、ECS service を定義する。 |
| `infrastructure/app/aws-rds.tf` | 可 | `resource` ブロックで RDS PostgreSQL、DB subnet group、parameter group、Secrets Manager secret を定義する。 |
| `infrastructure/app/aws-github-oidc.tf` | 可 | `resource` / `data` ブロックで GitHub Actions OIDC provider、CI/CD 用 IAM role と policy を定義する。 |
| `infrastructure/app/common-outputs.tf` | 可 | `output` ブロックで VPC ID、subnet IDs、security group IDs、IAM role ARNs、ECR repository URL、log group 名、ALB DNS 名、ECS cluster/service 名を出力する。 |
| `infrastructure/app/.terraform.lock.hcl` | 不可 | `terraform init` が provider のバージョンと checksum を記録する lock file。通常は手で編集しない。 |
| `infrastructure/app/envs/dev.tfvars.example` | 可 | dev 環境用の入力値サンプル。実値を入れた `dev.tfvars` は commit しない。 |
| `infrastructure/app/envs/dev.tfvars` | commit 不可 | local plan 用の実値ファイル。秘密値を含み得るため commit しない。必要な場合だけローカルで編集する。 |

## ファイル分割の考え方

Terraform は同じディレクトリ内の `.tf` ファイルをまとめて 1 つの configuration として読む。そのため、ファイル名によって Terraform の評価順が変わるわけではない。

このリポジトリでは、学習しやすくするために責務でファイルを分ける。

- `common-versions.tf`: `terraform` ブロックで Terraform と provider のバージョン境界。
- `backend.tf.example`: `terraform` ブロックで S3 backend 設定の例。`.tf.example` なので Terraform は読み込まない。
- `common-providers.tf`: `provider` ブロックで provider の接続先や region。
- `common-variables.tf`: `variable` ブロックで外から渡す入力値。
- `common-locals.tf`: `locals` ブロックで入力値から計算した内部共通値。
- `aws-subnet.tf`: `resource` ブロックで VPC と subnet などネットワーク。
- `aws-security-groups.tf`: `resource` ブロックで通信許可ルール。
- `aws-iam.tf`: `resource` ブロックで AWS 権限。
- `aws-ecr.tf`: `resource` ブロックで container image repository。
- `aws-cloudwatch.tf`: `resource` ブロックで application logs の保存先となる CloudWatch Logs log group。
- `aws-observability.tf`: `resource` ブロックで request/error/latency/readiness metric filter、ALB/ECS/RDS alarm、operations dashboard。
- `aws-alb.tf`: `resource` ブロックで public entrypoint と health check。
- `aws-ecs.tf`: `resource` ブロックで Fargate の実行単位と service。
- `aws-rds.tf`: `resource` ブロックで PostgreSQL database と database secret。
- `aws-github-oidc.tf`: `resource` / `data` ブロックで GitHub Actions OIDC と CI/CD 用 IAM。
- `common-outputs.tf`: `output` ブロックで実行後に確認・参照したい値。

## `common-variables.tf` と `common-locals.tf` の違い

`common-variables.tf` は実行者や環境が変える値を書く場所である。

例:

- `project_name`
- `environment`
- `aws_region`
- `vpc_cidr`
- `az_count`
- `tags`

`common-locals.tf` は、`locals` ブロックで variable から計算できる値をまとめる場所である。

例:

- `name_prefix = "${var.project_name}-${var.environment}"`
- `common_tags = merge(...)`
- `availability_zones = slice(...)`
- `public_subnet_cidrs = [...]`
- `private_subnet_cidrs = [...]`

この分担にすると、resource 側では命名規則や subnet 計算を何度も書かずに済む。

## 実行時の注意

`terraform apply` と `terraform destroy` は AWS 上の resource を作成・削除するため、明示的な承認なしに実行しない。
復旧、drift、cost、cleanup の運用手順は `docs/operations/` にまとめる。

検証だけを行う場合は、まず次のようなコマンドを使う。

```sh
terraform -chdir=infrastructure/app fmt -recursive
terraform -chdir=infrastructure/app init -backend=false
terraform -chdir=infrastructure/app validate
terraform -chdir=infrastructure/app plan -input=false -refresh=false -var-file=envs/dev.tfvars
```
