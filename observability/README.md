# Observability

このディレクトリには、TODO API の metrics と traces をローカルで収集・可視化するための設定を置いています。
負荷試験や安全な障害注入を通して、異常の検知、原因の切り分け、復旧確認を練習するための構成です。

## 構成

```text
k6 / curl -> Go API -> PostgreSQL
                |-- /metrics -> Prometheus -> Grafana
                `-- OTLP trace -> OpenTelemetry Collector -> Tempo -> Grafana
```

| パス | 役割 |
| --- | --- |
| `prometheus.yml` | TODO API の `/metrics` を5秒間隔で収集する Prometheus 設定 |
| `otel-collector.yaml` | API から OTLP/HTTP で受信した trace を Tempo に転送する設定 |
| `tempo.yaml` | trace をローカルに保存する Tempo 設定（保持期間は1時間） |
| `grafana/provisioning/datasources/datasources.yaml` | Prometheus と Tempo を Grafana のデータソースとして登録する設定 |
| `grafana/provisioning/dashboards/dashboards.yaml` | dashboard の自動読み込み設定 |
| `grafana/dashboards/todo-api.json` | request rate、5xx rate、p95 latency を表示する dashboard |

コンテナの定義はリポジトリ直下の `compose.observability.yaml`、操作用コマンドは
`make/observability.mk` にあります。

## 起動

Docker と Docker Compose が利用できる状態で、リポジトリのルートから実行します。

```bash
make observability-up
```

このコマンドは PostgreSQL、マイグレーション、API、Prometheus、Grafana、Tempo、
OpenTelemetry Collector を起動します。

| 対象 | URL |
| --- | --- |
| TODO API | <http://localhost:8080> |
| Prometheus | <http://localhost:9090> |
| Grafana | <http://localhost:3000> |

Grafana はローカル Lab 内で匿名 Editor として利用できます。`TODO API` フォルダの
`TODO API Observability Lab` dashboard を開いて signal を確認してください。Editor 権限は
Tempo の trace を Explore で確認するために使用します。

## 動作確認

### 1. 通常負荷

```bash
make loadtest
```

k6 は5 VUで30秒間、TODO 一覧 API に request を送ります。実行結果と Grafana で次を確認します。

- k6 の `checks` に失敗がなく、`status is 200` が100%である。
- `http_req_failed` が threshold の1%未満である。
- `http_req_duration` の p95 が threshold の500 ms未満である。
- Grafana の `Request rate` に `/v1/todos`、status `200` の traffic が表示される。
- Grafana の `5xx rate` が0またはほぼ0である。
- Grafana の `Latency p95` が0.5秒未満である。

Grafana のグラフが表示されない場合は、右上の時間範囲を `Last 5 minutes` にして更新します。
Prometheus の <http://localhost:9090/targets> では `todo-api` が `UP` であることを確認します。

### 2. 5xx 障害

```bash
make fault-5xx
```

次を確認します。

- コマンドの出力が `status=500` である。
- Grafana の `Request rate` に status `500` の request が表示される。
- Grafana の `5xx rate` が一時的に上昇する。
- Grafana の Explore で Tempo を選択し、`/debug/fault/5xx` の trace を確認できる。

単発 request の終了時点で障害は解消されるため、復旧コマンドは不要です。

### 3. 遅延障害

```bash
make fault-delay FAULT_DELAY=2s
```

次を確認します。

- response が返るまで約2秒かかる。
- Grafana の `Latency p95` が一時的に上昇する。
- Grafana の Explore で Tempo を選択し、遅延した `/debug/fault/delay` の trace を確認できる。

遅延は単発 request の終了時点で解消されます。`FAULT_DELAY` は最大5秒です。

### 4. DB 停止障害

PostgreSQL を停止します。

```bash
make fault-db-down
```

API の生存状態と request を処理できる準備状態を確認します。

```bash
curl -i http://localhost:8080/healthz
curl -i http://localhost:8080/readyz
```

期待する結果は次のとおりです。

| Endpoint | HTTP status | 意味 |
| --- | --- | --- |
| `/healthz` | `200` | API process は動作している |
| `/readyz` | `503` | PostgreSQL に接続できず、request を処理する準備ができていない |

確認後、PostgreSQL を復旧します。

```bash
make fault-db-recover
```

少し待ってから readiness を再確認します。

```bash
curl -i http://localhost:8080/readyz
```

HTTP status が `200` に戻れば復旧成功です。

### 5. 最終確認

```bash
make loadtest
```

次を確認します。

- k6 の failure rate が1%未満、p95 latency が500 ms未満である。
- Grafana の `5xx rate` と `Latency p95` が通常値へ戻る。
- Prometheus の `todo-api` target が `UP` である。
- Grafana の Explore から Tempo に新しい trace が届いている。

5xx と遅延の endpoint は Observability 構成でのみ有効になり、専用ヘッダーを要求します。

サービスの状態やログは次のコマンドで確認できます。

```bash
docker compose -f compose.yaml -f compose.observability.yaml ps
make observability-logs
```

## 終了

```bash
make observability-down
```

## 関連ドキュメント

- [Observability and Reliability Lab](../docs/design/observability-reliability-lab.md): metrics、SLI/SLO、障害シナリオの設計
- [Local Observability Lab Runbook](../docs/runbooks/local-observability-lab.md): 検知から復旧、事後対応までの手順
- [Observability and SLO Design](../docs/design/observability-slo.md): AWS 環境を含む長期的な監視設計
