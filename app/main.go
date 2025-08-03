package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/vsouzx/go-lambda-s3-event-trigger/configs"
)

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, s3Event events.S3Event) error {
	dynamoClient := configs.NewDynamoClient()
	tableName := "excel-import"

	for _, record := range s3Event.Records {
		bucketName := record.S3.Bucket.Name
		objectKey := record.S3.Object.Key
		println("Bucket:", bucketName, "Object Key:", objectKey)

		item := map[string]string{
			"funcionalChefe":       record.S3.Bucket.Name,
			"funcionalColaborador": record.S3.Object.Key,
			"departamento":         string(time.Now().Format("2006-01-02 15:04:05")),
		}

		av, err := attributevalue.MarshalMap(item)
		if err != nil {
			log.Fatalf("❌ Erro ao converter item: %v", err)
		}

		_, err = dynamoClient.PutItem(context.Background(), &dynamodb.PutItemInput{
			TableName: &tableName,
			Item:      av,
		})

		if err != nil {
			log.Fatalf("❌ Erro ao inserir item: %v", err)
		} else {
			fmt.Println("✅ Item inserido com sucesso no DynamoDB!")
		}
	}
	return nil
}
