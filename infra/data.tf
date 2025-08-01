data "aws_iam_policy_document" "assume_role"{
    statement{
        effect = "Allow"

        principals{
            type        = "Service"
            identifiers = ["lambda.amazonaws.com"]
        }

        actions = ["sts:AssumeRole"]
    }
}