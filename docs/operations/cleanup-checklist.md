# AWS Cleanup Checklist

## 方針

学習用 AWS resource は、必要な検証が終わったら削除する。
`terraform destroy` は明示的な承認なしに実行しない。
削除前に、対象 resource、想定停止時間、概算コスト、戻し方、残る snapshot/backup を確認する。

## Terraform 管理 resource

確認:

```sh
terraform -chdir=infrastructure/app state list
terraform -chdir=infrastructure/app plan -destroy -var-file=envs/dev.tfvars
```

削除実行は承認後のみ:

```sh
terraform -chdir=infrastructure/app destroy -var-file=envs/dev.tfvars
```

## 削除後確認

ALB:

```sh
aws elbv2 describe-load-balancers
aws elbv2 describe-target-groups
```

ECS:

```sh
aws ecs list-clusters
aws ecs list-services --cluster study-aws-dev-cluster
aws ecs list-tasks --cluster study-aws-dev-cluster
```

RDS:

```sh
aws rds describe-db-instances
aws rds describe-db-snapshots
aws rds describe-db-cluster-snapshots
aws rds describe-db-snapshot-attributes --db-snapshot-identifier "$SNAPSHOT_IDENTIFIER"
```

Networking:

```sh
aws ec2 describe-vpcs
aws ec2 describe-subnets
aws ec2 describe-security-groups
aws ec2 describe-nat-gateways
aws ec2 describe-addresses
```

CloudWatch:

```sh
aws logs describe-log-groups
aws cloudwatch describe-alarms
aws cloudwatch list-dashboards
```

Secrets Manager:

```sh
aws secretsmanager list-secrets
```

ECR:

```sh
aws ecr describe-repositories
aws ecr list-images --repository-name todo-api
```

IAM/OIDC:

```sh
aws iam list-open-id-connect-providers
aws iam list-roles --query 'Roles[?contains(RoleName, `study-aws-dev`)].RoleName'
```

## 残課金観点

- RDS final snapshot と manual snapshot。
- retained automated backups。
- CloudWatch Logs log group。
- CloudWatch alarms と dashboards。
- ECR image layers。
- NAT Gateway と Elastic IP。
- Secrets Manager secret の recovery window。
- S3 remote state bucket と ALB access log bucket。

## 机上確認結果

12週目では実 AWS resource の restore、destroy、cleanup は実行していない。
cleanup checklist は、AWS apply 後に残課金 resource を確認するための手順として読む。
