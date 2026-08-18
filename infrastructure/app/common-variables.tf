# 手動編集: 可 - 入力変数と検証を意図して変更する。
# リソース名とタグに使う環境名。
variable "environment" {
  description = "Deployment environment name, such as dev, stg, or prod."
  type        = string
  default     = "dev"

  # 2-21文字で、小文字英字から始まり、小文字英数字とハイフンだけを許可する。
  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{1,20}$", var.environment))
    error_message = "Environment must be 2-21 characters, start with a lowercase letter, and contain only lowercase letters, numbers, and hyphens."
  }
}

# アプリケーションインフラを作成する AWS リージョン。
variable "aws_region" {
  description = "AWS region for application infrastructure."
  type        = string
  default     = "ap-northeast-1"

  # ap-northeast-1 のように、2文字英字-英字の名前-数字の形式だけを許可する。
  validation {
    condition     = can(regex("^[a-z]{2}-[a-z]+-[0-9]+$", var.aws_region))
    error_message = "Aws_region must be a valid-looking AWS region name, such as ap-northeast-1."
  }
}

# ローカル plan や MiniStack 向けに AWS 認証情報の検証を省略するかどうか。
variable "skip_credentials_validation" {
  description = "Skip AWS credential validation for local or emulated provider usage."
  type        = bool
  default     = false
}

# ローカル plan や MiniStack 向けに EC2 metadata API の確認を省略するかどうか。
variable "skip_metadata_api_check" {
  description = "Skip EC2 metadata API check for local or emulated provider usage."
  type        = bool
  default     = false
}

# ローカル plan や MiniStack 向けに AWS account ID の取得を省略するかどうか。
variable "skip_requesting_account_id" {
  description = "Skip requesting AWS account ID for local or emulated provider usage."
  type        = bool
  default     = false
}

# MiniStack などの AWS 互換 API に向けるための endpoint URL。
# null の場合は AWS provider の標準 endpoint を使う。
variable "aws_endpoint_url" {
  description = "Optional AWS-compatible endpoint URL for MiniStack or local emulators."
  type        = string
  default     = null

  # null、または http:// / https:// で始まる URL だけを許可する。
  validation {
    condition     = var.aws_endpoint_url == null || can(regex("^https?://", var.aws_endpoint_url))
    error_message = "Aws_endpoint_url must be null or start with http:// or https://."
  }
}

# リソース名とタグに使う短いプロジェクト名。
variable "project_name" {
  description = "Project name used as a prefix for application resources."
  type        = string
  default     = "study-aws"

  # 3-31文字で、小文字英字から始まり、小文字英数字とハイフンだけを許可する。
  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{2,30}$", var.project_name))
    error_message = "Project_name must be 3-31 characters, start with a lowercase letter, and contain only lowercase letters, numbers, and hyphens."
  }
}

# アプリケーション VPC 用の CIDR ブロック。
variable "vpc_cidr" {
  description = "IPv4 CIDR block for the application VPC."
  type        = string
  default     = "10.0.0.0/16"

  # cidrhost で先頭アドレスを計算できる、妥当な CIDR ブロックだけを許可する。
  validation {
    condition     = can(cidrhost(var.vpc_cidr, 0))
    error_message = "Vpc_cidr must be a valid IPv4 CIDR block, such as 10.0.0.0/16."
  }
}

# パブリックサブネットとプライベートサブネットで使うアベイラビリティゾーン数。
variable "az_count" {
  description = "Number of availability zones to use for subnet placement."
  type        = number
  default     = 2

  # アベイラビリティゾーン数は 2 以上 3 以下だけを許可する。
  validation {
    condition     = var.az_count >= 2 && var.az_count <= 3
    error_message = "Az_count must be between 2 and 3 for this learning environment."
  }
}

