resource "aws_api_gateway_rest_api" "bucket_s3_gtw_api" {
    name = "bucket_s3_upload_api"
    description = "REST API for Bucket S3 files upload"

    binary_media_types = [
      "/*"
    ]

    endpoint_configuration {
      types = ["REGIONAL"]
    }
}

# =========================
# IAM ROLE PARA API GATEWAY ACESSAR S3
# =========================
resource "aws_iam_role" "apigw_s3_role" {
  name = "apigw_s3_upload_role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Principal = {
          Service = "apigateway.amazonaws.com"
        },
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role_policy" "apigw_s3_policy" {
  role = aws_iam_role.apigw_s3_role.id

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect   = "Allow",
        Action   = ["s3:PutObject", "s3:GetObject", "s3:DeleteObject", "s3:CreateBucket"],
        Resource = "*"
      }
    ]
  })
}

resource "aws_api_gateway_resource" "excel_gw_api_resource" {
    parent_id       = aws_api_gateway_rest_api.bucket_s3_gtw_api.root_resource_id
    path_part       = "upload"
    rest_api_id     = aws_api_gateway_rest_api.bucket_s3_gtw_api.id
}

# Novo recurso filho para capturar o {object}
resource "aws_api_gateway_resource" "excel_gw_api_resource_object" {
  parent_id   = aws_api_gateway_resource.excel_gw_api_resource.id
  path_part   = "{object}"
  rest_api_id = aws_api_gateway_rest_api.bucket_s3_gtw_api.id
}

//POST
resource "aws_api_gateway_method" "excel_gw_api_method_post" {
  authorization   = "NONE"
  http_method     = "POST"
  resource_id     = aws_api_gateway_resource.excel_gw_api_resource_object.id
  rest_api_id     = aws_api_gateway_rest_api.bucket_s3_gtw_api.id

      # Habilitar parâmetro de path
  request_parameters = {
    "method.request.path.object" = true
  }
}

resource "aws_api_gateway_integration" "excel_s3_integration_post" {
  rest_api_id             = aws_api_gateway_rest_api.bucket_s3_gtw_api.id
  resource_id             = aws_api_gateway_resource.excel_gw_api_resource_object.id
  http_method             = aws_api_gateway_method.excel_gw_api_method_post.http_method
  type                    = "AWS"
  integration_http_method = "PUT" # Método usado para envio de arquivo
  uri                     = "arn:aws:apigateway:${var.aws_region}:s3:path/${aws_s3_bucket.excel_bucket.bucket}/{object}"
  credentials             = aws_iam_role.apigw_s3_role.arn

  request_parameters = {
    "integration.request.path.object" = "method.request.path.object"
  }
}

resource "aws_api_gateway_method_response" "excel_response_200_post" {
  http_method = aws_api_gateway_method.excel_gw_api_method_post.http_method
  resource_id = aws_api_gateway_resource.excel_gw_api_resource_object.id
  rest_api_id = aws_api_gateway_rest_api.bucket_s3_gtw_api.id
  status_code = "200"
}

resource "aws_api_gateway_deployment" "api_deployment" {
    rest_api_id = aws_api_gateway_rest_api.bucket_s3_gtw_api.id

    triggers = {
      redeployment = timestamp()
    }

    lifecycle {
      create_before_destroy = true
    }

    depends_on = [ 
         aws_api_gateway_integration.excel_s3_integration_post,
     ]
}

resource "aws_api_gateway_stage" "api_stage" {
  deployment_id = aws_api_gateway_deployment.api_deployment.id
  rest_api_id   = aws_api_gateway_rest_api.bucket_s3_gtw_api.id
  stage_name    = var.stage_name
}