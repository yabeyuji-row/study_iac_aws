# 手動編集: 可 - IAM ロールとポリシーをここに定義する。

# ECS タスクが引き受ける IAM role の信頼ポリシー。
# ecs-tasks.amazonaws.com だけに AssumeRole を許可する。
data "aws_iam_policy_document" "ecs_tasks_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

# ECS agent が ECR pull や CloudWatch Logs 書き込みに使う execution role。
resource "aws_iam_role" "ecs_task_execution" {
  name               = "${local.name_prefix}-ecs-exec"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks_assume_role.json

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-ecs-exec"
      Role = "ecs-task-execution"
    }
  )
}

# ECR pull と CloudWatch Logs 書き込みに必要な AWS 管理ポリシーだけを付ける。
resource "aws_iam_role_policy_attachment" "ecs_task_execution" {
  role       = aws_iam_role.ecs_task_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

data "aws_iam_policy_document" "database_secret_read" {
  statement {
    actions = [
      "secretsmanager:DescribeSecret",
      "secretsmanager:GetSecretValue",
    ]
    resources = [aws_secretsmanager_secret.db.arn]
  }
}

resource "aws_iam_policy" "database_secret_read" {
  name        = "${local.name_prefix}-db-secret-read"
  description = "Allow reading the TODO API database secret only."
  policy      = data.aws_iam_policy_document.database_secret_read.json

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-db-secret-read"
      Role = "database-secret-read"
    }
  )
}

# ECS agent reads the secret when injecting it into the container environment.
resource "aws_iam_role_policy_attachment" "ecs_task_execution_database_secret" {
  role       = aws_iam_role.ecs_task_execution.name
  policy_arn = aws_iam_policy.database_secret_read.arn
}

# アプリケーションコンテナが AWS API を呼ぶときに使う task role。
# 7週目時点では追加権限を付けず、後続タスクで必要になった権限だけを追加する。
resource "aws_iam_role" "ecs_task" {
  name               = "${local.name_prefix}-ecs-task"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks_assume_role.json

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-ecs-task"
      Role = "ecs-task"
    }
  )
}

# The application task role is also scoped to the same secret for future direct SDK reads.
resource "aws_iam_role_policy_attachment" "ecs_task_database_secret" {
  role       = aws_iam_role.ecs_task.name
  policy_arn = aws_iam_policy.database_secret_read.arn
}