# ローカル plan 用に明示する AZ 名一覧。
# 空の場合は AWS API から利用可能な AZ 一覧を取得する。
variable "availability_zone_names" {
  description = "Optional availability zone names for local plan without querying AWS."
  type        = list(string)
  default     = []

  # 空の list、または ap-northeast-1a のような AZ 名だけを許可する。
  validation {
    condition = alltrue([
      for az in var.availability_zone_names :
      can(regex("^[a-z]{2}-[a-z]+-[0-9]+[a-z]$", az))
    ])
    error_message = "Availability_zone_names must be empty or contain valid-looking AZ names, such as ap-northeast-1a."
  }
}

# ECS タスク内の Go API が待ち受ける HTTP ポート。
variable "app_port" {
  description = "Application container port exposed by the ECS task."
  type        = number
  default     = 8080

  # TCP ポートとして使える 1-65535 の範囲だけを許可する。
  validation {
    condition     = var.app_port >= 1 && var.app_port <= 65535
    error_message = "App_port must be between 1 and 65535."
  }
}

# RDS PostgreSQL が待ち受けるポート。
variable "db_port" {
  description = "PostgreSQL port used by RDS."
  type        = number
  default     = 5432

  # TCP ポートとして使える 1-65535 の範囲だけを許可する。
  validation {
    condition     = var.db_port >= 1 && var.db_port <= 65535
    error_message = "Db_port must be between 1 and 65535."
  }
}

# アプリケーション用 PostgreSQL database 名。
variable "db_name" {
  description = "PostgreSQL database name for the TODO API."
  type        = string
  default     = "todo_api"

  validation {
    condition     = can(regex("^[A-Za-z][A-Za-z0-9_]{0,62}$", var.db_name))
    error_message = "Db_name must start with a letter and contain only letters, numbers, or underscores."
  }
}

# アプリケーション用 PostgreSQL username。
variable "db_username" {
  description = "PostgreSQL username for the TODO API."
  type        = string
  default     = "todo"

  validation {
    condition     = can(regex("^[A-Za-z][A-Za-z0-9_]{0,62}$", var.db_username))
    error_message = "Db_username must start with a letter and contain only letters, numbers, or underscores."
  }
}

# RDS PostgreSQL engine version。
variable "db_engine_version" {
  description = "RDS PostgreSQL engine version."
  type        = string
  default     = "16"

  validation {
    condition     = can(regex("^[0-9]+(\\.[0-9]+)?$", var.db_engine_version))
    error_message = "Db_engine_version must be a PostgreSQL major or minor version such as 16 or 16.8."
  }
}

# RDS parameter group family。
variable "db_parameter_group_family" {
  description = "RDS PostgreSQL parameter group family."
  type        = string
  default     = "postgres16"

  validation {
    condition     = can(regex("^postgres[0-9]+$", var.db_parameter_group_family))
    error_message = "Db_parameter_group_family must look like postgres16."
  }
}

# RDS instance class。
variable "db_instance_class" {
  description = "RDS instance class for the PostgreSQL database."
  type        = string
  default     = "db.t4g.micro"

  validation {
    condition     = can(regex("^db\\.[A-Za-z0-9]+\\.[A-Za-z0-9]+$", var.db_instance_class))
    error_message = "Db_instance_class must look like db.t4g.micro."
  }
}

# RDS allocated storage in GiB。
variable "db_allocated_storage" {
  description = "Allocated RDS storage in GiB."
  type        = number
  default     = 20

  validation {
    condition     = var.db_allocated_storage >= 20 && var.db_allocated_storage <= 100
    error_message = "Db_allocated_storage must be between 20 and 100 GiB for this learning environment."
  }
}

# RDS max allocated storage in GiB for autoscaling。
variable "db_max_allocated_storage" {
  description = "Maximum RDS storage in GiB for storage autoscaling."
  type        = number
  default     = 50

  validation {
    condition     = var.db_max_allocated_storage >= 20 && var.db_max_allocated_storage <= 100
    error_message = "Db_max_allocated_storage must be between 20 and 100 GiB for this learning environment."
  }
}

