# 12週間の学習ロードマップ

## 操作場所の表記

- `[ローカル]`: ローカルPC上のファイル編集、テスト、Docker、Terraform fmt/validate/plan など。
- `[MiniStack]`: MiniStack をローカルで起動し、AWS 互換 API に対して Terraform、AWS CLI、SDK の一部動作を検証する操作。
- `[AWS]`: AWS マネジメントコンソール、AWS CLI、または `terraform apply/destroy` で AWS 上のリソースを作成・変更・削除する操作。
- `[GitHub]`: GitHub リポジトリ、Pull Request、GitHub Actions、OIDC 設定に関する操作。
- `[ローカル/MiniStack]` や `[ローカル/AWS]` のような複合表記: ローカルファイルで手順を作成・検証し、必要に応じて MiniStack や AWS 上の状態確認を伴う操作。

MiniStack は AWS API の学習と一部統合確認に使う。VPC の実ネットワーク分離、security group の実通信制御、ALB health check、Fargate 実行、GitHub OIDC の実認証は AWS 実環境で確認する。
AWS 上で実リソースを変更する操作は、対象リソース、概算コスト、戻し方を確認してから明示的な承認を得る。

## 1週目

### 要件

- テーマ: プロジェクト初期化、設定、ログ、HTTP サーバー。
- Go: モジュール、`net/http`、コンテキストの基礎。
- PostgreSQL: なし。
- AWS: 目標アーキテクチャの読み取りのみ。
- Terraform: なし。
- SRE: 基本的なヘルスチェックの考え方。
- 完了条件: ローカルサーバーが起動する。
- 工数: 6-10 時間。
- AWS コスト: なし。
- 戻し方: ローカルファイルを削除するかブランチをリセットする。

### タスク

- [x] 1-1-1. [ローカル] 実装: フェーズ 1 で Go プロジェクトの雛形を作成する。
- [x] 1-1-2. [ローカル] テスト: 最初のユニットテスト。
- [x] 1-1-3. [ローカル] 障害演習: 不正な環境変数。
	- DATABASE_URL='not-a-valid-postgres-url' make run


## 2週目

### 要件

- テーマ: PostgreSQL とマイグレーション。
- Go: `database/sql` または pgx の使い方、エラー処理。
- PostgreSQL: スキーマ、制約、インデックス。
- AWS: RDS の概念。
- Terraform: なし。
- SRE: データベースの準備完了状態。
- 完了条件: リポジトリが TODO を永続化する。
- 工数: 8-12 時間。
- AWS コスト: ローカルではなし。
- 戻し方: `make migrate-down`。

### タスク

- [x] 2-1-1. [ローカル] 実装: マイグレーションコマンド、リポジトリ。
- [x] 2-1-2. [ローカル] テスト: PostgreSQL を使ったリポジトリテスト。
- [x] 2-1-3. [ローカル] 障害演習: DB 接続失敗。
	- DATABASE_URL='postgres://todo:todo_password@localhost:15432/todo_api?sslmode=disable' make run
## 3週目

### 要件

- テーマ: TODO CRUD。
- Go: ハンドラー、JSON、バリデーション。
- PostgreSQL: バージョン確認付きの更新。
- AWS: なし。
- Terraform: なし。
- SRE: 明確なエラーレスポンス。
- 完了条件: CRUD がローカルで動作する。
- 工数: 8-14 時間。
- AWS コスト: なし。
- 戻し方: マイグレーションを戻し、コンテナを片付ける。

### タスク

- [x] 3-1-1. [ローカル] 実装: API ハンドラー、サービス、バリデーション、楽観的ロック。
- [x] 3-1-2. [ローカル] テスト: CRUD と競合のテスト。
- [x] 3-1-3. [ローカル] 障害演習: 同時更新。

## 4週目

### 要件

- テーマ: カーソルページングと品質ゲート。
- Go: テーブルテスト、レース検出。
- PostgreSQL: 安定した並び順。
- AWS: なし。
- Terraform: なし。
- SRE: テストによる信頼性。
- 完了条件: ページングとテストが成功する。
- 工数: 6-10 時間。
- AWS コスト: なし。
- 戻し方: ローカルコンテナを削除する。

### タスク

- [x] 4-1-1. [ローカル] 実装: カーソルのエンコード/デコード、HTTP テスト、lint。
- [x] 4-1-2. [ローカル] テスト: `go test -race ./...`
- [x] 4-1-3. [ローカル] 障害演習: 不正なカーソル。

## 5週目

### 要件

