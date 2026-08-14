# 手動編集: 可 - CloudWatch metrics、alarms、dashboard をここに定義する。

locals {
  observability_namespace = "${var.project_name}/${var.environment}/api"
  alarm_name_prefix       = "${local.name_prefix}-alarm"
}

resource "aws_cloudwatch_log_metric_filter" "api_request_count" {
  name           = "${local.name_prefix}-api-request-count"
  log_group_name = aws_cloudwatch_log_group.api.name
  pattern        = "{ $.event = \"http_request_completed\" }"

  metric_transformation {
    name          = "RequestCount"
    namespace     = local.observability_namespace
    value         = "1"
    default_value = "0"
    unit          = "Count"
  }
}

resource "aws_cloudwatch_log_metric_filter" "api_error_count" {
  name           = "${local.name_prefix}-api-error-count"
  log_group_name = aws_cloudwatch_log_group.api.name
  pattern        = "{ $.event = \"http_request_completed\" && $.status >= 500 }"

  metric_transformation {
    name          = "ErrorCount"
    namespace     = local.observability_namespace
    value         = "1"
    default_value = "0"
    unit          = "Count"
  }
}

resource "aws_cloudwatch_log_metric_filter" "api_latency_ms" {
  name           = "${local.name_prefix}-api-latency-ms"
  log_group_name = aws_cloudwatch_log_group.api.name
  pattern        = "{ $.event = \"http_request_completed\" && $.duration_ms = * }"

  metric_transformation {
    name      = "LatencyMs"
    namespace = local.observability_namespace
    value     = "$.duration_ms"
    unit      = "Milliseconds"
  }
}

resource "aws_cloudwatch_log_metric_filter" "api_db_readiness_failure" {
  name           = "${local.name_prefix}-api-db-readiness-failure"
  log_group_name = aws_cloudwatch_log_group.api.name
  pattern        = "{ $.event = \"http_request_completed\" && $.path = \"/readyz\" && $.status = 503 }"

  metric_transformation {
    name          = "DBReadinessFailure"
    namespace     = local.observability_namespace
    value         = "1"
    default_value = "0"
    unit          = "Count"
  }
}

resource "aws_cloudwatch_metric_alarm" "alb_target_5xx" {
  alarm_name          = "${local.alarm_name_prefix}-alb-target-5xx"
  alarm_description   = "ALB target 5xx responses indicate API server-side failures."
  namespace           = "AWS/ApplicationELB"
  metric_name         = "HTTPCode_Target_5XX_Count"
  statistic           = "Sum"
  period              = 60
  evaluation_periods  = 5
  datapoints_to_alarm = 3
  threshold           = 5
  comparison_operator = "GreaterThanOrEqualToThreshold"
  treat_missing_data  = "notBreaching"

  dimensions = {
    LoadBalancer = aws_lb.app.arn_suffix
    TargetGroup  = aws_lb_target_group.api.arn_suffix
  }

  tags = merge(
    local.common_tags,
    {
      Name = "${local.alarm_name_prefix}-alb-target-5xx"
      Role = "alarm"
    }
  )
}

resource "aws_cloudwatch_metric_alarm" "alb_unhealthy_targets" {
  alarm_name          = "${local.alarm_name_prefix}-alb-unhealthy-targets"
  alarm_description   = "ALB has unhealthy API targets."
  namespace           = "AWS/ApplicationELB"
  metric_name         = "UnHealthyHostCount"
  statistic           = "Maximum"
  period              = 60
  evaluation_periods  = 3
  datapoints_to_alarm = 2
  threshold           = 1
  comparison_operator = "GreaterThanOrEqualToThreshold"
  treat_missing_data  = "notBreaching"

  dimensions = {
    LoadBalancer = aws_lb.app.arn_suffix
    TargetGroup  = aws_lb_target_group.api.arn_suffix
  }

  tags = merge(
    local.common_tags,
    {
      Name = "${local.alarm_name_prefix}-alb-unhealthy-targets"
      Role = "alarm"
    }
  )
}

resource "aws_cloudwatch_metric_alarm" "ecs_running_task_count_low" {
  alarm_name          = "${local.alarm_name_prefix}-ecs-running-task-count-low"
  alarm_description   = "ECS running task count is below desired count."
  namespace           = "ECS/ContainerInsights"
  metric_name         = "RunningTaskCount"
  statistic           = "Minimum"
  period              = 60
  evaluation_periods  = 3
  datapoints_to_alarm = 2
  threshold           = var.ecs_desired_count
  comparison_operator = "LessThanThreshold"
  treat_missing_data  = "breaching"

  dimensions = {
    ClusterName = aws_ecs_cluster.api.name
    ServiceName = aws_ecs_service.api.name
  }

  tags = merge(
    local.common_tags,
    {
      Name = "${local.alarm_name_prefix}-ecs-running-task-count-low"
      Role = "alarm"
    }
  )
}

resource "aws_cloudwatch_metric_alarm" "rds_cpu_high" {
  alarm_name          = "${local.alarm_name_prefix}-rds-cpu-high"
  alarm_description   = "RDS CPU utilization is high."
  namespace           = "AWS/RDS"
  metric_name         = "CPUUtilization"
  statistic           = "Average"
  period              = 300
  evaluation_periods  = 3
  datapoints_to_alarm = 2
  threshold           = 80
  comparison_operator = "GreaterThanOrEqualToThreshold"
  treat_missing_data  = "missing"

  dimensions = {
    DBInstanceIdentifier = aws_db_instance.app.identifier
  }

  tags = merge(
    local.common_tags,
    {
      Name = "${local.alarm_name_prefix}-rds-cpu-high"
      Role = "alarm"
    }
  )
}

