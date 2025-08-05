resource "aws_iam_role" "iam_for_lambda"{
    name                    = "iam_for_lambda"
    assume_role_policy      = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_policy" "lambda_logging" {
    name = "go_lambda_logging"
    path = "/"
    description = "IAM policy for logging from a lambda"

    policy = <<EOF
    {
        "Version":"2012-10-17",
        "Statement": [
            {
                "Action": [
                    "logs:CreateLogGroup",
                    "logs:CreateLogStream",
                    "logs:PutLogEvents"
                ],
                "Resource": "arn:aws:logs:*:*:*",
                "Effect": "Allow"
            },
            {
                "Action": [
                    "ec2:CreateNetworkInterface",
                    "ec2:DescribeNetworkInterfaces",
                    "ec2:DeleteNetworkInterface"
                ],
                "Resource": "*",
                "Effect": "Allow"
            },
            {
                "Action": [
                    "dynamodb:PutItem",
                    "dynamodb:BatchWriteItem"
                ],
                "Resource": "*",
                "Effect": "Allow"
            },
            {
                "Action": [
                    "s3:GetObject"
                ],
                "Resource": "*",
                "Effect": "Allow"
            }
        ]
    }
    EOF
}

resource "aws_iam_role_policy_attachment" "lambda_logs" {
    role = aws_iam_role.iam_for_lambda.name
    policy_arn = aws_iam_policy.lambda_logging.arn
}

resource "aws_lambda_function" "lambda" {
  function_name    = "s3_event_handler"
  role             = aws_iam_role.iam_for_lambda.arn
  runtime          = "provided.al2"
  handler          = "bootstrap"
  filename         = "lambda.zip"
  source_code_hash = filebase64sha256("${path.module}/lambda.zip")
  memory_size      = 1024
  timeout          = 900

  environment {
    variables = {
      DYNAMO_TABLE = var.dynamo_table
      WORKERS = var.workers
      BATCH_SIZE = var.batch_size
      BUFFER_SIZE = var.buffer_size
    }
  }
}

resource "aws_cloudwatch_log_group" "example"{
    name = "/aws/lambda/${aws_lambda_function.lambda.function_name}"
    retention_in_days = var.log_retention_days
}