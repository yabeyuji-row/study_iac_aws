# Postmortem Example: DB Secret Endpoint Mismatch

## Summary

- Incident title: DB restore 後の secret endpoint 更新漏れ。
- Date: 2026-08-13。
- Owner: TBD。
- Severity: SEV-2。
- Status: Exercise。

## Impact

- User impact: API が DB に接続できず、TODO CRUD が失敗する。
- Affected endpoints: `/readyz`, `/v1/todos`.
- Affected data: データ損失なし。restore 後 DB への切替が未完了。
- Start time: 2026-08-13T07:30:00Z。
- Detection time: 2026-08-13T07:32:00Z。
- Mitigation time: 2026-08-13T07:45:00Z。
- Recovery time: 2026-08-13T07:50:00Z。

## Timeline

| Time | Event |
| --- | --- |
| 2026-08-13T07:30:00Z | DB restore 完了後、ECS service を再 deploy。 |
| 2026-08-13T07:32:00Z | `/readyz` 503 と DB readiness failure alarm を確認。 |
| 2026-08-13T07:38:00Z | CloudWatch Logs で旧 DB endpoint への接続失敗を確認。 |
| 2026-08-13T07:45:00Z | Secrets Manager secret の endpoint を restored DB に更新し、ECS task を再起動。 |
| 2026-08-13T07:50:00Z | `/readyz` と TODO CRUD が復旧。 |

## Detection

- Triggered alarm: `api_db_readiness_failure`。
- Dashboard or log evidence: CloudWatch dashboard の DB readiness failure と API JSON logs。
- Who noticed: on-call。

## Root Cause

- Direct cause: RDS restore 後に Secrets Manager secret の `host` を更新していなかった。
- Contributing factors: restore 手順と deploy rollback 手順が分かれており、接続先切替の確認が checklist 化されていなかった。
- Why existing tests, alarms, or process did not catch it earlier: CI は AWS 上の restored DB endpoint までは検証しない。

## Mitigation

- Immediate action: secret の endpoint を restored DB に更新し、ECS service を新 task へ置き換えた。
- Rollback, restore, or config change: DB restore は維持し、application 接続先だけを修正した。
- Residual risk: secret 更新直後の task 再起動漏れで古い環境変数を使う task が残る可能性。

## Recovery Validation

- `/healthz`: 200。
- `/readyz`: 200。
- TODO CRUD: 成功。
- ECS service stable: stable。
- ALB target healthy: healthy。
- RDS metrics: DatabaseConnections が回復。

## Lessons

- What went well: readiness と alarm で DB 接続失敗を早く検知できた。
- What was confusing: restored DB endpoint と existing secret の対応関係が明記されていなかった。
- What was missing: restore 後の ECS task restart 確認。

## Action Items

| Action | Owner | Due Date | Status |
| --- | --- | --- | --- |
| RDS restore 手順に secret endpoint と ECS task restart の確認を追加する | TBD | TBD | Done |
| deploy rollback runbook から DB restore docs へリンクする | TBD | TBD | Done |
| DB readiness alarm の runbook に restore 後接続先確認を追加する | TBD | TBD | Done |

## Follow-up Links

- PR: TBD。
- Workflow run: TBD。
- Terraform plan: TBD。
- CloudWatch dashboard: `study-aws-dev-operations`。
- Runbook: [DB connection failure](../runbooks/db-connection-failure.md)。
