# ADR 0013: OpenTelemetry 導入タイミング

## ステータス

採用。

## 決定

初期実装では OpenTelemetry を後回しにする。

## 背景

初期フェーズでは、安定した CRUD 動作、ログ、readiness、テスト、Docker、Terraform が必要である。

完全な tracing は有用だが、基礎が整う前に dependency と運用概念を追加することになる。

## 結果

- 初期の observability では JSON log、CloudWatch metrics、alarm を使う。
- フェーズ 5 の runbook で具体的な tracing need が見つかった後に、
  OpenTelemetry を再検討できる。
