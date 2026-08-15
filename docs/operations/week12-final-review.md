# Week 12 Final Review

## Snapshot Restore Drill

想定:

- migration 直後に data corruption を検知。
- manual snapshot `study-aws-dev-postgres-before-migration-202608131700` へ戻す。
- 実 AWS resource は操作しない。

必要な入力値:

- 復元元 snapshot identifier。
- 復元先 DB instance identifier。
- DB subnet group `study-aws-dev-db-subnets`。
- RDS security group ID。
- parameter group `study-aws-dev-postgres`。
- 切替先 Secrets Manager secret。

想定停止時間:

- DB restore 完了待ち: 10-45 分。
- ECS task definition 更新と service stable 待ち: 5-15 分。
- 合計目安: 15-60 分。

確認結果:

- snapshot restore は新 DB instance 作成になるため、既存 DB への上書きではない。
- 接続先切替は Secrets Manager secret 更新または新 secret を参照する task definition 更新で行う。
- 旧 DB と snapshot を削除する前に復旧確認を完了する。

## PITR Drill

想定:

- 誤削除が 2026-08-13T07:30:00Z に発生。
- 最後に正常だった時刻を 2026-08-13T07:25:00Z と仮定。

確認結果:

- `db_backup_retention_days = 7` のため、retention 内の時刻なら PITR の対象になる。
- 復旧ポイント以降の書き込みは失われる可能性がある。
- 復旧後は new DB endpoint を secret/task definition に反映して ECS を再 deploy する。

## ECS Rollback Drill

想定:

- 新 image が `/readyz` で失敗し ALB target unhealthy になる。

確認結果:

- ECS deployment circuit breaker は `rollback = true`。
- 自動 rollback しない場合は、直前 task definition revision を指定して `aws ecs update-service` を実行する。
- rollback 後は ALB target、`/readyz`、CloudWatch Logs を確認する。

## Terraform Drift Drill

想定:

- console から ALB security group に意図しない ingress が追加された。

確認結果:

- 実 AWS 環境では `terraform plan -refresh-only -detailed-exitcode` で drift を確認する。
- セキュリティ影響がある drift は先に緩和し、その後 Terraform configuration と state を整合させる。
- 恒久変更なら Terraform configuration に反映し、PR で review する。

## Postmortem Exercise

仮想障害:

- Deploy 後、DB secret の接続先が古い endpoint のままで `/readyz` が 503 になった。

記入例:

- [Postmortem example: DB secret endpoint mismatch](postmortem-example-db-secret-endpoint.md)

再発防止 action:

| Action | Owner | Due Date | Status |
| --- | --- | --- | --- |
| deploy 前 checklist に secret endpoint 確認を追加する | TBD | TBD | Open |
| DB restore 後の ECS task definition 更新手順を runbook から参照する | TBD | TBD | Open |
| `/readyz` 503 alarm から DB restore docs へ誘導する | TBD | TBD | Open |

## Cleanup Review

確認対象:

- ALB、target group、ECS service、task、RDS instance、snapshot、Secrets Manager secret、CloudWatch Logs、alarms、dashboards、ECR images、NAT Gateway、Elastic IP。

確認結果:

- 12週目では実 AWS resource の create、restore、destroy は行っていない。
- AWS apply 後は `docs/operations/cleanup-checklist.md` に沿って残課金 resource を確認する。

## AWS Follow-up Tasks

12週目までの完了チェックは、ローカル、MiniStack、GitHub、または机上演習で確認できる範囲を示す。
次の項目は AWS 実環境でのみ本番相当の確認ができるため、未完了の follow-up として扱う。

| Task | Source | Status |
| --- | --- | --- |
| VPC、subnet、route table、security group、IAM role の実 AWS 反映確認 | `docs/learning-roadmap.md` 7-4 | Open |
| ECR push、ECS Fargate 起動、ALB target health、CloudWatch Logs 出力確認 | `docs/learning-roadmap.md` 8-4 | Open |
| RDS private 配置、Secrets Manager 経由の DB 接続、backup 設定確認 | `docs/learning-roadmap.md` 9-4 | Open |
| GitHub Actions OIDC、ECR push、manual gated ECS deploy の実 AWS 確認 | `docs/learning-roadmap.md` 10-4 | Open |
| CloudWatch alarm、dashboard、runbook の実 AWS 確認 | `docs/learning-roadmap.md` 11-4 | Open |
| RDS snapshot/restore/PITR、drift、cost、cleanup の実 AWS 確認 | `docs/learning-roadmap.md` 12-4 | Open |

AWS follow-up を実行する前に、対象 resource、概算コスト、停止影響、戻し方、削除手順を確認し、明示的な承認を得る。
