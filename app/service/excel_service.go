package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/xuri/excelize/v2"
)

const (
	batchSize   = 25 // limite do DynamoDB
	workerCount = 15 // número de goroutines para processar em paralelo
)

type ExcelService struct {
	dynamoClient *dynamodb.Client
}

func NewExcelService(dynamoClient *dynamodb.Client) *ExcelService {
	return &ExcelService{
		dynamoClient: dynamoClient,
	}
}

func (es *ExcelService) ConvertExcelToCSV(excelBytes []byte, outputPath string) error {
	f, err := excelize.OpenReader(bytes.NewReader(excelBytes))
	if err != nil {
		return fmt.Errorf("erro ao abrir excel: %w", err)
	}
	defer f.Close()

	sheet := f.GetSheetList()[0]
	rows, err := f.GetRows(sheet)
	if err != nil {
		return fmt.Errorf("erro ao ler linhas excel: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("erro criando arquivo csv: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("erro escrevendo linha csv: %w", err)
		}
	}
	return nil
}

func (es *ExcelService) ProcessCSVFile(csvPath string) error {
	tableName := "excel-import"

	file, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("erro abrindo csv: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	// Canal para enviar lotes para workers
	batchChan := make(chan []map[string]string, 100)
	var wg sync.WaitGroup

	// Workers que vão processar os lotes
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for batch := range batchChan {
				if err := es.flushBatch(context.Background(), tableName, batch, id); err != nil {
					fmt.Printf("[Worker %d] Erro ao inserir lote: %v\n", id, err)
				}
			}
		}(i)
	}

	// Leitura do CSV e envio para o canal
	line := 0
	var batch []map[string]string

	for {
		record, err := reader.Read()
		fmt.Println("Processando linha:", line)
		if err == io.EOF {
			fmt.Println("Fim do arquivo CSV")
			break
		}
		if err != nil {
			return fmt.Errorf("erro lendo csv: %w", err)
		}

		line++
		if line == 1 {
			continue
		}
		if len(record) < 3 {
			continue
		}

		batch = append(batch, map[string]string{
			"funcionalChefe":       record[0],
			"funcionalColaborador": record[1],
			"departamento":         record[2],
		})

		if len(batch) == batchSize {
			b := make([]map[string]string, len(batch))
			copy(b, batch)
			batchChan <- b
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		batchChan <- batch
	}

	close(batchChan)
	wg.Wait()

	fmt.Printf("Finalizado processamento do CSV com %d linhas\n", line)
	return nil
}

func (es *ExcelService) flushBatch(ctx context.Context, tableName string, batch []map[string]string, workerId int) error {
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
		resp, err := es.dynamoClient.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: request,
		})
		if err != nil {
			return fmt.Errorf("erro no BatchWriteItem: %w", err)
		}

		if len(resp.UnprocessedItems) == 0 {
			fmt.Printf("[Worker %d] Lote de %d itens inserido com sucesso!\n", workerId, len(batch))
			return nil
		}

		// Se ainda restam itens não processados, reenvia apenas eles
		request = resp.UnprocessedItems
		retries++
		fmt.Printf("[Worker %d] Retry %d - Restam %d itens não processados\n", workerId, retries, len(request[tableName]))

		// Backoff exponencial
		time.Sleep(time.Duration((1<<retries)*100+rand.Intn(300)) * time.Millisecond)
	}

	if len(request) > 0 {
		return fmt.Errorf("[Worker %d] %d itens não processados após %d tentativas", workerId, len(request[tableName]), retries)
	}
	return nil
}
