# 用語集

## 目的

この用語集は、TODO API の設計、AWS/IaC、CI/CD、オブザーバビリティ、SRE、
障害対応を説明するときに使う用語を、このリポジトリでの意味に合わせて整理する。

## API と Go

| 用語 | 説明 |
| --- | --- |
| REST API | HTTP method と URL で resource を操作する API。このリポジトリでは `/v1/todos` に対する作成、一覧、取得、更新、削除を提供する。 |
| Echo v4 | Go の HTTP web framework。routing と middleware に使い、service と repository は Echo に依存させない。 |
| Handler | HTTP request の解析、validation error の変換、response の書き込みを担当する層。 |
| Service | TODO の validation や business rule を担当する層。HTTP framework と DB 実装から分離する。 |
| Repository | データの保存と取得を抽象化する層。このリポジトリでは PostgreSQL と SQL の詳細を担当する。 |
| Middleware | 複数の endpoint に共通処理を適用する仕組み。request log、metrics、tracing で使用する。 |
| `context.Context` | request のキャンセル、deadline、trace context を処理の呼び出し先へ伝播する Go の仕組み。 |
| Graceful Shutdown | 停止 signal を受けた後、新規受付を止め、処理中 request の完了を一定時間待って終了すること。 |
| Cursor Pagination | 最後に取得した行の位置を cursor として次ページを取得する方式。この API では `created_at` と `id` を使い、同時刻のデータがあっても順序を安定させる。 |
| Optimistic Locking | 更新時に `version` が読み取り時と同じか確認し、同時更新による上書きを防ぐ方式。 競合時は `409 Conflict` を返す。 |

## PostgreSQL と Migration

| 用語 | 説明 |
| --- | --- |
| Connection Pool | DB connection を再利用する仕組み。ECS task 数と task ごとの最大接続数を掛けた値が、 PostgreSQL の接続上限を超えないようにする。 |
| Database Constraint | 不正なデータを DB 自身が拒否するための制約。`NOT NULL`、`CHECK`、primary key などを指す。 |
| Migration | DB schema を version 管理して変更すること。このリポジトリでは SQL file を明示的に実行し、 API 起動時には自動適用しない。 |
| Expand/Contract Migration | 互換性のある schema を先に追加し、application 切替と data 移行後に古い schema を削除する方式。 新旧 application が並行稼働できる期間を作り、rollback risk を抑える。 |
| Backfill | 新しい column や table に既存 data を後から移す処理。batch size、再実行可能性、lock、 完了確認を設計する必要がある。 |
| PITR | Point-in-Time Recovery。backup retention 内の指定時刻へ DB を復元する仕組み。 復旧時刻より後の書き込みは失われる可能性がある。 |
| RPO | Recovery Point Objective。障害時に、どの時点までの data loss を許容するかという目標。 |
| RTO | Recovery Time Objective。障害発生から service 復旧までに許容する時間の目標。 |

## Container と AWS

| 用語 | 説明 |
| --- | --- |
| Container Image | application と実行に必要な file をまとめた不変の成果物。ECR に保存し、ECS task から実行する。 |
| Docker Compose | 複数 container の構成をローカルで定義して起動する仕組み。API、PostgreSQL、監視 Lab に使う。 |
| ECR | Amazon Elastic Container Registry。Docker image を保存する AWS service。 |
| ECS Fargate | server を直接管理せずに container を実行する AWS の構成。このリポジトリでは Go API を実行する。 |
| Task Definition | ECS container の image、CPU、memory、port、environment、health check、log 出力を定義する version 付き設定。 |
| ALB | Application Load Balancer。client request を healthy な ECS task へ転送する入口。 |
| Target Group | ALB が request を転送する対象と health check を管理する単位。この構成では ECS task を登録する。 |
| RDS | Amazon Relational Database Service。このリポジトリでは private subnet の PostgreSQL として使う。 |
| VPC | AWS 上に作る論理的な network 境界。public/private subnet、route、security group を含む。 |
| Public Subnet | Internet Gateway への route を持つ subnet。ALB を配置し、dev 構成では条件付きで ECS task も配置する。 |
| Private Subnet | Internet Gateway への直接 route を持たない subnet。RDS を配置し、internet へ公開しない。 |
| Security Group | AWS resource 単位の stateful firewall。ALB、ECS、RDS 間で必要な通信元と port だけを許可する。 |
| IAM Role | AWS resource や workflow が一時 credential で権限を得る仕組み。ECS task execution role、 task role、GitHub OIDC role の責務を分ける。 |
| Secrets Manager | password などの secret を保存する AWS service。DB credential を source code や tfvars に置かず、 ECS task へ注入するために使う。 |
| CloudWatch | AWS の logs、metrics、alarms、dashboards を扱う service。AWS 環境の初期 observability に使う。 |

