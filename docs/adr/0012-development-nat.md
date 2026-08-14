# ADR 0012: 開発環境の NAT 戦略

## ステータス

提案中。

## 決定

後続フェーズで必要性が確認されるまでは、開発環境で常時稼働する NAT Gateway を避ける。
infrastructure を apply する前に、開発環境向けの public-IP ECS task、VPC endpoint、NAT Gateway、作成/削除 workflow を比較する。

## 背景

NAT Gateway には固定の月額コストがある。

開発環境は学習用であり、安く作成・削除できるべきである。

## 結果

- development は production と異なる可能性がある。
- apply 前にコストとセキュリティのトレードオフを文書化しなければならない。
- production では private workload とより強い network control を優先するべきである。
