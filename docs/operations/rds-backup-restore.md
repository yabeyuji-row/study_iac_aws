# RDS Backup And Restore

## 方針

このプロジェクトの dev 環境では、学習とコスト管理を優先しつつ、誤削除や deploy 失敗から戻れる最小限の復旧点を残す。
AWS 実リソースを操作する場合は、対象 resource、想定停止時間、概算コスト、戻し方を確認してから明示的な承認を得る。

参考:

- Amazon RDS automated backups: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_WorkingWithAutomatedBackups.html
- Amazon RDS snapshot restore: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_RestoreFromSnapshot.html
- Amazon RDS point-in-time restore: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_PIT.html
- Amazon RDS maintenance window: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_UpgradeDBInstance.Maintenance.html

## 現在の Terraform 設定

`infrastructure/app/aws-rds.tf` の初期方針:

- Automated backup retention: `db_backup_retention_days`。dev は `7` 日。
- Backup window: `17:00-18:00` UTC。
- Maintenance window: `sun:18:00-sun:19:00` UTC。
- Storage encryption: enabled。
- Final snapshot: destroy 時に取得する。
- Deletion protection: dev は `false`。
- RDS は private subnet に置き、public access は無効。

Backup window と maintenance window は重ならないようにする。
RDS maintenance は window 内に開始されるが、作業が window 終了後まで続く可能性がある。

## Retention

dev の retention は 7 日を初期値にする。
本番相当では RPO、規制、復旧演習の結果、backup storage cost を確認してから 14-35 日へ延長を検討する。

確認コマンド:

```sh
aws rds describe-db-instances \
  --db-instance-identifier study-aws-dev-postgres \
  --query 'DBInstances[0].{BackupRetentionPeriod:BackupRetentionPeriod,PreferredBackupWindow:PreferredBackupWindow,PreferredMaintenanceWindow:PreferredMaintenanceWindow,DeletionProtection:DeletionProtection}'
```

## Manual Snapshot

deploy、migration、破壊的な Terraform 変更の前には manual snapshot を検討する。

入力値:

- `DB_INSTANCE_IDENTIFIER`: 例 `study-aws-dev-postgres`
- `SNAPSHOT_IDENTIFIER`: 例 `study-aws-dev-postgres-before-migration-YYYYMMDDHHMM`
- 対象 commit、migration version、作業者、目的

作成:

```sh
aws rds create-db-snapshot \
  --db-instance-identifier "$DB_INSTANCE_IDENTIFIER" \
  --db-snapshot-identifier "$SNAPSHOT_IDENTIFIER"

aws rds wait db-snapshot-completed \
  --db-snapshot-identifier "$SNAPSHOT_IDENTIFIER"
```

確認:

```sh
aws rds describe-db-snapshots \
  --db-snapshot-identifier "$SNAPSHOT_IDENTIFIER" \
  --query 'DBSnapshots[0].{Status:Status,SnapshotCreateTime:SnapshotCreateTime,Encrypted:Encrypted}'
```

## Snapshot Restore

RDS snapshot restore は既存 DB instance へ上書き復元しない。
snapshot から新しい DB instance を作成し、検証後にアプリの接続先を切り替える。

入力値:

- 復元元 snapshot identifier
- 復元先 DB instance identifier
- subnet group、security group、parameter group
- instance class、storage、backup retention
- 切替対象の secret または task definition

手順:

```sh
aws rds restore-db-instance-from-db-snapshot \
  --db-instance-identifier "$RESTORED_DB_INSTANCE_IDENTIFIER" \
  --db-snapshot-identifier "$SNAPSHOT_IDENTIFIER" \
  --db-subnet-group-name study-aws-dev-db-subnets \
  --vpc-security-group-ids "$RDS_SECURITY_GROUP_ID" \
  --db-instance-class db.t4g.micro \
  --no-publicly-accessible

aws rds wait db-instance-available \
  --db-instance-identifier "$RESTORED_DB_INSTANCE_IDENTIFIER"
```

復元後:

1. restored DB の endpoint を確認する。
2. Secrets Manager の DB secret を新 endpoint に更新するか、新 secret を作成して task definition を差し替える。
3. ECS service を新しい task definition revision へ更新する。
4. `/readyz`、代表 TODO CRUD、CloudWatch Logs を確認する。
5. 旧 DB を削除する前に必要な snapshot と rollback 先を確認する。

想定停止時間は、DB restore 完了待ち、ECS 新 task 起動、ALB health check 通過に依存する。
dev では机上見積もりとして 15-60 分を初期想定にする。

## PITR

PITR は automated backup retention が 1 日以上であることが前提になる。
復旧ポイントは、障害発生時刻、最後に正常だった時刻、誤操作時刻、ログとユーザー影響を照合して決める。

入力値:

- `SOURCE_DB_INSTANCE_IDENTIFIER`
- `RESTORED_DB_INSTANCE_IDENTIFIER`
- `RESTORE_TIME`。UTC で記録する。
- DB subnet group、security group、parameter group
- 接続先切替の手順と rollback 先

手順:

```sh
aws rds restore-db-instance-to-point-in-time \
  --source-db-instance-identifier "$SOURCE_DB_INSTANCE_IDENTIFIER" \
  --target-db-instance-identifier "$RESTORED_DB_INSTANCE_IDENTIFIER" \
  --restore-time "$RESTORE_TIME" \
  --db-subnet-group-name study-aws-dev-db-subnets \
  --vpc-security-group-ids "$RDS_SECURITY_GROUP_ID" \
  --no-publicly-accessible

aws rds wait db-instance-available \
  --db-instance-identifier "$RESTORED_DB_INSTANCE_IDENTIFIER"
```

切替後は、復旧対象データの整合性、`/readyz`、代表 TODO CRUD、RDS metrics を確認する。
復旧ポイントより後の書き込みは失われる可能性があるため、影響範囲を postmortem に記録する。

## Deletion Protection

dev では、学習中の cleanup を優先して deletion protection を無効にする。
本番相当では deletion protection を有効にし、Terraform destroy や DB 削除の前に明示的な change approval を必須にする。

dev でも `terraform destroy` 前は final snapshot を取得する。
同名 final snapshot が残っていると次回 destroy に失敗する可能性があるため、cleanup checklist で snapshot 名を確認する。

## 机上演習メモ

- Snapshot restore は新 DB を作るため、接続先切替が必要。
- PITR は復旧ポイント以降の書き込みを失う可能性がある。
- Secrets Manager secret 更新後は ECS task の再起動または新 task definition 反映が必要。
- 実 AWS 操作は未実施。MiniStack は RDS の restore/PITR 本番相当確認には使わない。