- テーマ: Docker とローカルでの運用動作。
- Go: シグナル、サーバー停止、ビルドメタデータ。
- PostgreSQL: コネクションプール。
- AWS: ECS ヘルスチェックの概念。
- Terraform: なし。
- SRE: 生存確認と準備完了確認。
- 完了条件: Compose で API と DB が起動する。
- 工数: 8-12 時間。
- AWS コスト: なし。
- 戻し方: `make compose-down`。

### タスク

- [x] 5-1-1. [ローカル] 実装: Dockerfile、compose、グレースフルシャットダウン、health、ready、version。
- [x] 5-1-2. [ローカル] テスト: health と readiness のテスト。
- [x] 5-1-3. [ローカル] 障害演習: API プロセスの停止。

## 6週目

### 要件

- テーマ: Terraform の基礎。
- Go: なし。
- PostgreSQL: なし。
- AWS: S3 state bucket と IAM の概念。
- Terraform: state、variable、output、test。
- SRE: 変更レビュー。
- 完了条件: Terraform がローカルで検証できる。
- 工数: 8-12 時間。
- AWS コスト: 承認後に apply した場合は S3 の可能性あり。
- 戻し方: 承認済みの destroy のみ。

### タスク

- [x] 6-1-1. [ローカル] 実装: bootstrap 設計と検証。
- [x] 6-1-2. [ローカル] テスト: `terraform fmt`、`validate`、`test`。
- [x] 6-1-3. [ローカル] 障害演習: 不正な variable。


## 7週目

### 要件

- テーマ: ネットワークと IAM。
- Go: なし。
- PostgreSQL: なし。
- AWS: VPC、SG、IAM。
- Terraform: module と環境変数。
- MiniStack: Terraform provider の endpoint override、VPC/SG/IAM API の作成練習。
- SRE: 影響範囲。
- 完了条件: plan の内容を理解できる。
- 工数: 10-14 時間。
- AWS コスト: plan ではなし。apply した場合はリソース費用が発生する。
- 戻し方: 承認済みの destroy のみ。

### タスク

#### 実装

- [x] 7-1-1. [ローカル] `infrastructure/app/` を作成し、Terraform ファイルを責務ごとに分ける。
	- `infrastructure/app/` と `infrastructure/app/envs/dev.tfvars.example` を作成する。

- [x] 7-1-2. [ローカル] `common-variables.tf` の `variable` ブロックで environment 名、AWS region、project 名、CIDR、AZ 数、共通 tags を定義する。
	- `common-versions.tf`、`common-providers.tf`、`common-variables.tf`、`common-locals.tf`、`common-outputs.tf`、`aws-subnet.tf`、`aws-security-groups.tf`、`aws-iam.tf` を作成する。
	- ubuntu では次のコマンドで空ファイルを作成する。
		```
		mkdir -p infrastructure/app/envs
		cd infrastructure/app
		touch common-versions.tf common-providers.tf common-variables.tf common-locals.tf common-outputs.tf
		touch aws-subnet.tf aws-security-groups.tf aws-iam.tf
		touch envs/dev.tfvars.example
		```

- [x] 7-1-3. [ローカル] `common-locals.tf` の `locals` ブロックで name prefix、共通 tags、subnet CIDR、AZ 選択を一元管理する。
	- S3 backend 設定は有効な `.tf` ファイルには書かず、`backend.tf.example` の `terraform` ブロック例に留める。
	- `infrastructure/app/backend.tf.example` を作成し、S3 backend のサンプルだけを書く。

- [x] 7-1-4. [ローカル] VPC の Terraform 定義を作成し、DNS hostnames と DNS support を有効化する。
	- `aws-subnet.tf` の `resource` ブロックで VPC を定義し、`common-variables.tf` の `variable` ブロックで CIDR、AZ 数、project、environment、region、tags と validation を管理する。

- [x] 7-1-5. [ローカル] public subnet と private subnet の Terraform 定義を 2 AZ 以上で作成し、role が分かる tag を付ける。
	- `common-locals.tf` の `locals` ブロックで name prefix、共通 tags、subnet CIDR、AZ 選択を組み立てる。

- [x] 7-1-6. [ローカル] internet gateway、public route table、public route、route table association の Terraform 定義を作成する。
	- `aws-subnet.tf` の `resource` ブロックで VPC、subnet、internet gateway、route table を追加する。

- [x] 7-1-7. [ローカル] private subnet はこの週では NAT gateway なしの設計にし、ECS/RDS の配置意図を docs に残す。
	- `docs/design/aws-architecture.md` に NAT Gateway なしの private subnet 方針と ECS/RDS の配置意図を追記する。