resource "aws_cloudwatch_metric_alarm" "rds_free_storage_low" {
  alarm_name          = "${local.alarm_name_prefix}-rds-free-storage-low"
  alarm_description   = "RDS free storage is below 2 GiB."
  namespace           = "AWS/RDS"
  metric_name         = "FreeStorageSpace"
  statistic           = "Average"
  period              = 300
  evaluation_periods  = 3
  datapoints_to_alarm = 2
  threshold           = 2147483648
  comparison_operator = "LessThanThreshold"
  treat_missing_data  = "missing"

  dimensions = {
    DBInstanceIdentifier = aws_db_instance.app.identifier
  }

  tags = merge(
    local.common_tags,
    {
      Name = "${local.alarm_name_prefix}-rds-free-storage-low"
      Role = "alarm"
    }
  )
}

resource "aws_cloudwatch_metric_alarm" "rds_database_connections_high" {
  alarm_name          = "${local.alarm_name_prefix}-rds-db-connections-high"
  alarm_description   = "RDS database connections are close to the application pool budget."
  namespace           = "AWS/RDS"
  metric_name         = "DatabaseConnections"
  statistic           = "Average"
  period              = 300
  evaluation_periods  = 3
  datapoints_to_alarm = 2
  threshold           = 40
  comparison_operator = "GreaterThanOrEqualToThreshold"
  treat_missing_data  = "missing"

  dimensions = {
    DBInstanceIdentifier = aws_db_instance.app.identifier
  }

  tags = merge(
    local.common_tags,
    {
      Name = "${local.alarm_name_prefix}-rds-db-connections-high"
      Role = "alarm"
    }
  )
}

resource "aws_cloudwatch_metric_alarm" "api_db_readiness_failure" {
  alarm_name          = "${local.alarm_name_prefix}-api-db-readiness-failure"
  alarm_description   = "API readiness endpoint is returning DB failure."
  namespace           = local.observability_namespace
  metric_name         = "DBReadinessFailure"
  statistic           = "Sum"
  period              = 60
  evaluation_periods  = 3
  datapoints_to_alarm = 2
  threshold           = 1
  comparison_operator = "GreaterThanOrEqualToThreshold"
  treat_missing_data  = "notBreaching"

  depends_on = [aws_cloudwatch_log_metric_filter.api_db_readiness_failure]

  tags = merge(
    local.common_tags,
    {
      Name = "${local.alarm_name_prefix}-api-db-readiness-failure"
      Role = "alarm"
    }
  )
}

resource "aws_cloudwatch_dashboard" "operations" {
  dashboard_name = "${local.name_prefix}-operations"

  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 12
        height = 6
        properties = {
          title   = "ALB traffic and errors"
          region  = var.aws_region
          view    = "timeSeries"
          stacked = false
          metrics = [
            ["AWS/ApplicationELB", "RequestCount", "LoadBalancer", aws_lb.app.arn_suffix, { stat = "Sum" }],
            [".", "HTTPCode_Target_5XX_Count", ".", ".", "TargetGroup", aws_lb_target_group.api.arn_suffix, { stat = "Sum" }],
            [".", "TargetResponseTime", ".", ".", ".", ".", { stat = "Average" }],
          ]
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 0
        width  = 12
        height = 6
        properties = {
          title   = "Target health and ECS tasks"
          region  = var.aws_region
          view    = "timeSeries"
          stacked = false
          metrics = [
            ["AWS/ApplicationELB", "HealthyHostCount", "LoadBalancer", aws_lb.app.arn_suffix, "TargetGroup", aws_lb_target_group.api.arn_suffix, { stat = "Minimum" }],
            [".", "UnHealthyHostCount", ".", ".", ".", ".", { stat = "Maximum" }],
            ["ECS/ContainerInsights", "RunningTaskCount", "ClusterName", aws_ecs_cluster.api.name, "ServiceName", aws_ecs_service.api.name, { stat = "Minimum" }],
          ]
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 6
        width  = 12
        height = 6
        properties = {
          title   = "API log-derived metrics"
          region  = var.aws_region
          view    = "timeSeries"
          stacked = false
          metrics = [
            [local.observability_namespace, "RequestCount", { stat = "Sum" }],
            [".", "ErrorCount", { stat = "Sum" }],
            [".", "DBReadinessFailure", { stat = "Sum" }],
            [".", "LatencyMs", { stat = "Average", yAxis = "right" }],
          ]
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 6
        width  = 12
        height = 6
        properties = {
          title   = "RDS health"
          region  = var.aws_region
          view    = "timeSeries"
          stacked = false
          metrics = [
            ["AWS/RDS", "CPUUtilization", "DBInstanceIdentifier", aws_db_instance.app.identifier, { stat = "Average" }],
            [".", "DatabaseConnections", ".", ".", { stat = "Average" }],
            [".", "FreeStorageSpace", ".", ".", { stat = "Average", yAxis = "right" }],
          ]
        }
      },
    ]
  })
}
