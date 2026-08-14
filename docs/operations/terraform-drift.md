# Terraform Drift Handling

## 方針

Terraform 管理 resource は、原則として Terraform から変更する。
AWS console や CLI で手動変更した場合は drift として扱い、緊急対応後に Terraform configuration または実 resource のどちらを正とするかを決める。

参考:

- Terraform resource drift tutorial: https://developer.hashicorp.com/terraform/tutorials/state/resource-drift
- Terraform plan command: https://developer.hashicorp.com/terraform/cli/commands/plan
- Terraform refresh-only mode: https://developer.hashicorp.com/terraform/tutorials/state/refresh

## 確認手順

remote state を使う実 AWS 環境では、次の順で確認する。
このコマンドは実 infrastructure を変更しない。

```sh
terraform -chdir=infrastructure/app init
terraform -chdir=infrastructure/app plan \
  -refresh-only \
  -detailed-exitcode \
  -var-file=envs/dev.tfvars
```

exit code の見方:

- `0`: drift なし。
- `1`: plan 実行エラー。認証、backend、provider、variable を確認する。
- `2`: drift または state との差分あり。

通常の変更 review では次も実行する。

```sh
terraform -chdir=infrastructure/app plan \
  -input=false \
  -var-file=envs/dev.tfvars
```

## 判断基準

drift を見つけたら、plan の resource address と attribute を記録する。

- 緊急手動変更が正しい: Terraform configuration に同じ設定を反映し、PR で review する。
- Terraform configuration が正しい: `terraform apply` で実 resource を戻す。ただし apply 前に対象 resource、影響、概算コスト、戻し方の承認を得る。
- state だけが不整合: `terraform import`、`moved` block、state 操作の必要性を検討する。state 操作は事前 backup を取る。
- secret 値の差分: plan artifact に平文が出ていないか確認し、secret の rotation や ECS 再起動の必要性を判断する。

## 机上演習

想定 drift:

- ALB security group に想定外の ingress が追加された。
- RDS deletion protection が console で変更された。
- ECS desired count が手動変更された。

確認:

```sh
terraform -chdir=infrastructure/app plan -refresh-only -detailed-exitcode -var-file=envs/dev.tfvars
```

対応方針:

- ALB ingress は security incident として扱い、即時で狭める。その後 Terraform に差分がないことを確認する。
- RDS deletion protection は環境方針と照合し、dev なら無効、本番相当なら有効に揃える。
- ECS desired count は障害緩和のための一時 scale out かを確認し、恒久化するなら `ecs_desired_count` を更新する。

## 記録

drift 対応後は以下を issue、PR、または postmortem に残す。

- 検出日時と検出者。
- 変更された resource address。
- 手動変更の理由。
- Terraform configuration に反映したか、実 resource を戻したか。
- 再発防止策。
