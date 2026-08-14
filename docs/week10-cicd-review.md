# 10週目 CI/CD レビュー

## Workflow 構成

10週目では GitHub Actions workflow を 3 つ追加する。

- `.github/workflows/ci.yml`: Go test、race test、vet、govulncheck、build、Docker build。
- `.github/workflows/terraform.yml`: Terraform fmt、validate、TFLint、Checkov、plan artifact。
- `.github/workflows/deploy.yml`: ECR image push と、手動 gate 付き ECS service update。

## CI

CI は pull request と main branch push で実行する。
Go test、race test、lint、vulnerability check、binary build を先に実行し、成功した場合だけ Docker build を行う。
Docker build が失敗した場合、image push や deploy workflow には進まない。

## Terraform check

Terraform workflow は infrastructure と workflow 定義の変更時に実行する。
`envs/dev.tfvars.example` を `envs/dev.tfvars` にコピーし、secret を含まない local plan 用設定で確認する。
plan は `-refresh=false` と `-no-color` で生成し、`terraform-plan.txt` artifact として保存する。

plan artifact では以下を確認する。

- DB password は `(sensitive value)` と表示される。
- Secrets Manager `secret_string` は `(sensitive value)` と表示される。
- `.env` や real `tfvars` は workflow に渡さない。

## GitHub OIDC

Terraform は GitHub OIDC provider と 2 つの role を定義する。

- `github_plan`: Terraform plan 用の read-only role。
- `github_deploy`: ECR push と ECS deploy 用の role。

trust policy は `aud = sts.amazonaws.com` を必須にし、`sub` を repository と用途に絞る。

- plan role: default branch と pull request。
- deploy role: GitHub environment `dev`。

deploy role の policy は ECR push と ECS deploy に分割する。
ECR repository write は対象 repository ARN に絞る。
ECS service update は対象 service に絞る。
`ecr:GetAuthorizationToken`、`ecs:RegisterTaskDefinition`、`ecs:DescribeTaskDefinition` は AWS IAM の仕様上 `Resource = "*"` が必要なため、該当 statement に理由を残す。
`iam:PassRole` は ECS task execution role と task role に絞り、`iam:PassedToService = ecs-tasks.amazonaws.com` 条件を付ける。

## Deploy gate

`deploy.yml` は main branch または `v*` tag push で image push できる。
ECS service update は自動実行しない。
ECS deploy は `workflow_dispatch` で `deploy = true` を指定し、GitHub environment `dev` の承認 gate を通った場合だけ実行する。

GitHub repository variables:

- `AWS_REGION`
- `AWS_DEPLOY_ROLE_ARN`
- `ECR_REPOSITORY_URL`
- `ECS_CLUSTER`
- `ECS_SERVICE`
- `ECS_TASK_DEFINITION_FAMILY`

## Migration 方針

API 起動時に migration は実行しない。
DB migration は deploy 前の明示手順として扱う。

初期方針:

1. migration SQL を review する。
2. backup/snapshot の取得状態を確認する。
3. `make migrate-up` 相当の one-off 実行を手動 gate で行う。
4. migration 成功後に ECS deploy を実行する。
5. 失敗時は migration down が安全かを確認し、必要なら snapshot restore を選ぶ。

自動 migration は、互換性のある forward-only migration と rollback 手順が安定してから追加する。

## Rollback

ECS deployment circuit breaker を有効にしているため、新 task が healthy にならない場合は ECS が rollback する。
手動 rollback が必要な場合は、直前の task definition revision を指定して `aws ecs update-service` を実行する。

確認項目:

- ALB target group が healthy target を持つ。
- `/readyz` が 200 を返す。
- CloudWatch Logs に起動エラーがない。
- RDS 接続失敗や migration エラーがない。

## 障害演習

### Go test failure

一時的に Go test を失敗させると、`CI / Go checks` が失敗し Docker build job は実行されない。
戻す場合は一時変更を削除し、`go test ./...` を再実行する。

### Docker build failure

一時的に Dockerfile を壊すと、CI の Docker build job が失敗する。
deploy workflow は CI workflow とは独立しているため、branch protection で CI 成功を required check にする。

### OIDC role ARN failure

`AWS_DEPLOY_ROLE_ARN` を誤設定すると、`configure-aws-credentials` step で失敗する。
ECR push と ECS deploy は実行されない。
戻す場合は repository variable を Terraform output `github_deploy_role_arn` に戻す。

### Missing image tag

存在しない image tag を ECS deploy に指定すると、新 task が image pull で失敗する。
ECS deployment circuit breaker と service events で検出し、直前の task definition revision へ戻す。

## 確認コマンド

```sh
go test ./...
terraform -chdir=infrastructure/app fmt -recursive
terraform -chdir=infrastructure/app validate
tflint --chdir=infrastructure/app
docker run --rm -v "$PWD:/repo" bridgecrew/checkov -d /repo/infrastructure/app --framework terraform
actionlint
```


