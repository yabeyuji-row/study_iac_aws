# Rollback Procedures

## 方針

rollback は application、infrastructure、database migration を分けて判断する。
DB を巻き戻す操作はデータ損失の可能性があるため、最後の手段として扱う。

## Application Rollback

主な対象:

- Go API の不具合。
- Docker image の起動失敗。
- task definition の環境変数、secret、port、health check の不整合。

手順:

1. 現在と直前の task definition revision、image tag、commit SHA を確認する。
2. migration を伴う deploy か確認する。
3. ECS deployment circuit breaker が自動 rollback している場合は service stable を待つ。
4. 手動 rollback では直前 revision を指定する。

```sh
aws ecs describe-services \
  --cluster study-aws-dev-cluster \
  --services study-aws-dev-api-service \
  --query 'services[0].deployments[].{Status:status,TaskDefinition:taskDefinition,RolloutState:rolloutState,Desired:desiredCount,Running:runningCount}'

aws ecs update-service \
  --cluster study-aws-dev-cluster \
  --service study-aws-dev-api-service \
  --task-definition study-aws-dev-api:<previous-revision>

aws ecs wait services-stable \
  --cluster study-aws-dev-cluster \
  --services study-aws-dev-api-service
```

確認:

- ALB target が healthy。
- `/readyz` が 200。
- 代表 TODO CRUD が成功。
- CloudWatch Logs に起動失敗や panic がない。

## Infrastructure Rollback

主な対象:

- Terraform 変更で ALB、ECS、RDS、IAM、security group が壊れた。
- apply 後に plan では想定できなかった AWS 側の制約が出た。

手順:

1. 直前に apply した PR、commit、terraform plan artifact を確認する。
2. `git revert` で Terraform configuration を戻す PR を作る。
3. `terraform plan` で戻し差分を確認する。
4. apply が必要な場合は、対象 resource、影響、概算コスト、戻し方の承認を得る。

```sh
terraform -chdir=infrastructure/app plan \
  -input=false \
  -var-file=envs/dev.tfvars
```

注意:

- RDS replacement、subnet replacement、security group replacement は停止や接続断を伴う可能性がある。
- IAM policy の rollback は deploy や ECS secret injection に影響する可能性がある。
- `terraform apply` は承認なしに実行しない。

## Database Migration Rollback

主な対象:

- schema migration 後に API が失敗する。
- data migration で想定外のデータ破損が起きた。

判断:

- backward compatible な migration なら application rollback を先に行う。
- `migrate down` がデータを失わない場合だけ down migration を検討する。
- destructive migration や data corruption の場合は snapshot restore または PITR を検討する。

手順:

1. 適用済み migration version と SQL を確認する。
2. backup/snapshot の取得時刻を確認する。
3. application rollback だけで復旧できるか確認する。
4. DB rollback が必要なら `docs/operations/rds-backup-restore.md` に従って新 DB へ restore する。

ローカル確認:

```sh
make migrate-down
make migrate-up
make test
```

## 記録

- rollback 開始時刻、完了時刻。
- 戻した task definition revision、image tag、commit。
- DB restore を行った場合は snapshot/PITR の復旧ポイント。
- 失われた可能性があるデータ範囲。
- 再発防止 action。
