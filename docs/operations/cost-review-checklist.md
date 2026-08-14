# AWS Cost Review Checklist

## 方針

AWS resource を作成、変更、削除する前に、課金対象と削除方法を確認する。
Terraform plan、AWS console、Cost Explorer、Billing alerts を組み合わせて見る。
学習環境では、使わない時間帯は削除できる状態を保つ。

## 共通確認

- `terraform plan` で追加、変更、削除される resource を確認した。
- 常時課金、保存容量課金、リクエスト課金、データ転送課金を分けて確認した。
- 削除手順と削除後に残る resource を確認した。
- tag `Project`、`Environment`、`ManagedBy` が付いている。
- AWS Pricing Calculator または各サービス pricing page で概算を確認した。

## ALB

- ALB 本体の稼働時間課金。
- LCU 課金。
- access logs を有効化する場合の S3 storage cost。
- 不要になった target group、listener、security group が残っていない。

確認:

```sh
aws elbv2 describe-load-balancers
aws elbv2 describe-target-groups
```

## ECS Fargate

- task CPU と memory。
- desired count。
- deploy 中の一時的な追加 task 数。
- Container Insights の metrics cost。
- stopped task は課金停止するが、logs は残る。

確認:

```sh
aws ecs describe-services --cluster study-aws-dev-cluster --services study-aws-dev-api-service
aws ecs list-tasks --cluster study-aws-dev-cluster
```

## RDS

- instance class。
- allocated storage と autoscaling 上限。
- backup retention と manual snapshot。
- final snapshot。
- Multi-AZ、Performance Insights、Enhanced Monitoring の有無。
- stopped DB instance の制約と自動再起動に注意する。

確認:

```sh
aws rds describe-db-instances
aws rds describe-db-snapshots
```

## NAT Gateway

- NAT Gateway は時間課金とデータ処理課金が発生する。
- dev では原則作成しない。
- 必要な場合は、VPC endpoint や public IP 方式との比較を記録する。
- 不要化した NAT Gateway、Elastic IP、private route を削除する。

確認:

```sh
aws ec2 describe-nat-gateways
aws ec2 describe-addresses
```

## CloudWatch

- Logs retention days。
- log volume。
- metric filters。
- alarms。
- dashboards。
- Container Insights。

確認:

```sh
aws logs describe-log-groups
aws cloudwatch describe-alarms
aws cloudwatch list-dashboards
```

## Secrets Manager

- secret 数。
- recovery window 中の scheduled deletion secret。
- rotation を有効化する場合の Lambda や実行コスト。

確認:

```sh
aws secretsmanager list-secrets
```

## ECR

- image storage。
- untagged image lifecycle。
- 古い tag の削除方針。
- scan 結果の保存と脆弱性対応。

確認:

```sh
aws ecr describe-repositories
aws ecr list-images --repository-name todo-api
aws ecr describe-lifecycle-policy --repository-name todo-api
```

## レビュー記録

- レビュー日。
- 対象 environment。
- 今月の見込み額。
- 削除予定 resource。
- 例外的に残す resource と理由。
- 次回確認日。