- [x] 7-1-8. [ローカル] ALB security group は internet から HTTP のみ受ける Terraform 定義にする。
	- `aws-security-groups.tf` の `resource` ブロックで ALB security group rule を定義する。

- [x] 7-1-9. [ローカル] ECS security group は ALB security group からのアプリ port のみ受ける Terraform 定義にする。
	- `aws-security-groups.tf` の `resource` ブロックで ECS security group rule を定義する。

- [x] 7-1-10. [ローカル] RDS security group は ECS security group からの PostgreSQL port のみ受ける Terraform 定義にする。
	- `aws-security-groups.tf` の `resource` ブロックで RDS security group rule を定義する。

- [x] 7-1-11. [ローカル] ECS task execution role と task role を分けて Terraform 定義を作成する。
	- `aws-iam.tf` の `resource` ブロックで ECS task execution role と task role を追加する。

- [x] 7-1-12. [ローカル] task execution role は ECR pull と CloudWatch Logs 書き込みに必要な権限だけを付ける Terraform 定義にする。
	- `aws-iam.tf` の `resource` ブロックで task execution role の policy attachment を定義する。

- [x] 7-1-13. [ローカル] task role は初期状態では最小権限にする。
	- `aws-iam.tf` の `resource` ブロックで task role の権限を最小に保つ。

- [x] 7-1-14. [ローカル] `common-outputs.tf` の `output` ブロックで VPC ID、subnet IDs、security group IDs、IAM role ARNs を追加する。
	- `common-outputs.tf` の `output` ブロックに後続週で参照する ID と ARN を追加する。

- [x] 7-1-15. [ローカル] docs/design/aws-architecture.md にネットワーク境界、subnet 用途、IAM role の責務を追記する。
	- docs/design/aws-architecture.md にネットワークと IAM の設計意図を追記する。

#### テスト


- [x] 7-2-1. [ローカル] `terraform -chdir=infrastructure/app fmt -recursive` を実行する。
	- `terraform fmt`、`init -backend=false`、`validate` を実行する。

- [x] 7-2-2. [ローカル] `terraform -chdir=infrastructure/app init -backend=false` を実行する。
	- `terraform plan -refresh=false` を実行し、作成 resource と依存関係を読む。

- [x] 7-2-3. [ローカル] `terraform -chdir=infrastructure/app validate` を実行する。
	- TFLint 設定を追加し、`tflint --init` と `tflint` を実行する。

- [x] 7-2-4. [ローカル] `terraform -chdir=infrastructure/app plan -input=false -refresh=false -var-file=envs/dev.tfvars` を実行する。
	- tfsec または Checkov を選び、Makefile にセキュリティチェックコマンドを追加する。

- [x] 7-2-5. [ローカル] TFLint 設定ファイルを追加し、AWS plugin を有効化する。
	- plan とセキュリティチェックで public/private 境界、広すぎる ingress、IAM policy を確認する。

- [x] 7-2-6. [ローカル] `tflint --chdir=infrastructure/app --init` を実行する。
	- AWS apply は行わず、必要な場合は作成 resource と概算コストを報告して承認を得る。

- [x] 7-2-7. [ローカル] `tflint --chdir=infrastructure/app` を実行する。

- [x] 7-2-8. [ローカル] tfsec または Checkov のどちらを使うか決め、Makefile にコマンドを追加する。

- [x] 7-2-9. [ローカル] セキュリティチェックで public ingress、RDS 公開、広すぎる IAM policy の指摘を確認する。

- [x] 7-2-10. [ローカル] plan にこの週で意図しない有料 resource が含まれていないことを確認する。

- [x] 7-2-11. [ローカル/MiniStack] MiniStack 用の provider override 手順を docs に記録し、Terraform が MiniStack endpoint に向くことを確認する。
	- VPC、subnet、security group、IAM role は MiniStack 上では API 作成確認に留め、実ネットワーク制御や IAM 権限評価は AWS 実環境で確認する前提を明記する。

#### 障害演習


- [x] 7-3-1. [ローカル] ALB security group の HTTP ingress を一時的に削除し、plan で外部入口が消えることを確認する。
	- 障害演習用の一時変更は1種類ずつ入れる。

- [x] 7-3-2. [ローカル] ECS security group の ALB からの ingress を一時的に削除し、通信断の影響を説明する。
	- 変更後に plan、TFLint、セキュリティチェックのどれで検出できるか確認する。

