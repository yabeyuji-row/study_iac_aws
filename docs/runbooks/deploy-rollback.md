# Runbook: Deploy Rollback

## 検知

- Deploy workflow の `Wait for service stability` が失敗する。
- ECS deployment circuit breaker が rollback を開始する。
- ALB target unhealthy、ALB target 5xx、DB readiness failure のいずれかが ALARM になる。

## 一次確認

- ECS service events で deployment failure reason を確認する。
- 現在の task definition revision と直前 revision を確認する。
- 新 image tag、commit、build metadata を確認する。
- CloudWatch Logs で起動失敗、config validation failure、DB 接続失敗を確認する。

## 切り分け

- image pull 失敗なら ECR tag、repository URL、task execution role を確認する。
- container health check 失敗なら `/healthz`、port mapping、command を確認する。
- ALB health check 失敗なら `/readyz`、RDS、Secrets Manager、security group を確認する。
- migration 直後なら migration と application compatibility を確認する。

## 緩和

- ECS circuit breaker が自動 rollback した場合は、service が stable になるまで待つ。
- 手動 rollback が必要な場合は直前 revision を指定する。詳細は [Rollback procedures](../operations/rollback.md) を参照する。

```sh
aws ecs update-service \
  --cluster study-aws-dev-cluster \
  --service study-aws-dev-api-service \
  --task-definition study-aws-dev-api:<previous-revision>

aws ecs wait services-stable \
  --cluster study-aws-dev-cluster \
  --services study-aws-dev-api-service
```

- migration 起因で backward incompatible な場合は、[RDS backup and restore](../operations/rds-backup-restore.md) の database restore 手順の必要性を判断する。

## 復旧確認

- ECS service が stable になる。
- ALB target が healthy になる。
- `/readyz` が 200 を返す。
- CloudWatch Logs に新しい起動失敗が出ていない。
- 代表的な TODO CRUD 操作が成功する。

## 事後対応

- rollback した task definition revision、image tag、原因 commit を記録する。
- deploy gate、migration 手順、CI coverage の不足を確認する。
- 失敗した revision を再 deploy しないように GitHub workflow run と ECR image tag を確認する。
