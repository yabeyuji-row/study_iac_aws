# ADR 0010: REST のみ

## ステータス

採用。

## 決定

REST のみを使う。GraphQL、SSE、WebSocket は使わない。

## 背景

TODO CRUD は REST endpoint に自然に対応する。

リアルタイム更新や graph-shaped query は不要である。

## 結果

- クライアントと infrastructure がシンプルになる。
- ALB health と routing model が扱いやすくなる。
- 将来プロトコルを変更する場合は新しい ADR が必要になる。