## Terraform と CI/CD

| 用語 | 説明 |
| --- | --- |
| IaC | Infrastructure as Code。network、IAM、ECS、RDS などの infrastructure を code と review 可能な差分で管理する考え方。 |
| Terraform State | Terraform configuration と実 resource の対応を記録する情報。secret を含む可能性があるため安全に管理する。 |
| Plan | Terraform が予定する作成、変更、削除を事前表示する操作。plan は実 resource を変更しない。 |
| Apply | Terraform plan に基づいて実 resource を変更する操作。このリポジトリでは明示的な承認なしに実行しない。 |
| Drift | Terraform configuration/state と実 resource の状態が一致しなくなること。console の手動変更などで発生する。 |
| CI | Continuous Integration。test、lint、vulnerability check、build などを変更ごとに自動検証すること。 |
| CD | Continuous DeliveryまたはDeployment。検証済みの成果物を環境へ反映する仕組み。 このリポジトリの AWS deploy は manual gate で保護する。 |
| OIDC | OpenID Connect。GitHub Actions が長期 AWS access key を保存せず、条件付きで AWS role を引き受けるために使う。 |
| Rollback | 問題のある変更を直前の安定状態へ戻すこと。application、infrastructure、DB migration を分けて判断する。 |

## Observability

| 用語 | 説明 |
| --- | --- |
| Observability | 外部から得られる signals を使い、system 内部の状態や障害原因を調べられる性質。 この Lab では logs、metrics、traces を使用する。 |
| Structured Logging | 検索・集計できるよう、決められた field を持つ JSON として log を出す方式。 password、credential、connection string は記録しない。 |
| Metric | 時間とともに変化する数値。request 数、5xx 数、latency histogram、DB connections などを指す。 |
| Counter | 原則として増加だけする metric。`todo_api_http_requests_total` で request 完了数を記録する。 |
| Histogram | 観測値を複数の bucket に集計する metric。request latency の p95 などを計算するために使う。 |
| Label | Prometheus metric の系列を分類する属性。`method`、route template、`status` を使い、 TODO ID、query、request ID は使わない。 |
| Cardinality | metric label の組み合わせ数。値が無制限に増える high cardinality は memory、storage、query cost を増やす。 |
| Trace | 1 request が system 内を通る処理経路を記録したもの。複数の span から構成される。 |
| Span | trace 内の一つの処理区間。HTTP request、repository call、DB query などの開始、終了、属性を表す。 |
| Trace Context | trace ID や parent span の情報を process や service 間で伝播するための context。 |
| OpenTelemetry | logs、metrics、traces の生成、伝播、export に使う vendor-neutral な標準と SDK 群。 Phase 6 では Go API の HTTP trace に使用する。 |
| OTLP | OpenTelemetry Protocol。telemetry data を Collector などへ送る protocol。 この Lab では OTLP/HTTP を使う。 |
| OpenTelemetry Collector | telemetry を受信、処理、転送する component。API と backend を直接結合しないための中継点として使う。 |
| Prometheus | metric を endpoint から定期取得して保存し、PromQL で照会する監視 system。 API の `/metrics` を scrape する。 |
| Scrape | Prometheus が対象の metrics endpoint を定期的に取得すること。取得可能な target は `up` になる。 |
| PromQL | Prometheus Query Language。request rate、5xx ratio、p95 latency などを計算する query 言語。 |
| Grafana | metrics や traces を dashboard と Explore で可視化する tool。この Lab では Prometheus と Tempo を参照する。 |
| Tempo | Grafana の distributed tracing backend。OpenTelemetry Collector から trace を受け取る。 |
| Dashboard | 複数の panel に主要 signal をまとめ、system 状態を継続的に確認する画面。 |
| Alert | metric が定義した条件を満たしたときに異常を通知する仕組み。症状と利用者影響を優先して設計する。 |

## SRE と障害対応

