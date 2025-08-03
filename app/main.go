package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/vsouzx/go-lambda-s3-event-trigger/configs"
	"github.com/vsouzx/go-lambda-s3-event-trigger/service"
)

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, s3Event events.S3Event) error {
	s3Service := service.NewS3Service(configs.NewS3Client())
	excelService := service.NewExcelService(configs.NewDynamoClient())

	for _, record := range s3Event.Records {
		bucketName := record.S3.Bucket.Name
		objectKey := record.S3.Object.Key
		println("Bucket:", bucketName, "Object Key:", objectKey)

		fileBytes, err := s3Service.GetS3FileBytes(ctx, bucketName, objectKey)
		if err != nil {
			return err
		}

		excelService.ConvertExcelToCSV(fileBytes, "/tmp/output.csv")

		excelService.ProcessCSVFile("/tmp/output.csv")
	}
	return nil
}
