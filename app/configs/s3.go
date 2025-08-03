package configs

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewS3Client() *s3.Client {
	awsConfig := GetAWSConfig()
	return s3.NewFromConfig(*awsConfig)
}
