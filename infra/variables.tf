variable "aws_region" {
    type = string
}

variable "stage_name" {
    type = string
}

variable "log_retention_days" {
    type = number
}

variable "dynamo_table" {
  type = string
}

variable "bucket_name" {
  type = string
}