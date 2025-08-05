package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/vsouzx/go-lambda-s3-event-trigger/configs"
	"github.com/vsouzx/go-lambda-s3-event-trigger/repository"
	"github.com/vsouzx/go-lambda-s3-event-trigger/service"
)

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, s3Event events.S3Event) error {
	start := time.Now()

	dynamoConfig := configs.NewDynamoConfig(dynamodb.NewFromConfig(configs.GetAWSConfig()))
	s3Config := configs.NewS3Config(s3.NewFromConfig(configs.GetAWSConfig()))

	repository := repository.NewRepository(dynamoConfig.DynamoClient)
	s3Service := service.NewS3Service(s3Config.S3Client)
	excelProcessorService := service.NewExcelProcessorService(repository)

	for _, record := range s3Event.Records {
		bucketName := record.S3.Bucket.Name
		objectKey := record.S3.Object.Key
		fmt.Println("Bucket:", bucketName, "Object Key:", objectKey)

		fileBytes, err := s3Service.GetS3FileBytes(ctx, bucketName, objectKey)
		if err != nil {
			return err
		}

		if err := excelProcessorService.ProcessExcelFile(fileBytes); err != nil {
			return fmt.Errorf("erro ao processar registros do excel no dynamo: %w", err)
		}
	}

	duration := time.Since(start)
	fmt.Printf("Lambda concluída em %v\n", duration)
	return nil
}
