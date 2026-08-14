# ADR 0007: ECS Fargate

## ステータス

採用。

## 決定

アプリケーションランタイムには Amazon ECS Fargate を使う。

## 背景

Fargate は EC2 instance を管理せずにコンテナオーケストレーションを提供する。

Kubernetes を導入せずに、ALB、task definition、IAM role、CloudWatch logs、ローリングデプロイを学ぶ用途に適している。

## 結果

- EC2 host 管理が不要になる。
- コストは実行中の task に紐づく。
- ECS 固有のデプロイとヘルスチェックの挙動を理解する必要がある。
