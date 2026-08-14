# システム概要

## 目的

このプロジェクトでは、Go、PostgreSQL、Docker、AWS、Terraform、GitHub Actions、オブザーバビリティ、SRE、セキュリティ、バックアップと復旧、AWS コスト管理を実践的に学ぶための題材として、TODO 管理 REST API を構築する。

API 自体は意図的にシンプルにする。
学習上の価値は、本番に近い制約のもとで構築し運用するところにある。

## スコープ

このシステムは、以下のフィールドを持つ TODO レコードを管理する。

- `id`
- `title`
- `description`
- `status`
- `due_date`
- `created_at`
- `updated_at`
- `version`

TODO 状態の信頼できる情報源は PostgreSQL とする。
アプリケーションプロセスは、永続的な業務状態をメモリ内に保持してはならない。

## 目標アーキテクチャ

```text
クライアント
  |
Application Load Balancer
  |
Amazon ECS Fargate
  |
Go REST API
  |
Amazon RDS for PostgreSQL
```

## 初期技術決定

2026-08-03 時点の公式情報では、現在の安定版は以下のとおり。

- Go: 1.26.5
- Terraform: 1.15.8
- Terraform AWS Provider: 6.57.1

各ツールを最初に導入するフェーズで、実装上の正確なバージョン固定を追加する。

## アプリケーションレイヤー

アプリケーションは小さなレイヤー構造を採用する。

```text
HTTP ハンドラー -> サービス -> リポジトリ -> PostgreSQL
```

責務:

- ハンドラー: HTTP の解析、レスポンス書き込み、リクエスト ID、エラーマッピング。
- サービス: バリデーション、ビジネスルール、トランザクションの意図。
- リポジトリ: PostgreSQL アクセスと SQL の詳細。
- PostgreSQL: 永続状態とデータベース制約。

## スコープ外

初期システムには認証やユーザー管理を含めない。

これは学習のための意図的なトレードオフであり、インターネットへ公開する前にセキュリティリスクとして文書化しなければならない。

このプロジェクトでは、Redis、ElastiCache、SSE、WebSocket、GraphQL、Kafka、Kinesis、EventBridge、SNS、SQS、DynamoDB、Lambda、Kubernetes、EKS、CQRS、Event Sourcing、不要なキャッシュは使わない。

## フェーズ計画

- フェーズ 0: 設計文書と ADR のみ。
- フェーズ 1: PostgreSQL と Docker Compose を使ったローカル Go TODO API。
- フェーズ 2: ローカル API の運用品質。
- フェーズ 3: AWS インフラ用 Terraform。
- フェーズ 4: GitHub Actions CI/CD。
- フェーズ 5: SRE 文書、runbook、復旧、演習。

## フェーズ 0 の境界

フェーズ 0 では、設計文書、ADR、AGENTS.md、README の雛形を作成する。
Go コード、Dockerfile、Terraform コード、GitHub Actions、AWS リソース、デプロイは作成しない。
