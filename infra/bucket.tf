resource "aws_s3_bucket" "excel_bucket" {
  bucket = var.bucket_name
  force_destroy = true

  tags = {
    Name        = "excel-bucket"
    Environment = "prod"
  }
}

resource "aws_s3_bucket_public_access_block" "fotos_colaboradores_block" {
  bucket = aws_s3_bucket.s_facial_recognition_bucket.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_lambda_permission" "allow_s3_to_invoke_lambda" {
  statement_id  = "AllowS3InvokeLambda"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.process_excel.function_name
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.excel_bucket.arn
}

# Configuração de notificação S3 → Lambda
resource "aws_s3_bucket_notification" "excel_bucket_notification" {
  bucket = aws_s3_bucket.excel_bucket.id

  lambda_function {
    lambda_function_arn = aws_lambda_function.lambda.arn
    events              = ["s3:ObjectCreated:*"] # Dispara em qualquer upload
    filter_prefix       = ""                     # opcional: filtrar por prefixo
    filter_suffix       = ".xlsx"                # opcional: apenas arquivos Excel
  }

  depends_on = [aws_lambda_permission.allow_s3_to_invoke_lambda]
}