- [x] 7-3-3. [ローカル] RDS security group を一時的に `0.0.0.0/0` に広げ、セキュリティチェックが検出することを確認する。
	- 検出できないものは、人手レビューで見る観点として docs に記録する。

- [x] 7-3-4. [ローカル] IAM policy に一時的に `Action = "*"` を入れ、検出できるか確認する。
	- 一時変更を必ず戻し、正常な検証コマンドが通ることを確認する。

- [x] 7-3-5. [ローカル] すべての一時変更を戻し、正常な検証コマンドが通ることを確認する。

- [x] 7-3-6. [ローカル] 演習で見つかった検出可否と人手レビュー観点を docs に記録する。

## 8週目

### 要件

- テーマ: アプリケーション実行基盤。
- Go: コンテナの準備完了状態。
- PostgreSQL: なし。
- AWS: ECR、ECS、ALB、ログ。
- Terraform: task definition と service。
- MiniStack: ECR、ECS task definition、CloudWatch Logs API の作成練習。
- SRE: デプロイの健全性。
- 完了条件: plan から API をデプロイできる。
- 工数: 10-16 時間。
- AWS コスト: apply した場合は ALB、ECS、ログ。
- 戻し方: 承認済みの destroy のみ。

### タスク

#### 実装

- [x] 8-1-1. [ローカル] ECR repository の Terraform 定義を追加し、image tag immutability と lifecycle policy を設計する。
	- ECR、CloudWatch Logs、ALB、ECS 用の Terraform ファイルを既存構成に追加する。

- [x] 8-1-2. [ローカル] CloudWatch Logs log group の Terraform 定義を追加し、retention days を variable で管理する。
	- ECR repository 名、image tag、container port、health path、desired count を variable 化する。

- [x] 8-1-3. [ローカル] ALB、target group、listener、health check の Terraform 定義を追加する。
	- ALB は public subnet、ECS service は private subnet に配置する。

- [x] 8-1-4. [ローカル] ECS cluster の Terraform 定義を追加する。
	- ALB security group から ECS security group への通信だけを許可する。

- [x] 8-1-5. [ローカル] ECS task definition の Terraform 定義を追加し、container image、port mapping、health check、log configuration を設定する。
	- task execution role に ECR pull と logs 書き込み権限を付ける。

- [x] 8-1-6. [ローカル] ECS service の Terraform 定義を追加し、Fargate、private subnet、ALB target group と紐付ける。
	- docs/design/aws-architecture.md に ALB から ECS task までの通信経路を追記する。

- [x] 8-1-7. [ローカル] ECS service の desired count、deployment min/max healthy percent を variable にする。

- [x] 8-1-8. [ローカル] task definition の CPU、memory、container port、health path を variable にする。

- [x] 8-1-9. [ローカル] Go API の `/healthz` と `/readyz` が ALB/ECS health check の意図に合うことを確認する。

- [x] 8-1-10. [ローカル] docs/design/aws-architecture.md に ECS、ALB、CloudWatch Logs の関係を追記する。

#### テスト

- [x] 8-2-1. [ローカル] `make build` を実行する。
	- Docker image をローカルで build し、API が container 内で起動することを確認する。

- [x] 8-2-2. [ローカル] `docker build` または `make docker-build` を実行する。
	- Go の health endpoint と ALB target group health check path を照合する。

- [x] 8-2-3. [ローカル] `docker run` または compose で container が起動することを確認する。
	- Terraform fmt、validate、plan を実行する。

- [x] 8-2-4. [ローカル] `terraform -chdir=infrastructure/app fmt -recursive` を実行する。
	- plan で ALB、ECS service、task definition、target group の依存関係を読む。

- [x] 8-2-5. [ローカル] `terraform -chdir=infrastructure/app validate` を実行する。
	- セキュリティチェックで ALB public ingress と ECS private placement を確認する。

- [x] 8-2-6. [ローカル] `terraform -chdir=infrastructure/app plan -input=false -refresh=false -var-file=envs/dev.tfvars` を実行する。
	- AWS apply は行わず、必要な場合は ALB、ECS、Logs、ECR の概算コストを報告して承認を得る。

- [x] 8-2-7. [ローカル] plan に ALB、ECS、ECR、CloudWatch Logs が含まれ、RDS はまだ含まれないことを確認する。

- [x] 8-2-8. [ローカル] ALB target group health check path と Go API endpoint が一致することを確認する。

- [x] 8-2-9. [ローカル] セキュリティチェックで ALB public ingress と ECS private placement の妥当性を確認する。

