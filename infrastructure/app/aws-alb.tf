# 手動編集: 可 - public ALB と ECS 向け target group をここに定義する。

# インターネットから HTTP を受ける Application Load Balancer。
# ALB は public subnet に置き、security group で HTTP ingress を管理する。
resource "aws_lb" "app" {
  #values() は Terraform の関数で、map/object の値だけを list として取り出す
  #checkov:skip=CKV_AWS_91: access log は S3 bucket、retention、cost 設計後に扱う。
  #checkov:skip=CKV_AWS_150: 開発環境では削除しやすさを優先し、削除保護は後続で環境別に検討する。
  #checkov:skip=CKV2_AWS_20: HTTPS redirect は ACM 証明書と domain 設計の週で追加する。
  #checkov:skip=CKV2_AWS_28: WAF は public entrypoint の防御設計として後続の security フェーズで扱う。
  name                       = "${local.name_prefix}-alb"
  load_balancer_type         = "application"
  internal                   = false                           # 内部向けロードバランサーか
  security_groups            = [aws_security_group.alb.id]     # 関連付けるセキュリティグループ
  subnets                    = values(aws_subnet.public)[*].id # 作成した public subnet すべての ID を list にする
  drop_invalid_header_fields = true                            # 不正な HTTP ヘッダーの破棄

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-alb"
      Role = "public-entrypoint"
    }
  )
}

# ALB から ECS Fargate task へ HTTP を転送する target group。
# Fargate の awsvpc mode では target_type に ip を使う。
resource "aws_lb_target_group" "api" {
  #checkov:skip=CKV_AWS_378: 8週目は HTTP の学習構成とし、HTTPS 化は証明書設計とあわせて後続で扱う。
  name        = "${local.name_prefix}-api-tg"
  port        = var.app_port
  protocol    = "HTTP"
  target_type = "ip"
  vpc_id      = aws_vpc.main.id

  # ALB は DB 接続を含む readiness を確認し、受け付け可能な task だけへ流す。
  health_check {
    enabled             = true
    path                = var.health_check_path
    matcher             = "200"
    interval            = 30
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 2
  }

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-api-tg"
      Role = "api-targets"
    }
  )
}

# ALB の HTTP listener。
# 受けたリクエストを API target group へ転送する。
resource "aws_lb_listener" "http" {
  #checkov:skip=CKV_AWS_2: HTTPS listener は ACM 証明書と domain 設計の週で追加する。
  #checkov:skip=CKV_AWS_103: 8週目は HTTP listener の基本動作を plan で確認する。
  load_balancer_arn = aws_lb.app.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.api.arn
  }
}
