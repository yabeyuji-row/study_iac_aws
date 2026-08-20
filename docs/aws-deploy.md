# AWS デプロイ手順

この手順は dev 環境を AWS に作成し、TODO API を ECS Fargate で動かすためのものです。
AWS リソース作成には課金が発生します。不要になったら cleanup checklist に従って削除してください。

## 前提

- AWS CLI が対象アカウントに認証済みであること。
- Terraform CLI が利用できること。
- Docker が利用できること。
- 初回はローカルから `terraform apply` できる権限が必要です。

確認:

```sh
aws sts get-caller-identity
terraform version
docker version
```

## 1. 実 AWS 用 tfvars を用意する

`infrastructure/app/envs/dev.tfvars` は commit しません。
実 AWS では MiniStack 用 endpoint を外し、provider 検証 skip を false にします。
NAT Gateway を使わない dev 構成では、ECS task を public subnet に置き、public IP と HTTPS egress を有効にします。

```hcl
project_name = "study-aws"
environment  = "dev"
aws_region   = "ap-northeast-1"

skip_credentials_validation = false
skip_metadata_api_check     = false
skip_requesting_account_id  = false
aws_endpoint_url            = null

availability_zone_names = [
  "ap-northeast-1a",
  "ap-northeast-1c",
]

ecs_task_subnet_tier   = "public"
ecs_assign_public_ip   = true
ecs_allow_https_egress = true

github_repository         = "yabeyuji-row@316766040/study_iac_aws@1333772562"
github_default_branch     = "main"
github_deploy_environment = "dev"
```

既存の `dev.tfvars` がある場合は、上の差分だけ反映してください。

## 2. Terraform で AWS リソースを作成する

```sh
terraform -chdir=infrastructure/app init -backend=false
terraform -chdir=infrastructure/app fmt -recursive
terraform -chdir=infrastructure/app validate
terraform -chdir=infrastructure/app plan -var-file=envs/dev.tfvars -out=dev.tfplan
terraform -chdir=infrastructure/app apply dev.tfplan
```

作成後に値を確認します。

```sh
terraform -chdir=infrastructure/app output
```

GitHub Actions から deploy する場合は、以下の repository variables に output を設定します。

```text
AWS_REGION
AWS_PLAN_ROLE_ARN
AWS_DEPLOY_ROLE_ARN
ECR_REPOSITORY_URL
ECS_CLUSTER
ECS_SERVICE
ECS_TASK_DEFINITION_FAMILY
```

`ECS_TASK_DEFINITION_FAMILY` は通常 `study-aws-dev-api` です。

## 3. 最初のイメージを ECR に push する

GitHub Actions の `Deploy` workflow を手動実行します。
初回は `deploy = true` と `migrate = true` を指定すると、イメージ push 後に
database migration を one-off ECS task として実行し、ECS service も更新します。

ローカルから直接 push する場合:

```sh
aws ecr get-login-password --region ap-northeast-1 \
  | docker login --username AWS --password-stdin "$(terraform -chdir=infrastructure/app output -raw ecr_repository_url | cut -d/ -f1)"

image="$(terraform -chdir=infrastructure/app output -raw ecr_repository_url):$(git rev-parse HEAD)"

docker build \
  --build-arg VERSION="$(git rev-parse HEAD)" \
  --build-arg COMMIT="$(git rev-parse HEAD)" \
  --build-arg BUILD_TIME="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -t "$image" .

docker push "$image"
```

## 4. ECS service を更新する

GitHub Actions の `Deploy` workflow を `deploy = true` で実行するのが標準です。
DB schema を更新する場合は `migrate = true` も指定します。

確認:

```sh
alb_dns_name="$(terraform -chdir=infrastructure/app output -raw alb_dns_name)"
curl -fsS "http://${alb_dns_name}/healthz"
curl -fsS "http://${alb_dns_name}/readyz"
```

## 削除

不要になったら、先に [AWS クリーンアップ用チェックリスト](operations/cleanup-checklist.md) を確認してください。
RDS は final snapshot を作る設定なので、snapshot 名の衝突に注意します。
