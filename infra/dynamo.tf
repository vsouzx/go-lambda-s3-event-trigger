resource "aws_dynamodb_table" "dynamodb-table" {
  name           = var.dynamo_table
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "funcionalChefe"    # Partition Key
  range_key      = "funcionalColaborador" # Sort Key

  attribute {
    name = "funcionalChefe"
    type = "S"
  }

  attribute {
    name = "funcionalColaborador"
    type = "S"
  }

  tags = {
    Name        = var.dynamo_table
    Environment = var.stage_name
  }
}