# 手動編集: 可 - Terraform test を意図して変更する。
mock_provider "aws" {}

run "valid_state_bucket_name" {
  command = plan

  variables {
    state_bucket_name = "study-aws-tfstate-example"
    tags = {
      Project = "study-aws"
    }
  }

  assert {
    condition     = aws_s3_bucket.terraform_state.bucket == "study-aws-tfstate-example"
    error_message = "The Terraform state bucket name should match state_bucket_name."
  }
}

run "invalid_state_bucket_name" {
  command = plan

  variables {
    state_bucket_name = "INVALID_BUCKET_NAME"
  }

  expect_failures = [
    var.state_bucket_name,
  ]
}