- [x] 8-2-10. [ローカル/MiniStack] MiniStack で ECR repository、CloudWatch Logs log group、ECS cluster/task definition の作成確認を行う。
	- ALB health check、Fargate の実行、security group による通信制御は MiniStack では本番相当確認としない。
	- local MiniStack endpoint に対して `terraform apply` を実行し、state で作成済み resource を確認する。

#### 障害演習


- [x] 8-3-1. [ローカル] ECS task の container port を一時的に誤設定し、ALB health check が失敗する構成になることを plan で確認する。
	- 障害演習用の一時変更を1つずつ入れる。

- [x] 8-3-2. [ローカル] ECS service desired count を一時的に `0` にし、影響範囲を説明できることを確認する。
	- health check、desired count、ログ欠落が運用に与える影響を言語化する。

- [x] 8-3-3. [ローカル] CloudWatch Logs 設定を一時的に外し、運用時に何が見えなくなるかを記録する。
	- 一時変更を戻し、`make test`、Docker build、Terraform 検証を再実行する。

- [x] 8-3-4. [ローカル] すべての一時変更を戻し、Docker build と Terraform 検証が成功することを確認する。

## 9週目

### 要件

- テーマ: RDS とシークレット。
- Go: シークレット由来の DB 設定。
- PostgreSQL: バックアップ、プール調整。
- AWS: RDS、Secrets Manager。
- Terraform: sensitive value。
- MiniStack: Secrets Manager と RDS API の作成・取得確認。
- SRE: バックアップの基礎。
- 完了条件: RDS 設計と plan の準備ができている。
- 工数: 10-16 時間。
- AWS コスト: apply した場合は RDS と Secrets Manager。
- 戻し方: スナップショットを取得してから、承認済みの destroy。

### タスク

#### 実装


- [x] 9-1-1. [ローカル] RDS subnet group の Terraform 定義を private subnet で作成する。
	- RDS subnet group、parameter group、instance、Secrets Manager secret を Terraform に追加する。

- [x] 9-1-2. [ローカル] RDS PostgreSQL instance の Terraform 定義を追加し、publicly accessible を false にする。
	- RDS は private subnet のみ、publicly accessible は false にする。

- [x] 9-1-3. [ローカル] RDS parameter group の Terraform 定義を追加し、必要な PostgreSQL 設定を明示する。
	- DB 認証情報は secret として扱い、`.env` や `tfvars` に実値を書かない。

- [x] 9-1-4. [ローカル] DB name、username、port、engine version、instance class、storage を variable 化する。
	- ECS task role に secret 読み取り権限を追加し、対象 secret ARN に絞る。

- [x] 9-1-5. [ローカル] DB password は Terraform 変数の平文ではなく Secrets Manager または random password 連携で管理する。
	- Go API の config 読み込みを、ローカル `.env` と AWS secret の両方で説明できる形にする。

- [x] 9-1-6. [ローカル] Secrets Manager secret と secret version の Terraform 定義を追加する。
	- docs/design/database-design.md と docs/design/aws-architecture.md に DB 境界と backup 方針を追記する。

- [x] 9-1-7. [ローカル] ECS task role に Secrets Manager read 権限を最小範囲で追加する。

- [x] 9-1-8. [ローカル] Go API の DB 設定を secret 由来で組み立てられるようにする。

- [x] 9-1-9. [ローカル] readiness は DB 接続失敗時に失敗を返すことを確認する。

- [x] 9-1-10. [ローカル] docs/design/database-design.md に RDS、backup、connection pool 方針を追記する。

- [x] 9-1-11. [ローカル] docs/design/aws-architecture.md に RDS と Secrets Manager の接続境界を追記する。

#### テスト


- [x] 9-2-1. [ローカル] `make test` を実行する。
	- readiness が DB 接続を検査し、障害時に失敗することをテストする。

- [x] 9-2-2. [ローカル] DB 設定読み込みの unit test を追加または更新する。
	- DB 設定読み込みの unit test を追加または更新する。

- [x] 9-2-3. [ローカル] readiness の DB 成功/失敗テストを追加または更新する。
	- Terraform fmt、validate、plan を実行する。

- [x] 9-2-4. [ローカル] `terraform fmt` と `terraform validate` を実行する。
	- plan とセキュリティチェックで RDS 公開、暗号化、backup、secret leakage を確認する。

- [x] 9-2-5. [ローカル] `terraform plan -input=false -refresh=false` で RDS、subnet group、parameter group、Secrets Manager の差分を確認する。
	- AWS apply は行わず、必要な場合は RDS と Secrets Manager の概算コストを報告して承認を得る。

