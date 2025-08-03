package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/xuri/excelize/v2"
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

	line := 0
	var batch []map[string]string

	flushBatch := func() error {
		if len(batch) == 0 {
			return nil
		}

		// Montar requests
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

		// Enviar para DynamoDB
		_, err := es.dynamoClient.BatchWriteItem(context.Background(), &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				tableName: writeReq,
			},
		})
		if err != nil {
			return fmt.Errorf("erro no BatchWriteItem: %w", err)
		}

		fmt.Printf("✅ Inserido lote de %d itens\n", len(batch))
		batch = batch[:0] // limpa
		return nil
	}

	// Leitura do CSV
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("erro lendo csv: %w", err)
		}

		line++
		if line == 1 {
			continue // cabeçalho
		}
		if len(record) < 3 {
			continue
		}

		batch = append(batch, map[string]string{
			"funcionalChefe":       record[0],
			"funcionalColaborador": record[1],
			"departamento":         record[2],
		})

		// Quando atingir 25, envia
		if len(batch) == 25 {
			if err := flushBatch(); err != nil {
				return err
			}
		}
	}

	// Envia o restante
	return flushBatch()
}
