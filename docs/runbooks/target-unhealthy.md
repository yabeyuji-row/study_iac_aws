# Runbook: Target Unhealthy

## 検知

- CloudWatch alarm `study-aws-dev-alarm-alb-unhealthy-targets` が ALARM になる。
- ALB target group の `UnHealthyHostCount` が 1 以上になる。
- ECS service events に task health check failure が出る。

## 一次確認

- ALB target group の registered targets と health reason を確認する。
- `/healthz` と `/readyz` のどちらが失敗しているか確認する。
- ECS service の desired/running/pending task count を確認する。
- CloudWatch Logs で直近の `api stopped`、DB 接続失敗、panic を確認する。

## 切り分け

- `/healthz` 失敗なら、process 起動、container health check、port mapping、image を疑う。
- `/readyz` 失敗なら、RDS、Secrets Manager、security group、DB password を疑う。
- ALB health check path が Terraform の `health_check_path` と API endpoint に一致しているか確認する。

## 緩和

- 直前 deploy 後に発生した場合は、ECS service を直前 task definition revision に戻す。
- desired count が 0 または task 起動失敗なら、service events の stopped reason に合わせて修正する。
- DB 起因なら、DB connection failure runbook に切り替える。

## 復旧確認

- ALB target group の target が healthy になる。
- `/readyz` が 200 を返す。
- `UnHealthyHostCount` alarm が OK へ戻る。
- 新規 request が 2xx/4xx の想定内 status を返す。

## 事後対応

- health check path、deployment、DB readiness のどこで失敗したかを記録する。
- alarm threshold が早すぎる/遅すぎる場合は `docs/design/observability-slo.md` を更新する。
- 再発防止が code/Terraform/docs のどれかを判断し、issue または roadmap に残す。
