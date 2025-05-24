provider "aws" {
  region = "us-east-2"
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.1.1"

  name = "osv-crawler-vpc"
  cidr = "10.0.0.0/16"

  azs              = ["us-east-2a", "us-east-2b"]
  public_subnets   = ["10.0.1.0/24", "10.0.2.0/24"]
  private_subnets  = ["10.0.11.0/24", "10.0.12.0/24"]

  enable_nat_gateway   = true
  single_nat_gateway   = true
  enable_dns_hostnames = true
  enable_dns_support   = true
  create_igw           = true

  tags = {
    Project = "osv-crawler"
  }
}

resource "aws_cloudwatch_log_group" "ecs_logs" {
  name              = "/ecs/osv-crawler"
  retention_in_days = 7
}

resource "aws_ecr_repository" "crawler_repo" {
  name = "osv-crawler"
}

resource "aws_s3_bucket" "crawler_bucket" {
  bucket = "crawler-input-bucket"
  force_destroy = true
}

resource "aws_iam_role" "ecs_task_exec_role" {
  name = "osv-crawler-exec"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{
      Effect = "Allow",
      Principal = { Service = "ecs-tasks.amazonaws.com" },
      Action = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "ecs_task_exec_logs" {
  role       = aws_iam_role.ecs_task_exec_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role_policy" "crawler_s3_access" {
  name = "crawler-s3-access"
  role = aws_iam_role.ecs_task_exec_role.id

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{
      Action = ["s3:GetObject", "s3:PutObject"],
      Effect = "Allow",
      Resource = "${aws_s3_bucket.crawler_bucket.arn}/*"
    }]
  })
}

resource "aws_ecs_cluster" "crawler" {
  name = "osv-crawler-cluster"
}

resource "aws_ecs_task_definition" "crawler_task" {
  family                   = "osv-crawler"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_exec_role.arn
  task_role_arn            = aws_iam_role.ecs_task_exec_role.arn

  container_definitions = jsonencode([{
    name      = "osv-crawler"
    image     = "${aws_ecr_repository.crawler_repo.repository_url}:latest"
    essential = true
    portMappings = [{
      containerPort = 2112,
      protocol      = "tcp"
    }],
    logConfiguration = {
      logDriver = "awslogs",
      options = {
        awslogs-group         = aws_cloudwatch_log_group.ecs_logs.name,
        awslogs-region        = "us-east-2",
        awslogs-stream-prefix = "osv"
      }
    },
    environment = [
      { name = "LOG_FORMAT", value = "json" }
    ]
  }])
}

resource "aws_cloudwatch_event_rule" "on_targets_upload" {
  name = "osv-crawler-trigger"
  event_pattern = jsonencode({
    source = ["aws.s3"],
    "detail-type" = ["AWS API Call via CloudTrail"],
    detail = {
      eventName = ["PutObject"],
      requestParameters = {
        key = [{ "suffix": ".json" }]
      }
    }
  })
}

resource "aws_cloudwatch_event_target" "trigger_task" {
  rule     = aws_cloudwatch_event_rule.on_targets_upload.name
  arn      = aws_ecs_cluster.crawler.arn
  role_arn = aws_iam_role.ecs_task_exec_role.arn

  ecs_target {
    launch_type         = "FARGATE"
    task_definition_arn = aws_ecs_task_definition.crawler_task.arn
    network_configuration {
      subnets          = module.vpc.public_subnets
      security_groups  = [aws_security_group.ecs_sg.id]
      assign_public_ip = true
    }
  }

  input_transformer {
    input_paths = {
      key = "$.detail.requestParameters.key"
    }
    input_template = <<EOF
{
  "containerOverrides": [{
    "name": "osv-crawler",
    "environment": [
      { "name": "TARGETS_S3_URI", "value": "s3://${aws_s3_bucket.crawler_bucket.bucket}/<key>" },
      { "name": "OUTPUT_S3_BUCKET", "value": "${aws_s3_bucket.crawler_bucket.bucket}" },
      { "name": "OUTPUT_S3_PREFIX", "value": "output/" },
      { "name": "TASK_TAG", "value": "<key>" },
      { "name": "LOG_FORMAT", "value": "json" }
    ]
  }]
}
EOF
  }
}

resource "aws_security_group" "ecs_sg" {
  name        = "osv-crawler-sg"
  description = "Allow outbound for ECS"
  vpc_id      = module.vpc.vpc_id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
