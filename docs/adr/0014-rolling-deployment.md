# ADR 0014: ローリングデプロイ

## ステータス

採用。

## 決定

初期状態では ECS ローリングデプロイを使う。

## 背景

小さな TODO API にはローリングデプロイで十分であり、ALB health check、ECS desired count、deployment circuit breaker と自然に組み合わせられる。

## 結果

- Blue/Green deployment よりシンプルである。
- graceful shutdown と正しい health check が必要である。
- rollback 手順を文書化する必要がある。
