# TODO API 学習プロジェクト

このリポジトリは、Go、Echo v4、PostgreSQL、Docker、AWS、Terraform、
GitHub Actions、オブザーバビリティ、SRE、セキュリティ、バックアップと
リカバリ、コスト管理を使って、本番環境に近い TODO REST API を段階的に
構築するための学習プロジェクトです。

現在のフェーズ: Phase 6、Observability and Reliability Lab。

## アーキテクチャ

```text
Client
  |
Application Load Balancer
  |
Amazon ECS Fargate
  |
Go REST API
  |
Amazon RDS for PostgreSQL
```

## ローカル起動

PostgreSQL と MiniStack を起動し、マイグレーションを実行してから API を起動します。

```bash
make local-up
```

API はフォアグラウンドで起動します。`Ctrl+C` で停止してから、PostgreSQL とMiniStack を停止します。

```bash
make local-down
```

PostgreSQL だけを起動する場合:

```bash
make postgres-up
```

MiniStack だけを起動する場合:

```bash
make ministack-up
```

API だけを起動する場合:

```bash
make api-up
```

`make api-up` は、PostgreSQL が起動済みで、マイグレーションが適用済みである
ことを前提にしています。

別の方法として、Docker Compose で PostgreSQL と API を起動できます。

```bash
make compose-up
```

マイグレーションを実行します。

```bash
make migrate-up
```

Air のホットリロード付きで API を起動します。

```bash
make dev
```

API はデフォルトで `http://localhost:8080` をリッスンします。マイグレーションを
適用する前に API がすでに起動している場合は、TODO リクエストを送る前に
マイグレーションを適用してください。

Air がローカルにインストールされていない場合:

```bash
make install-air
```

ローカルの TODO UI をブラウザで開きます。

```bash
make open
```

## API

実装済みエンドポイント:

```text
POST   /v1/todos
GET    /v1/todos
GET    /v1/todos/{todo_id}
PUT    /v1/todos/{todo_id}
DELETE /v1/todos/{todo_id}
GET    /healthz
GET    /readyz
GET    /version
```

この API は `/v1/todos` の CRUD エンドポイントと、ヘルスチェック、
準備状態、ビルドバージョン確認のための運用エンドポイントを実装しています。

詳細な API 設計は [API 設計](docs/design/api-design.md) にあります。

## Curl 例

TODO を作成します。

```bash
curl -sS -X POST http://localhost:8080/v1/todos \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Terraformを学ぶ",
    "description": "VPCとECSをTerraformで構築する",
    "status": "pending",
    "due_date": "2026-08-31T23:59:59+09:00"
  }'
```

TODO 一覧を取得します。

```bash
curl -sS 'http://localhost:8080/v1/todos?limit=20&sort=created_at_desc'
```

TODO を取得します。

```bash
curl -sS http://localhost:8080/v1/todos/{todo_id}
```

TODO を更新します。

```bash
curl -sS -X PUT http://localhost:8080/v1/todos/{todo_id} \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Terraformを学ぶ",
    "description": "ECSまで完了させる",
    "status": "in_progress",
    "due_date": "2026-08-31T23:59:59+09:00",
    "version": 1
  }'
```

TODO を削除します。

```bash
curl -i -X DELETE http://localhost:8080/v1/todos/{todo_id}
```

## テストと Lint

```bash
make test
make test-race
make lint
make vuln
make build
```

`TEST_DATABASE_URL` が設定されている場合、リポジトリテストは実際の PostgreSQL を
使用します。

```bash
TEST_DATABASE_URL='postgres://todo:todo_password@localhost:5432/todo_api?sslmode=disable' \
  go test ./internal/todo -run TestPostgresRepositoryCRUD
```

