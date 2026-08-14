# ADR 0002: PostgreSQL Driver

## ステータス

初期実装で採用。

## 決定

必要に応じて、pgx の標準ライブラリ互換 interface または native pool を使う。

## 背景

このプロジェクトには PostgreSQL 対応、context の伝播、コネクションプール、適切なエラー処理が必要である。
小さな TODO schema には重い ORM は不要である。

## 結果

- PostgreSQL の挙動が明示的になる。
- 学習のために SQL が見える状態で残る。
- リポジトリテストは mock だけでなく実際の PostgreSQL を使う必要がある。
