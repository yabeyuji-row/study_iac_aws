# Observability and SLO Design

## Scope

11週目では、CloudWatch Logs、CloudWatch Metrics、CloudWatch Alarms、CloudWatch Dashboard を使って、
API、ALB、ECS、RDS の初期 observability を作る。

AWS 公式ドキュメントでは、Application Load Balancer は `AWS/ApplicationELB` metrics を CloudWatch に発行し、
RDS は 1 分間隔の CloudWatch metrics を発行する。
ECS の `RunningTaskCount` は Container Insights の `ECS/ContainerInsights` metric として扱う。
CloudWatch Logs metric filter は JSON log から custom metric を作るために使う。

## Structured Logging

Go API は request 完了時に JSON log を標準出力へ出す。
ECS task definition の `awslogs` driver がこのログを CloudWatch Logs log group へ送る。

必須項目:

- `time`: log timestamp。
- `level`: log level。
- `msg`: log message。
- `event`: metric filter 用 event name。
- `method`: HTTP method。
- `path`: query string を含まない path。
- `status`: HTTP status code。
- `duration_ms`: request latency in milliseconds。
- `remote_addr`: client address。
- `user_agent`: user agent。
- `request_id`: `X-Request-Id` header value。

禁止事項:

- DB password、secret JSON、`DATABASE_URL`、AWS credential を log に出さない。
- request query string は token を含み得るため metric log には含めない。

## SLI

Availability:

- Primary signal: ALB target health and target 5xx。
- Initial SLI: `1 - (HTTPCode_Target_5XX_Count / RequestCount)`。

Latency:

- Primary signal: ALB `TargetResponseTime`。
- App signal: JSON request log 由来の `LatencyMs` average。

Error rate:

- Primary signal: ALB `HTTPCode_Target_5XX_Count`。
- App signal: JSON request log 由来の `ErrorCount`。

DB readiness:

- Primary signal: `/readyz` が DB Ping に失敗して `503` を返す回数。
- App signal: JSON request log 由来の `DBReadinessFailure`。

## Initial SLO

この環境は学習用 dev environment のため、初期 SLO は厳しすぎない値にする。

- Availability SLO: 30日 rolling window で 99.0%。
- Latency SLO: 95% の request が 500 ms 未満。
- Error rate SLO: 5xx response が全 request の 1% 未満。
- DB readiness SLO: 5分連続の readiness failure を 0 件に近づける。

30日 99.0% availability の error budget は約 7時間12分。
この予算を超えそうな場合は、新規 deploy より復旧と原因分析を優先する。

## Alarm Policy

| Alarm | Metric | Threshold | Evaluation | Missing data | Reason |
| --- | --- | --- | --- | --- | --- |
| ALB target 5xx | `HTTPCode_Target_5XX_Count` sum | 5以上 | 1分 x 5回中3回 | not breaching | 低 traffic の一時的な失敗で鳴りすぎないようにする。 |
| ALB unhealthy targets | `UnHealthyHostCount` maximum | 1以上 | 1分 x 3回中2回 | not breaching | desired count 1 の dev では 1 target unhealthy が影響大。 |
| ECS running task count low | `RunningTaskCount` minimum | desired count 未満 | 1分 x 3回中2回 | breaching | task が 0 のとき metric 欠落も障害として扱う。 |
| RDS CPU high | `CPUUtilization` average | 80%以上 | 5分 x 3回中2回 | missing | 短い spike は様子見し、継続高負荷を見る。 |
| RDS free storage low | `FreeStorageSpace` average | 2 GiB 未満 | 5分 x 3回中2回 | missing | 小さな dev DB でも復旧猶予を残す。 |
| RDS connections high | `DatabaseConnections` average | 40以上 | 5分 x 3回中2回 | missing | app pool と運用接続の余裕を残す。 |
| DB readiness failure | `DBReadinessFailure` sum | 1以上 | 1分 x 3回中2回 | not breaching | DB 接続不能が継続したときだけ検知する。 |

Alarm action は 11週目では未設定にする。
通知先 SNS topic、on-call、escalation は後続の SRE/security フェーズで決める。

## Dashboard

初期 dashboard は 4 つの widget に分ける。

- ALB traffic and errors: `RequestCount`、`HTTPCode_Target_5XX_Count`、`TargetResponseTime`。
- Target health and ECS tasks: `HealthyHostCount`、`UnHealthyHostCount`、`RunningTaskCount`。
- API log-derived metrics: `RequestCount`、`ErrorCount`、`DBReadinessFailure`、`LatencyMs`。
- RDS health: `CPUUtilization`、`DatabaseConnections`、`FreeStorageSpace`。

## Cost Notes

Terraform plan だけでは AWS cost は発生しない。
apply した場合は CloudWatch alarm、dashboard、custom metrics、Container Insights、CloudWatch Logs ingestion/storage に追加料金が発生する。
apply 前に作成 resource と概算コストを報告し、明示承認を得る。

## Failure Drill Notes

Target unhealthy:

- `health_check_path` を存在しない path に一時変更すると、ALB target group は task を unhealthy と判定する想定。
- Terraform plan では target group health check path の差分として見える。

High error rate:

- API が 500 を返す想定では、ALB `HTTPCode_Target_5XX_Count` と app log `ErrorCount` が増える。
- 11週目では実障害注入 endpoint を追加せず、既存 handler error と runbook で机上確認する。

DB readiness failure:

- DB 接続失敗時は `/readyz` が `503` を返す。
- ALB target unhealthy と app log `DBReadinessFailure` の両方で検知する。

## References

- AWS Application Load Balancer metrics: https://docs.aws.amazon.com/elasticloadbalancing/latest/application/load-balancer-cloudwatch-metrics.html
- AWS RDS CloudWatch metrics: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/monitoring-cloudwatch.html
- AWS ECS Container Insights metrics: https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Container-Insights-metrics-ECS.html
- CloudWatch Logs filter pattern syntax: https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html
