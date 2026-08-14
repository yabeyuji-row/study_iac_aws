# AWS アーキテクチャ

## 目標ランタイム

```text
クライアント
  |
Application Load Balancer
  |
Go API を実行する ECS Fargate タスク
  |
プライベートサブネット内の RDS PostgreSQL
```

## マネージドサービス

- コンテナイメージには ECR を使う。
- アプリケーションランタイムには ECS Fargate を使う。
- HTTP ルーティングとヘルスチェックには ALB を使う。
- 永続データには RDS PostgreSQL を使う。
- データベース認証情報には Secrets Manager を使う。
- オブザーバビリティには CloudWatch Logs と Alarms を使う。
- Terraform リモート state には S3 を使う。

## 8-1 時点の ALB/ECS/Logs 構成

8-1 の Terraform では、Go API を ECS Fargate で動かすための外形を定義する。

```text
Client
  |
  | HTTP :80
  v
public subnet の ALB
  |
  | target group health check: /readyz
  | traffic: app_port
  v
private subnet の ECS Fargate task
  |
  | stdout/stderr
  v
CloudWatch Logs log group
```

ECR repository は API image の保存先である。
同じ tag の上書きを避けるため tag immutability を有効にし、tag が外れた image は lifecycle policy で一定日数後に削除する。

ALB は public subnet に配置し、internet からの HTTP を受ける。
ALB security group は internet から 80/tcp を受け、ECS security group の `app_port` へだけ送信する。
ECS security group は ALB security group からの `app_port` だけを受ける。
これにより、ECS task へ internet から直接到達する経路は作らない。

ALB target group の health check は `/readyz` を使う。
`/readyz` は DB 接続性を含めて確認する readiness check なので、DB に接続できない task へ ALB が traffic を流さないために使う。
ECS container health check は `/healthz` を使う。
`/healthz` はプロセスの生存確認だけを行う liveness check なので、コンテナ自体が応答できるかを確認するために使う。

CloudWatch Logs log group は ECS task の `awslogs` log driver から参照する。
ログ保持期間は `log_retention_days` で管理し、開発環境では短く保つ。

### 11-1 時点の Observability 構成

11-1 の Terraform では、CloudWatch Logs metric filter、CloudWatch alarm、CloudWatch dashboard を追加する。
Go API は request 完了時に JSON structured log を出し、CloudWatch Logs metric filter が
request count、error count、latency、DB readiness failure の custom metric を作る。

ALB では target 5xx と unhealthy target を見る。
ECS では Container Insights を有効化し、`RunningTaskCount` で service の実行 task 数を見る。
RDS では CPU、free storage、database connections を見る。

詳細な SLI/SLO、alarm threshold、dashboard、runbook は `docs/design/observability-slo.md` と
`docs/runbooks/` に記録する。

8-1 時点では RDS と Secrets Manager をまだ作成しないため、ECS task には本番用の `DATABASE_URL` を渡さない。
DB 接続情報は 9週目の RDS/Secrets Manager タスクで、secret を Terraform ファイルへ直書きしない形で追加する。
また private subnet には NAT Gateway がないため、実 AWS で ECS task を起動して ECR pull や CloudWatch Logs 送信を行うには、後続で NAT Gateway、VPC endpoint、または開発環境限定の public IP 方式を選ぶ必要がある。

## ネットワーク

予定するネットワーク:

- 1 つの VPC。
- ALB 用の 2 つのパブリックサブネット。
- 必要に応じて ECS と RDS 用の 2 つのプライベートサブネット。
- パブリックルーティング用の Internet Gateway。
- 必要最小限の ingress を持つ security group。

RDS はパブリックにアクセス可能であってはならない。
データベースへの ingress はアプリケーション security group からのみ許可する。

### 9-1 時点の RDS/Secrets 構成

9-1 の Terraform では、RDS PostgreSQL と Secrets Manager secret を追加する。

```text
ECS Fargate task
  |
  | PostgreSQL :5432
  v
private subnet の RDS PostgreSQL

ECS task definition
  |
  | DATABASE_SECRET_JSON
  v
Secrets Manager secret
```

RDS は private subnet の DB subnet group にのみ配置する。
`publicly_accessible` は `false` とし、security group ingress は ECS task security group からの
`db_port` のみに絞る。
RDS storage encryption と automated backup retention を有効にし、destroy 時は final snapshot を取得する。

DB password は `random_password` で生成し、RDS instance と Secrets Manager secret version に渡す。
Terraform plan では RDS `password` と secret version `secret_string` が sensitive として隠れる。
ECS は Secrets Manager secret を `DATABASE_SECRET_JSON` としてコンテナへ注入する。
この注入は ECS agent が行うため、task execution role に対象 secret の read 権限を付ける。
アプリケーションが将来 AWS SDK で secret を直接読む場合に備え、task role にも同じ secret ARN に限定した read 権限を付ける。

MiniStack では Secrets Manager/RDS API の作成形だけを確認対象にする。
RDS の性能、PITR、private subnet 内到達性、TLS は AWS 実環境でのみ本番相当の確認対象とする。

### 7-1 時点のサブネット設計

