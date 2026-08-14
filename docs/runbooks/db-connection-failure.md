# Runbook: DB Connection Failure

## 検知

- `/readyz` が 503 を返す。
- CloudWatch alarm `study-aws-dev-alarm-api-db-readiness-failure` が ALARM になる。
- ALB target group が target unhealthy を示す。
- App log に DB ping、DB open、repository error が出る。

## 一次確認

- `/healthz` が 200 か確認し、process 自体が生きているか分ける。
- `/readyz` の 503 が継続しているか確認する。
- RDS instance status、CPU、DatabaseConnections、FreeStorageSpace を確認する。
- Secrets Manager secret の ARN と ECS task definition の `DATABASE_SECRET_JSON` を確認する。

## 切り分け

- secret 欠落や JSON 不備なら config validation failure として扱う。
- RDS security group ingress が ECS security group からの `db_port` のみを許可しているか確認する。
- ECS task が private subnet から Secrets Manager/RDS へ到達できる設計か確認する。
- DB password 変更や rotation があったか確認する。

## 緩和

- secret ARN 誤設定なら Terraform 管理の `aws_secretsmanager_secret.db.arn` に戻す。
- security group 誤変更なら `rds_from_ecs` ingress を復元する。
- DB 高負荷なら不要な接続元を止め、必要なら task desired count を一時的に増減する前に接続上限を確認する。
- RDS 障害なら [RDS backup and restore](../operations/rds-backup-restore.md) の snapshot/PITR restore 手順へ切り替える。

## 復旧確認

- `/readyz` が 200 を返す。
- ALB target が healthy になる。
- `DBReadinessFailure` metric が増えなくなる。
- CRUD の代表操作が成功する。

## 事後対応

- secret、network、RDS resource、migration のどこで壊れたかを記録する。
- config validation や Terraform validation で早期検出できる余地を確認する。
- backup/restore が必要だった場合は [RDS backup and restore](../operations/rds-backup-restore.md) と postmortem に反映する。