| 用語 | 説明 |
| --- | --- |
| SRE | Site Reliability Engineering。software engineering の手法で reliability、運用、変更速度のバランスを管理する考え方。 |
| SLI | Service Level Indicator。service 品質を測る指標。このリポジトリでは availability、latency、 5xx error rate、DB readiness を扱う。 |
| SLO | Service Level Objective。一定期間に達成したい SLI の目標。初期 availability SLO は30日で99.0%。 |
| SLA | Service Level Agreement。顧客との契約上の service level と違反時の扱い。 内部の改善目標である SLO とは区別する。 |
| Error Budget | SLO が許容する失敗量。30日99.0% availability なら停止時間換算で約7時間12分。 |
| Burn Rate | error budget を想定より何倍速く消費しているかを示す値。短期と長期の window を組み合わせて alert に使う。 |
| Availability | 利用者が期待する request を正常に処理できる割合。このリポジトリでは主に5xxを失敗として計算する。 |
| Latency | request を受けてから response を返すまでの時間。average だけでなく p95 や p99 を確認する。 |
| p95 | 観測値の95%がその値以下に収まる percentile。遅い一部の request を average より把握しやすい。 |
| Liveness | process が生存しているかの確認。`/healthz` は DB に問い合わせず、API process の状態を返す。 |
| Readiness | request を処理できる準備ができているかの確認。`/readyz` は DB connection を確認し、失敗時は503を返す。 |
| Fault Injection | 制御された方法で意図的に障害を発生させ、検知と復旧を検証すること。 Phase 6 では5xx、最大5秒の遅延、PostgreSQL停止を扱う。 |
| Load Test | 一定の request を発生させ、性能、安定性、SLO threshold を確認する試験。この Lab では k6 を使う。 |
| k6 | Grafana Labs が開発する load test tool。JavaScript で scenario、Virtual User（VU）、実行時間、合格条件となる threshold を定義する。このリポジトリでは5 VUで30秒間 TODO 一覧 API を呼び出し、request failure rate が1%未満、request latency の p95 が500ms未満であることを `make loadtest` で検証する。 |
| VU | Virtual User。k6 で同時に処理を繰り返す仮想利用者。この Lab の `vus: 5` は、5つの仮想利用者が並行して request を送る設定を表す。VU数と1秒あたりのrequest数は同じとは限らない。 |
| Threshold | k6 test の合否を決める条件。例えば `http_req_failed: rate<0.01` は失敗率1%未満、`http_req_duration: p(95)<500` はrequest latencyの95%が500ms未満であることを要求する。 |
| Runbook | 障害時の検知、一次確認、切り分け、緩和、復旧確認、事後対応をまとめた手順書。 |
| Incident | service の品質や安全性に影響する事象。学習用 Lab では開始、検知、緩和、復旧の時刻を記録する。 |
| Mitigation | 根本修正の前に利用者影響を小さくする対応。rollback、障害 component の再起動、負荷源停止などを指す。 |
| Postmortem | 障害の timeline、影響、原因、対応、再発防止を非難ではなく改善目的で記録する振り返り。 |
| Root Cause | 障害を発生させた根本的な技術・process 上の原因。直前に観測された症状とは区別する。 |
| Corrective Action | 再発可能性または影響を減らす改善項目。owner、期限、status を明確にする。 |
| MTTD | Mean Time to Detect。障害発生から検知までにかかった時間。 |
| MTTR | Mean Time to Restore。障害発生から service を復旧するまでにかかった時間。 |

## Security と運用

| 用語 | 説明 |
| --- | --- |
| Least Privilege | 処理に必要な最小限の権限だけを付与する原則。IAM role、security group、CI/CD token に適用する。 |
| Secret | password、token、private key など、漏えいさせてはならない情報。source code、log、image、 real `terraform.tfvars` に含めない。 |
| Vulnerability Check | 使用する dependency や Go code に既知の脆弱性がないか確認する検査。このリポジトリでは `make vuln` を使う。 |
| Backup | 障害や誤操作から data を復元するための複製。backup が存在するだけでなく restore 演習が必要になる。 |
| Restore | backup または snapshot から DB を復元する操作。RDS では通常、新しい DB instance として復元する。 |
| Cost Management | 設計、plan、運用、cleanup の各段階で継続課金 resource と利用量を確認すること。 実 AWS resource の作成前に対象と概算コストを確認する。 |