- [x] 9-2-6. [ローカル] セキュリティチェックで RDS 非公開、暗号化、backup retention、secret の扱いを確認する。

- [x] 9-2-7. [ローカル] plan に DB password や接続文字列が表示されないことを確認する。

- [x] 9-2-8. [ローカル/MiniStack] MiniStack で Secrets Manager secret の作成・取得と RDS instance API の作成確認を行う。
	- RDS の性能、backup/PITR、private subnet 内の到達性は MiniStack では本番相当確認としない。

#### 障害演習


- [x] 9-3-1. [ローカル] Secrets Manager secret ARN を一時的に誤設定し、ECS task が DB 設定を読めない影響を整理する。
	- secret 失敗、security group 失敗、password 欠落を1つずつ作る。

- [x] 9-3-2. [ローカル] RDS security group の ECS ingress を一時的に削除し、readiness が失敗する想定を記録する。
	- それぞれの失敗が config validation、readiness、Terraform plan のどこで見えるか確認する。

- [x] 9-3-3. [ローカル] DB password を欠落させた設定で config validation が失敗することを確認する。
	- 影響範囲を記録し、一時変更を戻して Go test と Terraform 検証を通す。

- [x] 9-3-4. [ローカル] すべての一時変更を戻し、Go test と Terraform 検証が成功することを確認する。

## 10週目

### 要件

- テーマ: CI/CD。
- Go: CI チェック。
- PostgreSQL: マイグレーション手順の設計。
- AWS: OIDC role、ECR push、ECS update。
- Terraform: PR での plan。
- SRE: デプロイゲート。
- 完了条件: CI/CD workflow が存在する。
- 工数: 8-14 時間。
- AWS コスト: 使用した場合は GitHub Actions minutes と ECR storage。
- 戻し方: workflow を削除するか secret を無効化する。

### タスク

#### 実装


- [x] 10-1-1. [ローカル] GitHub Actions の CI workflow ファイルを追加し、Go test、race test、lint、vuln、Docker build を実行する設定を書く。
	- `.github/workflows/ci.yml` を追加し、Go と Docker の基本チェックを定義する。

- [x] 10-1-2. [ローカル] Terraform check workflow ファイルを追加し、fmt、validate、TFLint、セキュリティチェックを実行する設定を書く。
	- `.github/workflows/terraform.yml` を追加し、Terraform fmt、validate、TFLint、セキュリティチェックを定義する。

- [x] 10-1-3. [ローカル] GitHub OIDC 用 IAM role と trust policy の Terraform 定義を追加する。
	- OIDC role と policy を Terraform に追加し、GitHub repository と branch 条件を絞る。

- [x] 10-1-4. [ローカル] OIDC role の権限を ECR push、ECS update、Terraform plan に必要な範囲へ分割する。
	- `.github/workflows/deploy.yml` は deploy gate 付きで設計し、必要なら disabled または manual dispatch にする。

- [x] 10-1-5. [ローカル] ECR push workflow を設計し、main branch または tag で image push する条件を定義する。
	- migration を deploy 前後のどちらで実行するか、失敗時の扱いを docs に記録する。

- [x] 10-1-6. [ローカル] ECS deploy workflow を設計し、task definition 更新と service update の流れを定義する。
	- 本番相当 deploy は自動実行しない設計にする。

- [x] 10-1-7. [ローカル] Terraform plan を PR comment または artifact として確認できる workflow 設定にする。

- [x] 10-1-8. [ローカル] migration 実行タイミングを設計し、自動実行するか手動 gate にするか決める。

- [x] 10-1-9. [ローカル] deploy gate と rollback 手順を docs に記録する。

#### テスト


- [x] 10-2-1. [ローカル] `act` または GitHub Actions の dry run 相当で workflow 構文を確認する。
	- workflow の構文を actionlint などで検査する。

- [x] 10-2-2. [ローカル] `yamllint` または actionlint を導入して workflow を検査する。
	- OIDC trust policy の subject 条件と IAM policy の権限範囲を読む。

- [x] 10-2-3. [GitHub] pull request 上で CI が失敗した場合に merge できない設定を確認する。
	- CI、Terraform check、image push、deploy の実行条件を確認する。

- [x] 10-2-4. [ローカル] OIDC trust policy の subject 条件が repository、branch、environment に絞られていることを確認する。
	- Terraform plan artifact と workflow logs に secret が出ないことを確認する。

- [x] 10-2-5. [ローカル] ECR push と ECS deploy の権限に `*` が不要に含まれていないことを確認する。
	- 必要な場合は対象環境、権限、戻し方を報告して承認を得る。