# RDS backup retention period in days。
variable "db_backup_retention_days" {
  description = "RDS automated backup retention period in days."
  type        = number
  default     = 7

  validation {
    condition     = var.db_backup_retention_days >= 1 && var.db_backup_retention_days <= 35
    error_message = "Db_backup_retention_days must be between 1 and 35."
  }
}

# ECS API が PostgreSQL に接続するときの SSL mode。
variable "db_sslmode" {
  description = "PostgreSQL sslmode used by the API when connecting to RDS."
  type        = string
  default     = "require"

  validation {
    condition     = contains(["disable", "allow", "prefer", "require", "verify-ca", "verify-full"], var.db_sslmode)
    error_message = "Db_sslmode must be a valid PostgreSQL sslmode."
  }
}

# ECR repository の名前。
# 実際の名前には environment を含む local.name_prefix を前につける。
variable "ecr_repository_name" {
  description = "Short ECR repository name for the API image."
  type        = string
  default     = "todo-api"

  # 2-128文字で、小文字英数字、ハイフン、アンダースコア、スラッシュだけを許可する。
  validation {
    condition     = can(regex("^[a-z0-9][a-z0-9_/-]{1,127}$", var.ecr_repository_name))
    error_message = "Ecr_repository_name must be 2-128 characters and contain only lowercase letters, numbers, underscores, hyphens, or slashes."
  }
}

# ECS task definition が参照するコンテナイメージの tag。
variable "image_tag" {
  description = "Container image tag deployed by the ECS task definition."
  type        = string
  default     = "latest"

  # Docker tag として扱いやすい、空白を含まない 1-128文字だけを許可する。
  validation {
    condition     = can(regex("^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$", var.image_tag))
    error_message = "Image_tag must be a valid Docker tag-like value."
  }
}

# CloudWatch Logs にアプリログを残す日数。
variable "log_retention_days" {
  description = "CloudWatch Logs retention period in days."
  type        = number
  default     = 7

  # CloudWatch Logs が対応している retention days の値だけを許可する。
  validation {
    condition = contains([
      1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180,
      365, 400, 545, 731, 1096, 1827, 2192, 2557, 3653
    ], var.log_retention_days)
    error_message = "Log_retention_days must be one of the retention values supported by CloudWatch Logs."
  }
}

# ALB target group が使う readiness check の path。
variable "health_check_path" {
  description = "HTTP path used by the ALB target group health check."
  type        = string
  default     = "/readyz"

  # HTTP path として / から始まり、空白を含まない値だけを許可する。
  validation {
    condition     = can(regex("^/[^[:space:]]*$", var.health_check_path))
    error_message = "Health_check_path must start with / and must not contain whitespace."
  }
}

# ECS container health check が使う liveness check の path。
variable "container_health_check_path" {
  description = "HTTP path used by the ECS container health check."
  type        = string
  default     = "/healthz"

  # HTTP path として / から始まり、空白を含まない値だけを許可する。
  validation {
    condition     = can(regex("^/[^[:space:]]*$", var.container_health_check_path))
    error_message = "Container_health_check_path must start with / and must not contain whitespace."
  }
}

# ECS service が通常起動しておくタスク数。
variable "ecs_desired_count" {
  description = "Desired number of ECS service tasks."
  type        = number
  default     = 1

  # 学習用の dev 環境として 1-10 個だけを許可する。
  validation {
    condition     = var.ecs_desired_count >= 1 && var.ecs_desired_count <= 10
    error_message = "Ecs_desired_count must be between 1 and 10."
  }
}

# ECS rolling deploy 中に維持する最小正常率。
variable "ecs_deployment_minimum_healthy_percent" {
  description = "Minimum healthy percent for ECS rolling deployments."
  type        = number
  default     = 100

  # ECS が受け付ける 0-100 の範囲だけを許可する。
  validation {
    condition     = var.ecs_deployment_minimum_healthy_percent >= 0 && var.ecs_deployment_minimum_healthy_percent <= 100
    error_message = "Ecs_deployment_minimum_healthy_percent must be between 0 and 100."
  }
}