7-1 のローカル Terraform では、プライベートサブネットに NAT Gateway を作成しない。
パブリックサブネットだけを public route table に関連付け、`0.0.0.0/0` は Internet Gateway へ向ける。
プライベートサブネットには Internet Gateway への default route を持たせない。

ALB はパブリックサブネットに配置する想定にする。
ECS Fargate タスクは ALB からの通信だけを受けるアプリケーション層として、プライベートサブネットに配置する方針にする。
RDS PostgreSQL はデータ層として、プライベートサブネットにのみ配置し、publicly accessible は false にする。

この時点では、ECS タスクが ECR、CloudWatch Logs、Secrets Manager などへ外向き通信する経路はまだ確定しない。
後続フェーズで、開発環境のみ public IP を付ける方式、VPC endpoint、NAT Gateway、必要時だけ作成して削除する運用を比較して決める。
そのため 7-1 では、ネットワーク境界と配置意図を先に固定し、常時課金される NAT Gateway は追加しない。

## 開発環境

開発環境は学習とコスト管理を優先して最適化する。

- ECS タスクは 1 つでよい。
- Single-AZ RDS でよい。
- ログ保持期間は短くする。
- バックアップ保持期間は短くする。
- 削除保護は無効にする。
- 環境は必要な時だけ作成できるようにする。

NAT Gateway には固定の月額コストがある。
開発環境では以下を比較する。

- NAT Gateway.
- 開発環境でのみ ECS タスクに public IP を付与する方式。
- VPC endpoint。
- 必要な時に環境を作成し、不要になったら削除する workflow。

開発環境の初期方針として、後続フェーズで必要性が確認されるまでは、常時稼働する NAT Gateway を避ける。

## 本番設計

実際にデプロイしない場合でも、設計の完全性のために本番用 Terraform を書いてよい。

- 少なくとも 2 つの ECS タスク。
- 複数 Availability Zone。
- RDS Multi-AZ を検討する。
- 削除保護を有効にする。
- バックアップとログの保持期間を長くする。
- より強いアラームとロールバック手順。

## デプロイ

初期デプロイ戦略は、以下を組み合わせた ECS ローリングデプロイとする。

- ALB ヘルスチェック。
- ECS デプロイメントサーキットブレーカー。
- 環境ごとに調整した最小正常率と最大率。
- Go プロセスでのグレースフルシャットダウン。

Blue/Green deployment は意図的に後回しにする。
最初の学習版には不要な運用上の複雑さが増えるためである。

## Terraform State

bootstrap state とアプリケーションインフラ state は分離する。

```text
infrastructure/
  bootstrap/
  modules/
  environments/
    development/
    production/
```

bootstrap は S3 state bucket と関連する IAM policy を作成する。
アプリケーション環境は、bootstrap を明示的に apply した後に remote state を使う。

## セキュリティ

- GitHub Actions は静的な AWS access key ではなく OIDC を使う。
- DB 認証情報は Secrets Manager に保存する。
- コンテナは non-root で実行する。
- CloudWatch logs に secret を含めてはならない。
- security group は、public ALB の HTTP または HTTPS を除き、広い ingress を避ける。

## IAM

ECS の IAM role は、実行基盤用とアプリケーション用で分ける。

- ECS task execution role は ECS agent が使う。ECR からの image pull と CloudWatch Logs への書き込みに必要な権限だけを持つ。
- ECS task role はアプリケーションコンテナが使う。9-1 では DB secret ARN に絞った読み取り権限だけを持つ。
- ECS task execution role は、ECS agent が secret を環境変数として注入するため、同じ DB secret ARN の読み取り権限を持つ。

IAM policy では `Action = "*"` や `Resource = "*"` を安易に使わない。
AWS 管理ポリシーを使う場合も、役割と用途が明確なものだけを attachment する。

## コストメモ

フェーズ 0 では AWS リソースを作成しない。

今後のフェーズでは、課金対象リソースを作成する前に月額見積もりを報告しなければならない。
固定費が発生するサービスとして、ALB、NAT Gateway、RDS、VPC endpoint、CloudWatch Logs、Secrets Manager、backup storage を特に注意して確認する。

## 12-1 時点の復旧とコスト運用

復旧、drift、cost、cleanup は手順を分けて管理する。

- RDS snapshot restore と PITR は [RDS backup and restore](../operations/rds-backup-restore.md) に従う。
- ECS、Terraform、DB migration の rollback は [Rollback procedures](../operations/rollback.md) に従う。
- Terraform drift は [Terraform drift handling](../operations/terraform-drift.md) に従い、`plan -refresh-only` で確認する。
- AWS 課金観点は [AWS cost review checklist](../operations/cost-review-checklist.md) で ALB、ECS、RDS、NAT、CloudWatch、Secrets Manager、ECR を確認する。
- 学習環境の削除後は [AWS cleanup checklist](../operations/cleanup-checklist.md) で残課金 resource を確認する。

RDS restore、PITR、destroy は実 AWS resource を作成または削除する可能性があるため、実行前に対象 resource、想定停止時間、概算コスト、戻し方を報告して承認を得る。