- [x] 10-2-6. [GitHub] Terraform plan artifact に secret が含まれないことを確認する。

#### 障害演習


- [x] 10-3-1. [ローカル/GitHub] Go test を一時的に失敗させ、CI が止まることを確認する。
	- CI、Docker build、AWS auth、deploy の失敗経路を1つずつ作る。

- [x] 10-3-2. [ローカル/GitHub] Docker build を一時的に失敗させ、image push が実行されないことを確認する。
	- 後続 job が止まること、または deploy gate で止まることを確認する。

- [x] 10-3-3. [ローカル/GitHub/AWS] OIDC role ARN を一時的に誤設定し、AWS 認証で失敗することを確認する。
	- 失敗時の戻し方を docs に記録する。

- [x] 10-3-4. [ローカル/GitHub/AWS] ECS deploy の task definition image tag を一時的に存在しない値にし、失敗時の戻し方を確認する。
	- 一時変更を戻し、workflow lint とローカル test を再実行する。

- [x] 10-3-5. [ローカル] すべての一時変更を戻し、workflow lint とローカル test が成功することを確認する。

## 11週目

### 要件

- テーマ: オブザーバビリティと SLO。
- Go: メトリクスフック。
- PostgreSQL: DB メトリクス。
- AWS: CloudWatch alarm。
- Terraform: alarm resource。
- SRE: アラートポリシー。
- 完了条件: アラートポリシーと runbook が存在する。
- 工数: 8-12 時間。
- AWS コスト: apply した場合は CloudWatch alarm とログ。
- 戻し方: 承認済みの destroy のみ。

### タスク

#### 実装


- [x] 11-1-1. [ローカル] SLI を availability、latency、error rate、DB readiness で定義する。
	- docs に SLI/SLO と alarm policy を追加する。

- [x] 11-1-2. [ローカル] SLO と error budget の初期値を docs に記録する。
	- Go API の request/latency/error 計測箇所を決める。

- [x] 11-1-3. [ローカル] Go API に request count、latency、error count のメトリクス計測方針を追加する。
	- CloudWatch alarm と dashboard を Terraform に追加する。

- [x] 11-1-4. [ローカル] structured logging の必須項目を整理する。
	- runbook を障害種別ごとに作成する。

- [x] 11-1-5. [ローカル] CloudWatch alarm の Terraform 定義を ALB 5xx、target unhealthy、ECS task count、RDS CPU/storage/connections に分けて追加する。
	- alarm threshold、evaluation period、treat missing data の理由を docs に記録する。

- [x] 11-1-6. [ローカル] alarm threshold、evaluation period、treat missing data の理由を docs に記録する。

- [x] 11-1-7. [ローカル] runbook を target unhealthy、high error rate、DB connection failure、deploy rollback で作成する。

- [x] 11-1-8. [ローカル] dashboard の初期構成を設計する。

#### テスト


- [x] 11-2-1. [ローカル] Go のメトリクス追加箇所に unit test または handler test を追加する。
	- Go test と Terraform 検証を実行する。

- [x] 11-2-2. [ローカル] `make test` を実行する。
	- Terraform plan で alarm の対象 metric、threshold、period を読む。

- [x] 11-2-3. [ローカル] `terraform fmt` と `terraform validate` を実行する。
	- runbook が「検知」「一次確認」「切り分け」「緩和」「復旧確認」「事後対応」を含むことを確認する。

- [x] 11-2-4. [ローカル] Terraform plan で alarm、dashboard、log metric filter の差分を確認する。
	- alarm apply は行わず、必要な場合は CloudWatch alarm と dashboard の概算コストを報告して承認を得る。

- [x] 11-2-5. [ローカル] alarm threshold が通常負荷で誤検知しにくく、障害時に検出できる値かレビューする。

- [x] 11-2-6. [ローカル] runbook が「検知」「一次確認」「切り分け」「緩和」「復旧確認」「事後対応」を含むことを確認する。

#### 障害演習


- [x] 11-3-1. [ローカル] target group health check path を一時的に誤設定し、target unhealthy の検知経路を確認する。
	- target unhealthy、5xx、DB readiness failure の想定を1つずつ作る。

- [x] 11-3-2. [ローカル] API が 500 を返す想定を作り、ALB 5xx alarm の発火条件を確認する。
	- runbook に従って検知から復旧確認までを机上確認する。

- [x] 11-3-3. [ローカル] DB 接続失敗時の readiness と alarm の関係を確認する。
	- runbook の不足を更新する。

