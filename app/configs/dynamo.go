package configs

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoConfig struct {
	DynamoClient *dynamodb.Client
}

func NewDynamoConfig(dynamoClient *dynamodb.Client) *DynamoConfig {
	return &DynamoConfig{
		DynamoClient: dynamoClient,
	}
}
