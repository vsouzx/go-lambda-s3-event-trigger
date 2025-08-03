package configs

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func NewDynamoClient() *dynamodb.Client {
	awsConfig := GetAWSConfig()
	return dynamodb.NewFromConfig(*awsConfig)
}
