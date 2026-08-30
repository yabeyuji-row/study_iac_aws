# Runbook: Local Observability Lab Failure

## 検知

- Grafana で 5xx ratio または p95 latency が閾値を超える。
- `/readyz` が 503 を返す。
- k6 threshold が失敗する。
- Tempo に fault endpoint の trace が記録される。

## 一次確認

- 演習開始時刻と実行した `make fault-*` target を記録する。
- `curl -i http://localhost:8080/healthz` と `/readyz` を比較する。
- Prometheus targets で `todo-api` が UP か確認する。
- `docker compose -f compose.yaml -f compose.observability.yaml ps` で各 service を確認する。

## 切り分け

- 5xx だけなら status 別 metric、API log、Tempo trace の同じ時刻を見る。
- latency だけなら route 別 p95 と slow trace を見る。
- `/healthz` は成功し `/readyz` が失敗するなら PostgreSQL の状態を確認する。
- metrics が欠落し API が正常なら Prometheus scrape、Collector、Grafana datasource を確認する。

## 緩和

- 5xx と delay は単発 request の終了を待つ。
- DB 停止なら `make fault-db-recover` を実行する。
- Lab component の設定不備なら `make observability-down` 後に構成を直して再起動する。

## 復旧確認

- `/healthz` と `/readyz` が 200 を返す。
- `make loadtest` が threshold を満たす。
- Prometheus target が UP で、新しい trace が Tempo に到着する。
- 5xx rate と p95 latency が通常値へ戻る。

## 事後対応

- detection time、mitigation time、recovery time と error budget 消費を記録する。
- logs、metrics、traces のうち診断に不足した signal を記録する。
- [Postmortem exercise](../operations/phase6-postmortem-exercise.md) に原因と action を反映する。