`govulncheck` がローカルで利用できない場合は、先にインストールしてください。

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
```

## マイグレーション

SQL ファイルベースのマイグレーションは明示的なコマンドで実行します。API 起動時に
無条件で実行されることはありません。

```bash
make migrate-up
make migrate-down
```

## Docker

API イメージをビルドします。

```bash
make docker-build
```

ローカルの PostgreSQL と API を起動します。

```bash
make compose-up
```

各サービスを個別に起動します。

```bash
make postgres-up
make ministack-up
make api-up
```

PostgreSQL、MiniStack、マイグレーション、API をまとめて起動します。

```bash
make local-up
```

Compose の API サービスは Air を使用し、Go テンプレート、Go ファイル、HTML、
CSS、JavaScript ファイルが変更されるとリロードします。

ローカルサービスを停止します。

```bash
make compose-down
```

## MiniStack

MiniStack は、学習と一部の統合チェックに使うローカルの AWS 互換エンドポイントを
提供します。

```bash
make ministack-up
make ministack-health
make ministack-down
```

Terraform プロバイダーの override 例と、MiniStack でのチェックと実 AWS での
チェックの境界については、[MiniStack セットアップ](docs/ministack-setup.md) を参照して
ください。

## Terraform

アプリケーション用 Terraform は `infrastructure/app/` 配下にあります。ローカルの
検証フローでは、MiniStack 互換の plan のために `envs/dev.tfvars` とダミーの AWS
認証情報を使用します。

```bash
cp infrastructure/app/envs/dev.tfvars.example infrastructure/app/envs/dev.tfvars
make terraform-init
make terraform-fmt
make terraform-validate
make terraform-tflint
make terraform-security
make terraform-plan
```

[Terraform ファイル概要](docs/terraform-file-overview.md) と
[MiniStack Terraform アプリチェック](docs/ministack-terraform-app.md) も参照してください。

明示的な承認なしに `terraform apply` や `terraform destroy` を実行しないでください。

## AWS デプロイ

GitHub Actions は静的アクセスキーではなく OIDC を使用します。
イメージの push は `main`、`v*` タグ、または手動実行から行えます。
ECS サービス更新は、`deploy = true` を指定した手動実行と GitHub の `dev`
環境承認ゲートで保護されています。

初回に AWS リソースを作成する場合は、[AWS デプロイ手順](docs/aws-deploy.md) に従います。
NAT Gateway を使わない dev 構成では、ECS task を public subnet に配置し、
public IP と HTTPS egress を有効にします。

AWS 運用の入口:

- [RDS バックアップとリストア](docs/operations/rds-backup-restore.md)
- [ロールバック手順](docs/operations/rollback.md)
- [Terraform drift 対応](docs/operations/terraform-drift.md)
- [AWS コストレビュー用チェックリスト](docs/operations/cost-review-checklist.md)
- [AWS クリーンアップ用チェックリスト](docs/operations/cleanup-checklist.md)
- [ポストモーテムテンプレート](docs/templates/postmortem-template.md)

## CI/CD

GitHub Actions のワークフローは `.github/workflows/` 配下にあります。
Pull Request のチェックには、ワークフロー lint、Go テスト、race テスト、lint、
脆弱性チェック、Docker ビルド、Terraform fmt/validate、TFLint、Checkov、
Terraform plan artifact が含まれます。

デプロイワークフローのイメージ push は `main`、`v*` タグ、または手動実行から行えます。
ECS サービス更新は、`deploy = true` を指定した手動実行と GitHub の `dev`
環境承認ゲートで保護されています。

## 監視と SRE

初期のオブザーバビリティでは、JSON リクエストログ、CloudWatch Logs metric filter、
CloudWatch アラーム、CloudWatch ダッシュボードを使用します。
[オブザーバビリティと SLO 設計](docs/design/observability-slo.md) と
[ランブック](docs/runbooks/) を参照してください。

最終的なリカバリ訓練とレビュー記録については、
[Week 12 最終レビュー](docs/operations/week12-final-review.md) を参照してください。

## Observability Lab

Phase 6 では、ローカルだけで API、PostgreSQL、Prometheus、Grafana、Tempo、
OpenTelemetry Collector を起動し、負荷と障害を安全に再現できます。

```bash
make observability-up
make loadtest
make fault-5xx
make fault-delay FAULT_DELAY=2s
make fault-db-down
curl -i http://localhost:8080/readyz
make fault-db-recover
make observability-down
```

Grafana は `http://localhost:3000`、Prometheus は `http://localhost:9090`、API metrics は
`http://localhost:8080/metrics` です。障害注入ルートは Observability Lab の API でのみ
有効になり、さらに `X-Fault-Injection: enabled` が必要です。

設計と演習手順は [Observability and Reliability Lab](docs/design/observability-reliability-lab.md)
を参照してください。

## このリポジトリで説明できるスキル

- Echo v4 と PostgreSQL を使った Go REST API のレイヤー設計、テスト、migration。
- Docker でのローカル再現と、Terraform による ALB、ECS Fargate、RDS、IAM の設計。
- GitHub Actions OIDC、品質ゲート、段階的 deploy と rollback の判断。
- logs、metrics、traces を関連付けた SLI/SLO、error budget、dashboard、alert の設計。
- k6 負荷試験、安全な障害注入、runbook に沿った復旧、postmortem と再発防止。
- backward-compatible migration、backup/PITR、drift、security、AWS cost を含む運用判断。

## コストに関する注意

対象リソースと概算コストを確認する前に、AWS リソースを作成しないでください。
ALB、NAT Gateway、RDS、VPC endpoint、CloudWatch Logs、Secrets Manager、
データ転送では継続的な料金が発生する可能性があります。

## 設計ドキュメント

- [用語集](docs/glossary.md)
- [システム概要](docs/design/system-overview.md)
- [API 設計](docs/design/api-design.md)
- [データベース設計](docs/design/database-design.md)
- [AWS アーキテクチャ](docs/design/aws-architecture.md)
- [学習ロードマップ](docs/learning-roadmap.md)
- [ADRs](docs/adr/)
