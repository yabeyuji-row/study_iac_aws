# 9週目 RDS と Secrets Manager レビュー

## 実装内容

9週目では、RDS PostgreSQL と Secrets Manager を Terraform に追加した。

- RDS subnet group は private subnet のみを参照する。
- RDS PostgreSQL は `publicly_accessible = false` とする。
- RDS security group は ECS security group からの PostgreSQL port のみを許可する。
- DB password は `random_password` で生成し、`.env` や `tfvars` には書かない。
- Secrets Manager secret version の `secret_string` に DB 接続 JSON を保存する。
- ECS task definition は `DATABASE_SECRET_JSON` として DB secret を注入する。
- Go API は `DATABASE_URL` または `DATABASE_SECRET_JSON` から DB 接続設定を読む。

## Backup 方針

開発環境の初期値では、RDS automated backup retention を 7 日にする。
storage は暗号化し、snapshot にも tag をコピーする。
Terraform destroy 時は final snapshot を取得する設定にし、RDS 削除の前に復旧点を残す。

本番相当の運用では、backup retention、Multi-AZ、Performance Insights、deletion protection を
コストと復旧要件に合わせて見直す。

## Secret と plan の確認

`make terraform-plan` で以下を確認した。

- `aws_db_instance.app.password` は `(sensitive value)` と表示される。
- `aws_secretsmanager_secret_version.db.secret_string` は `(sensitive value)` と表示される。
- `random_password.db.result` は `(sensitive value)` と表示される。
- DB 接続文字列や DB password は plan に表示されない。

表示される値は DB 名、username、port、instance class、storage などの非 secret 設定に限る。

## Readiness

`/ready` と `/readyz` は DB `Ping` を行う。
DB 接続に失敗した場合は `503 Service Unavailable` と `{"status":"unready"}` を返す。
ALB target group health check は `/readyz` を使うため、DB に接続できない task へ traffic を流さない。

## MiniStack

MiniStack では Secrets Manager secret と RDS API の作成形だけを確認対象にする。
このリポジトリのルールにより、Terraform apply は明示承認なしに実行しない。
RDS の性能、backup/PITR、private subnet 内到達性、TLS は MiniStack では本番相当確認としない。

## 障害演習

### Secret ARN 誤設定

想定: ECS task definition の `DATABASE_SECRET_JSON` が存在しない secret ARN を参照する。

影響:

- ECS agent が secret を注入できず、task 起動に失敗する。
- アプリケーションが起動しても `DATABASE_SECRET_JSON` が欠落すると config validation で失敗する。
- ALB から見ると healthy target が増えない。

検出:

- ECS service events。
- ECS task stopped reason。
- CloudWatch Logs にアプリ起動前のログが出ない可能性。

戻し方:

- task definition の secret ARN を Terraform 管理の `aws_secretsmanager_secret.db.arn` に戻す。
- Terraform plan で task definition replacement のみであることを確認する。

### RDS security group ingress 削除

想定: `aws_vpc_security_group_ingress_rule.rds_from_ecs` を削除する。

影響:

- ECS task から RDS への TCP 接続が通らない。
- API の `/readyz` は DB Ping 失敗により 503 を返す。
- ALB target group は task を unhealthy と判定する。

検出:

- `/readyz` の 503。
- ALB target unhealthy。
- アプリログの DB 接続失敗。

戻し方:

- ECS security group から RDS security group への `db_port` ingress を復元する。
- Go test と Terraform validate を再実行する。

### DB password 欠落

想定: `DATABASE_SECRET_JSON` から `password` を欠落させる。

影響:

- Go API の config validation が失敗し、起動しない。
- DB へ誤った接続を試行する前に停止できる。

検出:

- `Load()` が `password is required` を含むエラーを返す。
- unit test `TestLoadRejectsDatabaseSecretWithoutPassword` で確認する。

戻し方:

- Secrets Manager secret version の JSON に `password` を含める。
- Terraform 管理の secret version に戻す。

## 確認コマンド

```sh
go test ./...
terraform -chdir=infrastructure/app fmt -recursive
terraform -chdir=infrastructure/app validate
terraform -chdir=infrastructure/app plan -input=false -refresh=false -var-file=envs/dev.tfvars
```
