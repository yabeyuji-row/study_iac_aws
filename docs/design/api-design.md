# API 設計

## 基本原則

- HTTP 上の REST。
- リクエスト本文とレスポンス本文は JSON。
- 一貫したエラーエンベロープ。
- すべてのレスポンスとログエントリにリクエスト ID を付与する。
- 初期スコープでは認証を扱わない。

## エンドポイント

```text
POST   /v1/todos
GET    /v1/todos
GET    /v1/todos/{todo_id}
PUT    /v1/todos/{todo_id}
DELETE /v1/todos/{todo_id}
GET    /health
GET    /healthz
GET    /ready
GET    /readyz
GET    /version
```

## TODO フィールド

```text
id           UUID
title        必須、trim 済み、1..200 文字
description  任意、最大 2,000 文字
status       pending | in_progress | completed
due_date     リクエストでは任意のタイムゾーン付きタイムスタンプ
created_at   UTC タイムスタンプ
updated_at   UTC タイムスタンプ
version      integer の楽観的ロック用 version
```

`due_date`、`created_at`、`updated_at` は UTC で保存する。クライアントは
オフセット付きタイムスタンプを送信でき、アプリケーションは保存前に UTC へ正規化する。

## TODO 作成

```text
POST /v1/todos
```

リクエスト:

```json
{
  "title": "Terraformを学ぶ",
  "description": "VPCとECSをTerraformで構築する",
  "status": "pending",
  "due_date": "2026-08-31T23:59:59+09:00"
}
```

レスポンス: `201 Created`

## TODO 一覧取得

```text
GET /v1/todos?status=pending&limit=20&cursor=...&sort=created_at_desc
```

対応するクエリパラメータ:

- `status`
- `limit`
- `cursor`
- `sort`

ルール:

- デフォルトの `limit`: 20。
- 最大 `limit`: 100。
- オフセットページングではなくカーソルページングを使う。
- 主な並び順: `created_at`。
- 不正なクエリパラメータは `400 Bad Request` を返す。

最初の実装では `created_at_desc` と `created_at_asc` をサポートする。
複数レコードが同じタイムスタンプを持つ場合でも安定したページングを提供するため、
カーソルには最後に見た `created_at` と `id` をエンコードする。

## TODO 取得

```text
GET /v1/todos/{todo_id}
```

TODO が存在しない場合は `404 Not Found` を返す。

## TODO 更新

```text
PUT /v1/todos/{todo_id}
```

リクエスト:

```json
{
  "title": "Terraformを学ぶ",
  "description": "ECSまで完了させる",
  "status": "in_progress",
  "due_date": "2026-08-31T23:59:59+09:00",
  "version": 1
}
```

`version` は必須。現在のデータベース上の `version` と異なる場合は
`409 Conflict` を返す。

## TODO 削除

```text
DELETE /v1/todos/{todo_id}
```

物理削除を行う。成功時は `204 No Content`、TODO が存在しない場合は
`404 Not Found` を返す。

## ヘルスエンドポイント

- `/health` と `/healthz`: プロセスの生存確認のみ。DB クエリは実行しない。ECS container health check では `/healthz` を使う。
- `/ready` と `/readyz`: タイムアウト付きで DB 接続性を確認する。失敗時は `503` を返す。ALB target group health check では `/readyz` を使う。
- `/version`: `version`、`commit`、`build_time` を含むビルドメタデータ。

## エラー形式

```json
{
  "error": {
    "code": "TODO_NOT_FOUND",
    "message": "todo が見つかりません",
    "request_id": "..."
  }
}
```

エラーカテゴリ:

- バリデーションエラー: `400`
- 不正な JSON: `400`
- リソース未検出: `404`
- `version` 競合: `409`
- DB readiness 失敗: `503`
- 内部エラー: `500`

内部エラーの詳細をクライアントへ返してはならない。
