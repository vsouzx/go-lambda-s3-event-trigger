output "invoke_url" {
    value = aws_api_gateway_deployment.api_deployment.invoke_url
}

output "rds_endpoint" {
    value = aws_db_instance.default.endpoint
}