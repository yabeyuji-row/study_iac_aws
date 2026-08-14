# ADR 0009: Redis を使わない

## ステータス

採用。

## 決定

Redis または ElastiCache は使わない。

## 背景

TODO API には caching、queue、pub/sub、distributed lock は不要である。
Redis を追加すると、現時点での必要性がないまま運用作業が増える。

## 結果

- PostgreSQL を唯一の業務データの信頼できる情報源として維持する。
- 課金対象の AWS リソースが少なくなる。
- キャッシュ無効化を避けられる。
