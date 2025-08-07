package repository

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/vsouzx/go-lambda-s3-event-trigger/dto"
)

type Repository struct {
	dynamoClient *dynamodb.Client
}

func NewRepository(dynamoClient *dynamodb.Client) *Repository {
	return &Repository{
		dynamoClient: dynamoClient,
	}
}

func (es *Repository) BatchInsert(ctx context.Context, tableName string, batch []dto.Acesso, workerId int) error {
	fmt.Println("Iniciando batch insert")
	if len(batch) == 0 {
		return nil
	}

	writeReq := make([]types.WriteRequest, 0, len(batch))
	for _, item := range batch {
		av, err := attributevalue.MarshalMap(item)
		if err != nil {
			return fmt.Errorf("erro ao converter item: %w", err)
		}
		writeReq = append(writeReq, types.WriteRequest{
			PutRequest: &types.PutRequest{Item: av},
		})
	}

	request := map[string][]types.WriteRequest{
		tableName: writeReq,
	}

	retries := 0
	for len(request) > 0 && retries < 10 {
		fmt.Println("Realizando batch insert")
		resp, err := es.dynamoClient.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: request,
		})
		if err != nil {
			return fmt.Errorf("erro no BatchWriteItem: %w", err)
		}

		if len(resp.UnprocessedItems) == 0 {
			return nil
		}

		// Se ainda restam itens não processados, reenvia apenas eles
		request = resp.UnprocessedItems
		retries++

		// Backoff exponencial
		time.Sleep(time.Duration((1<<retries)*100+rand.Intn(300)) * time.Millisecond)
	}

	if len(request) > 0 {
		return fmt.Errorf("[Worker %d] %d itens não processados após %d tentativas", workerId, len(request[tableName]), retries)
	}
	return nil
}
