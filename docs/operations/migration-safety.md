# Migration Safety

## 方針

application rollback と database rollback を独立させる。通常 deploy では expand/contract を使い、
旧 application と新 application が同じ schema で並行稼働できる期間を設ける。

## 手順

1. Expand: nullable column、互換 index、新 table など backward-compatible な変更を先に適用する。
2. Deploy: 新旧 schema の両方を扱える application を deploy する。
3. Backfill: timeout、batch size、再実行可能性、監視方法を決めて data を移す。
4. Observe: error rate、latency、DB connections、lock、readiness を確認する。
5. Contract: 旧 application が存在しないことを確認してから古い column や constraint を削除する。

## Review Checklist

- migration は transaction 内で安全に実行できるか。
- table rewrite、長時間 lock、full scan、disk 増加が起きないか。
- API rollback 後も schema compatibility が保たれるか。
- down migration が data loss を起こす場合、その事実と restore 手順が明記されているか。
- backup/PITR の復旧可能時刻と RTO/RPO を確認したか。
- migration と application deploy の担当、停止条件、観測 query を決めたか。

## Stop And Rollback Criteria

- 5xx rate が1%を継続して超える。
- p95 latency が500 msを継続して超える。
- readiness failure、lock wait、connection exhaustion が発生する。
- backfill の結果件数または checksum が期待値と一致しない。

backward-compatible なら application を先に rollback する。destructive migration は安易に
`make migrate-down` せず、[Rollback Procedures](rollback.md) と
[RDS backup and restore](rds-backup-restore.md) に従って restore の要否を判断する。
