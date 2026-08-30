# Observability and Reliability Lab

## 目的

Phase 6 は、監視製品を並べることではなく、負荷または障害を発生させ、検知し、
runbook で緩和し、SLO への影響と改善案を説明できる状態を作る。
すべてローカルで完結させ、既存 AWS/Terraform resource は変更しない。
略語や監視用語は [用語集](../glossary.md) を参照する。

## 構成

```text
k6 / curl -> Go API -> PostgreSQL
                |-- /metrics -> Prometheus -> Grafana
                `-- OTLP trace -> OpenTelemetry Collector -> Tempo -> Grafana
```

JSON structured log、`/healthz`、`/readyz` は既存仕様を維持する。Prometheus metrics は
CloudWatch の代替ではなく、同じ SLI をローカルで検証するための信号とする。

## OpenTelemetry 方針

- `OTEL_EXPORTER_OTLP_ENDPOINT` が空なら exporter を作らず、通常起動の外部依存を増やさない。
- Lab では HTTP server request を span にし、OTLP/HTTP で Collector へ送る。
- service name は `todo-api` とする。
- password、connection string、request body、query string を attribute に含めない。
- route template を attribute と metric label に使い、TODO ID などの高カーディナリティ値を避ける。
- exporter 停止が API availability を下げないよう batch export とする。
- 次の段階で repository/DB span、trace ID 付き log、sampling policy を追加する。

## Metrics

| Metric | Labels | 用途 |
| --- | --- | --- |
| `todo_api_http_requests_total` | `method`, `route`, `status` | traffic、availability、5xx rate |
| `todo_api_http_request_duration_seconds` | `method`, `route` | p50/p95/p99 latency |

`/metrics`、health/readiness、fault endpoint は dashboard の user traffic 集計から除外する。
raw path、query、request ID は label にしない。

## SLI、SLO、Error Budget

ローカルの短時間 window は SLO 達成を証明するものではなく、式と検知経路を検証する。
長期目標は既存の [Observability and SLO Design](observability-slo.md) と同じにする。

| SLI | PromQL | 30日 SLO |
| --- | --- | --- |
| Availability | `1 - 5xx requests / all user requests` | 99.0% |
| Latency | request duration histogram の p95 | 500 ms 未満 |
| Error rate | `5xx requests / all user requests` | 1% 未満 |
| DB readiness | `/readyz` の 503 継続時間 | 5分連続を発生させない |

99.0% availability の30日 error budget は約7時間12分。演習では障害開始から復旧までを
budget 消費として記録し、消費が速い場合は feature work より復旧と再発防止を優先する。

## Alert と Dashboard

初期 alert 候補:

- warning: 5分 window の 5xx rate が1%を超える。
- critical: 5分 window の 5xx rate が5%を超える。
- warning: 5分 window の p95 latency が500 msを超える。
- critical: `/readyz` 503 が2分継続する。
- collector/exporter 障害は telemetry loss として扱い、API 障害と分ける。

Grafana dashboard は request rate、status 別 request、5xx ratio、route 別 p95 latency を
最小構成とする。短い演習で確認しやすいよう query window は1-5分にする。

## 安全な障害注入

`FAULT_INJECTION_ENABLED=true` の場合だけ `/debug/fault/*` を登録し、各 request に
`X-Fault-Injection: enabled` を要求する。通常の `compose.yaml`、AWS task definition、
既定値では無効である。

| Scenario | 操作 | 期待結果 | 復旧 |
| --- | --- | --- | --- |
| 5xx | `make fault-5xx` | 500、5xx counter と trace 増加 | request 終了時に復旧済み |
| latency | `make fault-delay FAULT_DELAY=2s` | p95 latency 上昇 | request 終了時に復旧済み |
| DB unavailable | `make fault-db-down` | `/healthz` 200、`/readyz` 503 | `make fault-db-recover` |

遅延は最大5秒に制限する。disk fill、packet loss、CPU exhaustion、data corruption は
初期 Lab では扱わない。

## 実行手順

```bash
make observability-up
make loadtest
make fault-5xx
make fault-delay FAULT_DELAY=2s
make fault-db-down
curl -i http://localhost:8080/healthz
curl -i http://localhost:8080/readyz
make fault-db-recover
make observability-down
```

Grafana の `TODO API / TODO API Observability Lab` dashboard と Explore の Tempo を確認する。
演習では開始時刻、検知時刻、緩和開始、復旧時刻、参照した signal を記録する。

## AWS への対応関係

Prometheus request/latency は ALB metrics と CloudWatch log-derived metrics、Grafana panel は
CloudWatch dashboard、local alert candidate は CloudWatch alarm に対応する。Tempo は将来
ADOT/X-Ray 等へ置き換えられるが、Phase 6 では Terraform や AWS resource を変更しない。
