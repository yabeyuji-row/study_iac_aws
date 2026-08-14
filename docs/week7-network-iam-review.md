# 7週目 ネットワークと IAM レビュー

## 実装した境界

- ALB security group は internet から HTTP/80 だけを受ける。
- Public subnet は Internet Gateway への route を持つが、subnet 側の自動 public IP 割り当ては無効にする。
- ALB から ECS security group へは `app_port` だけを許可する。
- ECS security group は ALB security group からの `app_port` だけを受ける。
- ECS から RDS security group へは `db_port` だけを許可する。
- RDS security group は ECS security group からの `db_port` だけを受ける。
- Private subnet には NAT Gateway と Internet Gateway への default route を置かない。

## IAM role の責務

- `ecs_task_execution` role は ECS agent 用で、ECR pull と CloudWatch Logs 書き込みのために使う。
- `ecs_task` role はアプリケーションコンテナ用で、7週目時点では追加権限を持たない。
- Secrets Manager などへの読み取り権限は、後続週で対象 secret ARN に絞って追加する。

## Plan で確認すること

- `aws_security_group.alb`, `aws_security_group.ecs`, `aws_security_group.rds` が作成される。
- ALB ingress は `0.0.0.0/0` の TCP/80 のみである。
- ECS ingress は ALB security group 参照の `app_port` のみである。
- RDS ingress は ECS security group 参照の `db_port` のみである。
- `aws_iam_role.ecs_task_execution` と `aws_iam_role.ecs_task` が別 role として作成される。
- task role に inline policy や広い managed policy が付いていない。

## 有料 resource の確認

7週目の plan には、常時課金される NAT Gateway、ALB、ECS service、RDS instance は含めない。

この週で Terraform に含める主な resource は VPC、subnet、Internet Gateway、route table、security group、IAM role である。
これらは設定自体では主に課金対象ではないが、AWS apply は実リソース作成を伴うため、実行前に必ず対象 resource と概算コストを確認する。

## Checkov の意図的な skip

- `CKV_AWS_260`: public ALB の HTTP/80 ingress は 7週目の要件として意図的に許可する。HTTPS 化は後続フェーズで扱う。
- `CKV2_AWS_5`: ALB、ECS service、RDS instance は後続週で追加するため、7週目時点の security group は未アタッチになる。
- `CKV2_AWS_11`: VPC Flow Logs は observability フェーズで扱うため、7週目では追加しない。

## 障害演習の検出可否

### ALB HTTP ingress を削除した場合

外部から ALB への入口がなくなる。
`terraform plan` では `aws_vpc_security_group_ingress_rule.alb_http` の削除として確認できる。
TFLint や Checkov は「必要な ingress が消えた」こと自体は業務要件としては判断しにくいため、人手レビューで確認する。

### ECS ingress を削除した場合

ALB から ECS タスクへ通信できなくなる。
`terraform plan` では `aws_vpc_security_group_ingress_rule.ecs_from_alb` の削除として確認できる。
ALB health check やアプリ通信が失敗する想定で、plan と人手レビューで確認する。

### RDS ingress を `0.0.0.0/0` に広げた場合

RDS が広い送信元から PostgreSQL port を受ける危険な構成になる。
Checkov は広い ingress を検出できる可能性がある。
検出できない場合でも、RDS ingress は ECS security group 参照だけかを人手レビューで確認する。

### IAM policy に `Action = "*"` を入れた場合

過剰権限の IAM policy になる。
Checkov は wildcard action を検出できる可能性がある。
7週目の正常構成では task role に追加 policy を付けず、execution role は AWS 管理ポリシーだけに留める。

## 正常系コマンド

```sh
terraform -chdir=infrastructure/app fmt -recursive
terraform -chdir=infrastructure/app init -backend=false
terraform -chdir=infrastructure/app validate
terraform -chdir=infrastructure/app plan -input=false -refresh=false -var-file=envs/dev.tfvars
tflint --chdir=infrastructure/app --init
tflint --chdir=infrastructure/app
checkov -d infrastructure/app --framework terraform
```

## 実行結果

- `terraform -chdir=infrastructure/app fmt -recursive`: 成功。
- `terraform -chdir=infrastructure/app init -backend=false`: 成功。
- `terraform -chdir=infrastructure/app validate`: 成功。
- `terraform -chdir=infrastructure/app plan -input=false -refresh=false -var-file=envs/dev.tfvars`: 成功。22 resource to add、0 change、0 destroy。
- `tflint --chdir=infrastructure/app --init`: 成功。
- `tflint --chdir=infrastructure/app`: 成功。
- `checkov -d infrastructure/app --framework terraform`: 成功。Passed 67、Failed 0、Skipped 5。
- MiniStack health check: 成功。`ec2`、`iam`、`sts` などが available。

Plan に NAT Gateway、ALB、ECS service、RDS instance は含まれていない。
7週目の範囲では、VPC、subnet、Internet Gateway、route table、security group、IAM role の作成予定だけを確認した。
