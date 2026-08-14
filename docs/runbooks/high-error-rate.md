# Runbook: High Error Rate

## 検知

- CloudWatch alarm `study-aws-dev-alarm-alb-target-5xx` が ALARM になる。
- App log derived metric `ErrorCount` が増える。
- PR または deploy 直後に user-facing 5xx が増える。

## 一次確認

- ALB `HTTPCode_Target_5XX_Count` と `RequestCount` を同じ時間帯で見る。
- CloudWatch Logs で `status >= 500` の `http_request_completed` event を確認する。
- 直近 deploy、migration、Terraform change の有無を確認する。
- RDS CPU、connections、free storage を確認する。

## 切り分け

- 特定 path だけなら handler、validation、repository error を疑う。
- 全 path で失敗なら config、DB、Secrets Manager、network、deploy を疑う。
- `/readyz` も失敗するなら DB connection failure runbook に切り替える。
- deploy 直後なら新旧 task definition image tag と build metadata を確認する。

## 緩和

- 直前 deploy が原因なら ECS service を直前 task definition revision へ rollback する。
- DB 高負荷なら traffic を止める前に読み取り/書き込み状況を確認し、不要な負荷源を止める。
- migration 起因なら migration down が安全か確認し、危険なら snapshot restore 手順へ進む。

## 復旧確認

- ALB target 5xx alarm が OK へ戻る。
- `ErrorCount` が通常値に戻る。
- `/readyz` が 200 を返す。
- CloudWatch Logs に同種 error が継続していない。

## 事後対応

- 失敗 path、status、原因 commit、rollback 有無を記録する。
- test、migration review、deploy gate のどこで防げたかを確認する。
- SLO error budget への影響を見積もる。
