# ADR 0011: Terraform ディレクトリ構成

## ステータス

フェーズ 3 で採用。

## 決定

Terraform を導入するときは、以下の構成を使う。

```text
infrastructure/
  bootstrap/
  modules/
    network/
    ecs-service/
    database/
    observability/
  environments/
    development/
    production/
```

## 背景

bootstrap state と application state は lifecycle を分ける必要がある。

development とproduction で再利用する価値がある箇所に module を置く。

## 結果

- フェーズ 3 を明確な state 境界から始められる。
- resource が存在する前の過剰な module 化を避けられる。
- environment 差分が見える状態で残る。