# ECS rolling deploy 中に一時的に増やせる最大率。
variable "ecs_deployment_maximum_percent" {
  description = "Maximum percent for ECS rolling deployments."
  type        = number
  default     = 200

  # ECS が受け付ける 100-200 の範囲だけを許可する。
  validation {
    condition     = var.ecs_deployment_maximum_percent >= 100 && var.ecs_deployment_maximum_percent <= 200
    error_message = "Ecs_deployment_maximum_percent must be between 100 and 200."
  }
}

# Fargate task definition に割り当てる CPU unit。
variable "ecs_task_cpu" {
  description = "CPU units for the ECS Fargate task definition."
  type        = number
  default     = 256

  # Fargate でよく使う CPU unit の候補だけを許可する。
  validation {
    condition     = contains([256, 512, 1024, 2048, 4096], var.ecs_task_cpu)
    error_message = "Ecs_task_cpu must be one of 256, 512, 1024, 2048, or 4096."
  }
}

# Fargate task definition に割り当てる memory MiB。
variable "ecs_task_memory" {
  description = "Memory in MiB for the ECS Fargate task definition."
  type        = number
  default     = 512

  # この学習環境では 512 MiB 以上 30 GiB 以下だけを許可する。
  validation {
    condition     = var.ecs_task_memory >= 512 && var.ecs_task_memory <= 30720
    error_message = "Ecs_task_memory must be between 512 and 30720."
  }
}

# ECR lifecycle policy で untagged image を削除するまでの日数。
variable "ecr_lifecycle_untagged_days" {
  description = "Days to keep untagged ECR images before lifecycle expiration."
  type        = number
  default     = 7

  # ECR lifecycle policy の days として 1日以上を許可する。
  validation {
    condition     = var.ecr_lifecycle_untagged_days >= 1
    error_message = "Ecr_lifecycle_untagged_days must be 1 or greater."
  }
}

# アプリケーションリソースに追加で付与する共通タグ。
variable "tags" {
  description = "Common tags applied to application resources."
  type        = map(string)
  default     = {}

  # すべてのタグで、空白を除いたキーと値が空ではなく、キーが aws: で始まらないことを求める。
  validation {
    condition = alltrue([
      for key, value in var.tags :
      length(trimspace(key)) > 0 &&   # キーが１文字以上であること
      length(trimspace(value)) > 0 && # 値が１文字以上であること
      !startswith(lower(key), "aws:") # キーが aws: で始まらないこと
    ])
    error_message = "Tags must have non-empty keys and values, and tag keys must not start with aws:."
  }
}

# GitHub Actions OIDC trust を許可する repository identifier。
variable "github_repository" {
  description = "GitHub repository identifier allowed to assume OIDC roles, in owner/name form. GitHub may include stable numeric IDs as owner@id/name@id in OIDC sub claims."
  type        = string
  default     = "yabeyuji-row@316766040/study_iac_aws@1333772562"

  validation {
    condition     = can(regex("^[A-Za-z0-9_.-]+(@[0-9]+)?/[A-Za-z0-9_.-]+(@[0-9]+)?$", var.github_repository))
    error_message = "Github_repository must be in owner/name form, with optional numeric IDs as owner@id/name@id."
  }
}

# Terraform plan workflow を許可する default branch。
variable "github_default_branch" {
  description = "Default GitHub branch allowed for branch-scoped OIDC subjects."
  type        = string
  default     = "main"

  validation {
    condition     = can(regex("^[A-Za-z0-9_./-]+$", var.github_default_branch))
    error_message = "Github_default_branch must be a valid branch-like name."
  }
}

# 手動 deploy workflow に使う GitHub environment 名。
variable "github_deploy_environment" {
  description = "GitHub environment name allowed to assume the deploy OIDC role."
  type        = string
  default     = "dev"

  validation {
    condition     = can(regex("^[A-Za-z0-9_.-]+$", var.github_deploy_environment))
    error_message = "Github_deploy_environment must contain only letters, numbers, dots, underscores, or hyphens."
  }
}
