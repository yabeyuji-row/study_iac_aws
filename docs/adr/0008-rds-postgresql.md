# ADR 0008: RDS PostgreSQL

## ステータス

採用。

## 決定

Amazon RDS for PostgreSQL を使う。

## 背景

PostgreSQL は必須の信頼できる情報源である。

RDS を使うことで、マネージドバックアップ、モニタリング、parameter group、subnet group、暗号化、運用学習の機会が加わる。

## 結果

- RDS は private subnet で実行しなければならない。
- instance が存在する間はコストが継続する。
- backup と restore の手順を文書化し、テストする必要がある。
