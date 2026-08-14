# ADR 0004: UUID 生成

## ステータス

採用。

## 決定

UUID はアプリケーションで生成する。

## 背景

アプリケーション側で UUID を生成すると挿入が明示的になり、ローカル PostgreSQL とRDS 環境の両方で PostgreSQL extension を必須にせずに済む。

## 結果

- テストで既知の ID を使える。
- ID 作成エラーはアプリケーションが扱う。
- データベースは引き続き UUID 型と primary key uniqueness を強制する。
