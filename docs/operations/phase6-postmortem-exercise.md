# Phase 6 Postmortem Exercise

## Scenario

k6 実行中に PostgreSQL を停止した。API process は生存していたが `/readyz` が 503 となり、
TODO request が 5xx を返した。これはローカル演習であり、実ユーザーと AWS resource への影響はない。

## Timeline Template

| Time | Event | Evidence |
| --- | --- | --- |
| T+00 | `make fault-db-down` | shell history |
| T+XX | readiness failure を検知 | `/readyz`、request metric |
| T+XX | DB 起因と判断 | PostgreSQL container、API log、trace |
| T+XX | `make fault-db-recover` | Compose event |
| T+XX | readiness と負荷試験が回復 | `/readyz`、k6、Grafana |

## Analysis Questions

- `/healthz` と `/readyz` の違いから、どこまで故障していると判断できたか。
- 最初に役立った signal と、ノイズになった signal は何か。
- 検知と緩和に何分かかり、30日 error budget の何%を消費したか。
- DB 復旧直後の connection retry、pool、request error はどう変化したか。
- 実 AWS なら ALB、ECS、RDS、CloudWatch のどの signal に置き換わるか。

## Expected Root Cause

意図的に PostgreSQL container を停止したため、DB ping と repository query が失敗した。
API process 自体は動作しているため liveness は成功し、readiness だけが失敗する設計どおりの挙動である。

## Corrective Actions

| Action | Owner | Due Date | Status |
| --- | --- | --- | --- |
| readiness failure の継続時間を記録できる panel を追加する | TBD | TBD | Open |
| DB query span と pool metrics の必要性を評価する | TBD | TBD | Open |
| multi-window burn-rate alert の rule test を追加する | TBD | TBD | Open |

## Completion

[Postmortem template](../templates/postmortem-template.md) に実測時刻、スクリーンショットまたは
PromQL、判断、action owner を転記して完了とする。
