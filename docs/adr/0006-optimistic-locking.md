# ADR 0006: 楽観的ロック

## ステータス

採用。

## 決定

楽観的ロックには integer の `version` column を使う。

## 背景

2 つのクライアントが同じ version から同じ row を更新すると、TODO 更新は競合する可能性がある。
この API では悲観的に row をロックする必要はない。

## 結果

- `PUT` では `version` を必須にする。
- 更新成功時に `version` を増やす。
- `version` 不一致は `409 Conflict` を返す。