- [x] 11-3-4. [ローカル] runbook に従って一次確認から復旧確認までを机上演習する。
	- 一時変更を戻し、Go test と Terraform 検証を通す。

- [x] 11-3-5. [ローカル] 一時変更を戻し、Go test と Terraform 検証が成功することを確認する。

## 12週目

### 要件

- テーマ: レジリエンス、復旧、コスト、最終ドキュメント。
- Go: 演習から見つかった運用上の修正。
- PostgreSQL: PITR とリストア手順。
- AWS: RDS snapshot と restore。
- Terraform: drift 対応。
- SRE: インシデント演習。
- 完了条件: README でローカル環境を再現でき、ドキュメントで AWS 運用を説明している。
- 工数: 10-18 時間。
- AWS コスト: apply したリソースとリストア演習による。
- 戻し方: 承認済みの destroy とクリーンアップ確認。

### タスク

#### 実装


- [x] 12-1-1. [ローカル] RDS backup retention、maintenance window、deletion protection の方針を docs に整理する。
	- RDS backup、snapshot restore、PITR、rollback、drift、cost review の docs を作成する。

- [x] 12-1-2. [ローカル] RDS snapshot と restore の手順書を作成する。
	- README にローカル再現手順、検証コマンド、AWS 操作時の承認ルールをまとめる。

- [x] 12-1-3. [ローカル] PITR の前提条件、復旧ポイントの決め方、復旧後の接続先切替を整理する。
	- postmortem template と cost review checklist を作成する。

- [x] 12-1-4. [ローカル] Terraform drift の確認手順を作成する。
	- cleanup checklist に対象 resource、確認コマンド、残課金観点をまとめる。

- [x] 12-1-5. [ローカル] rollback 手順を application、infrastructure、database migration に分けて作成する。

- [x] 12-1-6. [ローカル] postmortem template を作成する。

- [x] 12-1-7. [ローカル/AWS] AWS cost review checklist を作成し、ALB、ECS、RDS、NAT、CloudWatch、Secrets Manager、ECR の確認観点を整理する。

- [x] 12-1-8. [ローカル] README を更新し、ローカル起動、テスト、MiniStack 検証、Terraform 検証、AWS 運用の入口をまとめる。

- [x] 12-1-9. [ローカル/AWS] 不要 resource の cleanup checklist を作成する。

#### テスト


- [x] 12-2-1. [ローカル] `make test` を実行する。
	- 最終テストスイートを実行し、失敗があれば修正する。

- [x] 12-2-2. [ローカル] `make test-race` を実行する。
	- Terraform 全ディレクトリで fmt、validate、lint、security check を実行する。

- [x] 12-2-3. [ローカル] `make lint` を実行する。
	- README の手順だけでローカル環境を再現できることを確認する。

- [x] 12-2-4. [ローカル] `make vuln` を実行する。
	- runbook、restore 手順、rollback 手順を読み合わせ、不足を修正する。

- [x] 12-2-5. [ローカル] `make build` を実行する。

- [x] 12-2-6. [ローカル] Docker build を実行する。

- [x] 12-2-7. [ローカル] Terraform fmt、validate、TFLint、セキュリティチェックを全ディレクトリで実行する。

- [x] 12-2-8. [ローカル] README の手順だけでローカル環境を再現できることを確認する。

- [x] 12-2-9. [ローカル] runbook、restore 手順、rollback 手順を読み合わせ、不足を修正する。

#### 障害演習


- [x] 12-3-1. [ローカル/AWS] RDS snapshot restore の手順を机上演習し、必要な入力値と停止時間を整理する。
	- backup restore と rollback は、実リソースを操作せず机上演習で手順の穴を確認する。

- [x] 12-3-2. [ローカル/AWS] PITR の復旧ポイントを仮定し、復旧後にアプリが参照する DB endpoint の切替手順を確認する。
	- PITR、ECS rollback、Terraform drift、postmortem、cleanup を1つずつ確認する。

- [x] 12-3-3. [ローカル/AWS] 直前 image への ECS rollback 手順を確認する。
	- 実リソースで restore や destroy を行う場合は、対象 resource、想定停止時間、概算コスト、戻し方を報告して承認を得る。

- [x] 12-3-4. [ローカル] Terraform drift を仮定し、plan 結果から修正方針を決める演習を行う。
	- postmortem template と cost review checklist を使い、最終レビュー結果を docs に残す。

- [x] 12-3-5. [ローカル] postmortem template に仮想障害を記入し、再発防止 action を作る。

- [x] 12-3-6. [ローカル/AWS] cleanup checklist に沿って、残課金 resource がないか確認する手順を読む。
