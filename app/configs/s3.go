package configs

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Config struct {
	S3Client *s3.Client
}

func NewS3Config(s3Client *s3.Client) *S3Config {
	return &S3Config{
		S3Client: s3Client,
	}
}
