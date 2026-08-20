# 手動編集: 可 - GitHub Actions OIDC role と policy をここに定義する。

locals {
  github_oidc_provider_url = "https://token.actions.githubusercontent.com"
  github_oidc_audience     = "sts.amazonaws.com"

  github_plan_subjects = [
    "repo:${var.github_repository}:ref:refs/heads/${var.github_default_branch}",
    "repo:${var.github_repository}:pull_request",
  ]

  github_deploy_subject = "repo:${var.github_repository}:environment:${var.github_deploy_environment}"
}

resource "aws_iam_openid_connect_provider" "github_actions" {
  url = local.github_oidc_provider_url

  client_id_list = [
    local.github_oidc_audience,
  ]

  thumbprint_list = [
    "6938fd4d98bab03faadb97b34396831e3780aea1",
    "1c58a3a8518e8759bf075b76b750d4f2df264fcd",
  ]

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-github-oidc"
      Role = "github-oidc"
    }
  )
}

data "aws_iam_policy_document" "github_plan_assume_role" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [aws_iam_openid_connect_provider.github_actions.arn]
    }

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = [local.github_oidc_audience]
    }

    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      values   = local.github_plan_subjects
    }
  }
}

data "aws_iam_policy_document" "github_deploy_assume_role" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [aws_iam_openid_connect_provider.github_actions.arn]
    }

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = [local.github_oidc_audience]
    }

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:sub"
      values   = [local.github_deploy_subject]
    }
  }
}

resource "aws_iam_role" "github_plan" {
  name               = "${local.name_prefix}-github-plan"
  assume_role_policy = data.aws_iam_policy_document.github_plan_assume_role.json

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-github-plan"
      Role = "github-plan"
    }
  )
}

resource "aws_iam_role" "github_deploy" {
  name               = "${local.name_prefix}-github-deploy"
  assume_role_policy = data.aws_iam_policy_document.github_deploy_assume_role.json

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-github-deploy"
      Role = "github-deploy"
    }
  )
}

data "aws_iam_policy_document" "github_terraform_plan" {
  #checkov:skip=CKV_AWS_356: Terraform plan needs broad describe/list access to discover existing AWS resources without write permissions.
  statement {
    actions = [
      "ec2:Describe*",
      "ecr:Describe*",
      "ecr:List*",
      "ecs:Describe*",
      "ecs:List*",
      "elasticloadbalancing:Describe*",
      "iam:Get*",
      "iam:List*",
      "logs:Describe*",
      "rds:Describe*",
      "secretsmanager:DescribeSecret",
      "sts:GetCallerIdentity",
    ]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "github_terraform_plan" {
  name        = "${local.name_prefix}-github-terraform-plan"
  description = "Read-only permissions for GitHub Actions Terraform plan."
  policy      = data.aws_iam_policy_document.github_terraform_plan.json

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-github-terraform-plan"
      Role = "github-terraform-plan"
    }
  )
}

resource "aws_iam_role_policy_attachment" "github_plan" {
  role       = aws_iam_role.github_plan.name
  policy_arn = aws_iam_policy.github_terraform_plan.arn
}

data "aws_iam_policy_document" "github_ecr_push" {
  #checkov:skip=CKV_AWS_356: ecr:GetAuthorizationToken only supports Resource "*"; repository writes are scoped below.
  statement {
    actions = [
      "ecr:GetAuthorizationToken",
    ]
    resources = ["*"]
  }

  statement {
    actions = [
      "ecr:BatchCheckLayerAvailability",
      "ecr:CompleteLayerUpload",
      "ecr:DescribeImages",
      "ecr:DescribeRepositories",
      "ecr:InitiateLayerUpload",
      "ecr:PutImage",
      "ecr:UploadLayerPart",
    ]
    resources = [aws_ecr_repository.api.arn]
  }
}

resource "aws_iam_policy" "github_ecr_push" {
  name        = "${local.name_prefix}-github-ecr-push"
  description = "Allow GitHub Actions to push API images to the scoped ECR repository."
  policy      = data.aws_iam_policy_document.github_ecr_push.json

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-github-ecr-push"
      Role = "github-ecr-push"
    }
  )
}

resource "aws_iam_role_policy_attachment" "github_deploy_ecr_push" {
  role       = aws_iam_role.github_deploy.name
  policy_arn = aws_iam_policy.github_ecr_push.arn
}

data "aws_iam_policy_document" "github_ecs_deploy" {
  statement {
    actions = [
      "ecs:DescribeServices",
      "ecs:UpdateService",
    ]
    resources = [aws_ecs_service.api.id]
  }

  #checkov:skip=CKV_AWS_356: GitHub deploy runs one-off migration tasks against the freshly registered task definition revision.
  statement {
    actions = [
      "ecs:DescribeTasks",
      "ecs:RunTask",
    ]
    resources = ["*"]
  }

  #checkov:skip=CKV_AWS_356: ecs:RegisterTaskDefinition and ecs:DescribeTaskDefinition require Resource "*" in IAM.
  statement {
    actions = [
      "ecs:DescribeTaskDefinition",
      "ecs:RegisterTaskDefinition",
    ]
    resources = ["*"]
  }

  statement {
    actions = [
      "iam:PassRole",
    ]
    resources = [
      aws_iam_role.ecs_task_execution.arn,
      aws_iam_role.ecs_task.arn,
    ]

    condition {
      test     = "StringEquals"
      variable = "iam:PassedToService"
      values   = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_policy" "github_ecs_deploy" {
  name        = "${local.name_prefix}-github-ecs-deploy"
  description = "Allow GitHub Actions to register task definitions and update the scoped ECS service."
  policy      = data.aws_iam_policy_document.github_ecs_deploy.json

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-github-ecs-deploy"
      Role = "github-ecs-deploy"
    }
  )
}

resource "aws_iam_role_policy_attachment" "github_deploy_ecs" {
  role       = aws_iam_role.github_deploy.name
  policy_arn = aws_iam_policy.github_ecs_deploy.arn
}
