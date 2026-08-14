# 8週目 ECS/ALB 障害演習

## 目的

ECS Fargate、ALB target group、CloudWatch Logs の設定を一時的に崩し、
Terraform plan や validation でどのように検出できるかを確認する。

## 8-3-1 container port 誤設定

一時的に ECS task definition の `portMappings.containerPort` と `hostPort` を
`var.app_port + 1` に変更した。

plan では `aws_ecs_task_definition.api` の replacement と
`aws_ecs_service.api` の task definition 更新として確認できた。

影響:

- ALB/ECS service は `var.app_port` の container port を参照する。
- task definition 側の port mapping がずれると、target group 登録や deployment が失敗する可能性がある。
- ALB health check は期待する port の task に到達できず、healthy target が増えない。

## 8-3-2 desired count 0

一時的に `envs/dev.tfvars` の `ecs_desired_count` を `0` に変更した。

plan は生成されず、variable validation で停止した。

```text
ecs_desired_count must be between 1 and 10.
```

影響:

- desired count が 0 になると ECS service は task を起動しない。
- ALB target group に healthy target がなくなり、API は利用できない。
- 現在の Terraform validation は dev 環境でこの誤設定を入力段階で防ぐ。

## 8-3-3 CloudWatch Logs 設定削除

一時的に ECS container definition から `logConfiguration` を削除した。

plan では `aws_ecs_task_definition.api` の replacement と
`aws_ecs_service.api` の task definition 更新として確認できた。

影響:

- アプリケーションログが CloudWatch Logs log group に出なくなる。
- 起動失敗、readiness 失敗、panic、リクエストエラーの調査材料が減る。
- ECS task の状態だけではアプリ内部の失敗理由を追いにくくなる。

## 正常状態への復旧確認

一時変更をすべて戻した後、次のコマンドで正常状態を確認する。

```sh
make test
make docker-build
terraform -chdir=infrastructure/app fmt -recursive
terraform -chdir=infrastructure/app validate
terraform -chdir=infrastructure/app plan -input=false -refresh=false -var-file=envs/dev.tfvars
```
