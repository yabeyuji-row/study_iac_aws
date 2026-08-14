# MiniStack セットアップ

MiniStack はローカルで AWS 互換 API を起動し、Terraform、AWS CLI、SDK の一部動作を検証するために使う。
このプロジェクトでは AWS 実リソースの代替ではなく、学習とローカル統合確認の補助として扱う。

## 起動

```bash
make ministack-up
```

ヘルスチェック:

```bash
make ministack-health
```

ログ確認:

```bash
make ministack-logs
```

停止:

```bash
make ministack-down
```

MiniStack は `http://localhost:4566` で待ち受ける。AWS CLI から使う場合は `--endpoint-url`、region、ローカル用のダミー認証情報を指定する。

```bash
AWS_ACCESS_KEY_ID=test \
AWS_SECRET_ACCESS_KEY=test \
aws --endpoint-url=http://localhost:4566 --region ap-northeast-1 s3 ls
```

この repository の Makefile では `MINISTACK_ENDPOINT` と `AWS_REGION` を既定値として持っているため、Secrets Manager の確認は次で実行できる。

```bash
make ministack-list-secrets
```

## Terraform Provider 例

MiniStack 用の provider override は AWS 実環境用の provider と分ける。
本番用の `common-providers.tf` に直接混ぜず、local 専用ファイルや docs の例として扱う。

```hcl
provider "aws" {
  region                      = "ap-northeast-1"
  access_key                  = "test"
  secret_key                  = "test"
  s3_use_path_style           = true
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    cloudwatch     = "http://localhost:4566"
    ec2            = "http://localhost:4566"
    ecr            = "http://localhost:4566"
    ecs            = "http://localhost:4566"
    iam            = "http://localhost:4566"
    logs           = "http://localhost:4566"
    rds            = "http://localhost:4566"
    s3             = "http://localhost:4566"
    secretsmanager = "http://localhost:4566"
    sts            = "http://localhost:4566"
  }
}
```

## 確認に向いているもの

- Terraform が AWS API に対して作成する resource の形。
- S3、ECR、CloudWatch Logs、Secrets Manager などの API 作成・取得。
- RDS や ECS task definition などの API 作成確認。
- AWS CLI や SDK の endpoint override の動作。

## AWS 実環境で確認するもの

- VPC、subnet、route table の実ネットワーク挙動。
- security group による実通信制御。
- ALB target group health check。
- ECS Fargate の本番相当実行。
- RDS の性能、backup、PITR、private subnet 内到達性。
- GitHub Actions OIDC による AWS 実認証。

## Docker Network

`compose.ministack.yaml` は `DOCKER_NETWORK` の既定値を `study_aws_default` にしている。
Docker Compose の project name を変える場合は、起動時にネットワーク名を指定する。

```bash
MINISTACK_DOCKER_NETWORK=<compose-project>_default make ministack-up
```

## 参考

- MiniStack docs: https://ministack.org/docs/
- MiniStack IaC docs: https://ministack.org/docs/iac
