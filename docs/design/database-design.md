# データベース設計

## データベース

PostgreSQL を永続的な信頼できる情報源とする。
SQLite とインメモリ永続化はスコープ外とする。

## テーブル

予定するテーブル: `todos`

カラム:

```text
id           uuid primary key
title        text not null
description  text not null default ''
status       text not null
due_date     timestamptz null
created_at   timestamptz not null
updated_at   timestamptz not null
version      integer not null
```

## 制約

予定するデータベース制約:

- アプリケーションで trim した後の `title` の長さは 1 から 200。
- `description` の長さは最大 2,000。
- `status` は `pending`、`in_progress`、`completed` の検査制約。
- `version >= 1`。
- `updated_at >= created_at`。

アプリケーションでも入力を検証するが、正しさをアプリケーションのバリデーションだけに依存してはならない。

## UUID 生成

UUID はアプリケーションで生成する。
これにより挿入が明示的になり、ローカルPostgreSQL と RDS 環境でデータベース拡張に依存せず、テストも考えやすくなる。
UUID ADR を参照する。

## タイムゾーン方針

すべてのタイムスタンプは UTC で保存する。
`due_date` ではオフセット付きのクライアントタイムスタンプを受け入れ、永続化前に正規化する。

## インデックス

予定するインデックス:

- `id` の主キー。
- カーソルページング用の `(created_at, id)` 複合インデックス。
- status で絞り込むカーソルページング用の `(status, created_at, id)` 複合インデックス。

TODO title の重複は有効なため、主キー以外の一意制約は予定しない。

## カーソルページング

オフセットページングは使わない。
カーソルには最後の項目のソート位置をエンコードする。

```text
created_at + id
```

複数行が同じ `created_at` を持つ場合でも、`id` を同値時の並び替えキーとして使うことで
並び順を安定させる。

## 楽観的ロック

更新時に `version` を確認する。

```text
where id = $1 and version = $2
```

更新が成功した場合は `version` を 1 増やす。
更新された行がなく、かつ行自体は存在する場合、API は `409 Conflict` を返す。

## トランザクション

作成、更新、削除操作では、必要に応じて明示的なトランザクション境界を持たせる。
初期実装では単一ステートメントの操作をシンプルに保ってよいが、更新競合の検出は明示的に行い、テストする必要がある。

## タイムアウトとプーリング

リポジトリメソッドは `context.Context` を受け取り、サービスまたはリクエスト経路からDB タイムアウトを適用する。
コネクションプール設定は設定値で制御する。

初期の RDS 接続プールは、学習用 dev 環境の小さい instance class に合わせて控えめにする。
API プロセスあたりの接続数は `MaxConns = 10`、常時維持する接続数は `MinConns = 1` とし、
接続 lifetime は 30 分、idle timeout は 5 分にする。
RDS instance class や ECS task 数を増やす場合は、総接続数が PostgreSQL の許容量を超えないように、
`ECS desired count * MaxConns` を確認してから調整する。

## RDS とバックアップ

9週目の Terraform では、TODO API の永続データ用に RDS PostgreSQL を定義する。
RDS は private subnet の DB subnet group にのみ配置し、`publicly_accessible = false` とする。
DB security group は ECS task security group からの PostgreSQL port のみを受ける。

開発環境の初期値:

- Engine: PostgreSQL 16 系。
- Instance class: `db.t4g.micro`。
- Storage: gp3, 20 GiB 開始, 50 GiB まで autoscaling。
- Storage encryption: enabled。
- Backup retention: 7 days。
- Final snapshot: destroy 時に取得する。

RDS の性能、Multi-AZ、Performance Insights、より長い backup retention は、実運用想定やコスト確認後に調整する。
AWS apply が必要な場合は、RDS と Secrets Manager の概算コストを先に報告し、承認を得る。
snapshot restore、PITR、deletion protection、maintenance window の運用方針は
[RDS backup and restore](../operations/rds-backup-restore.md) にまとめる。

## DB 認証情報

DB password は `.env`、`tfvars`、Terraform ファイルへ平文で書かない。
Terraform は `random_password` で password を生成し、RDS の master password と
Secrets Manager の secret version に渡す。
plan では `password` と `secret_string` が sensitive value として扱われることを確認する。

Go API はローカルでは `DATABASE_URL` を読む。
ECS では Secrets Manager secret を `DATABASE_SECRET_JSON` 環境変数として注入し、API 起動時に
`postgres://...` の接続文字列を組み立てる。
これによりアプリケーションログや Docker image に DB 接続文字列を含めない。

## SQL インジェクション

SQL ではパラメータ化クエリを使わなければならない。
ユーザー値に対して文字列結合を使ってはならない。

## マイグレーション

マイグレーションは SQL ファイルベースとし、明示的なコマンドで実行する。
API は起動時に無条件のマイグレーションを実行してはならない。
