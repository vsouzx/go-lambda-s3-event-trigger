package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

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

	// Leitura do CSV
	for {
		record, err := reader.Read()
		if err == io.EOF {
			fmt.Println("Fim do arquivo CSV")
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
			if err := es.flushBatch(context.Background(), tableName, batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	return es.flushBatch(context.Background(), tableName, batch)
}

func (es *ExcelService) flushBatch(ctx context.Context, tableName string, batch []map[string]string) error {
    if len(batch) == 0 {
        return nil
    }

    // Monta WriteRequests
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

    // Payload inicial
    request := map[string][]types.WriteRequest{
        tableName: writeReq,
    }

    // 🔥 Loop até esvaziar UnprocessedItems
    retries := 0
    for len(request) > 0 && retries < 20 { // 20 tentativas antes de desistir
        resp, err := es.dynamoClient.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
            RequestItems: request,
        })
        if err != nil {
            return fmt.Errorf("erro no BatchWriteItem: %w", err)
        }

        if len(resp.UnprocessedItems) == 0 {
            fmt.Printf("✅ Lote de %d itens inserido com sucesso!\n", len(batch))
            break
        }

        fmt.Printf("⚠️ %d itens não processados, retry %d...\n",
            len(resp.UnprocessedItems[tableName]), retries+1)

        request = resp.UnprocessedItems
        retries++

        time.Sleep(time.Duration((1<<retries)*100+rand.Intn(200)) * time.Millisecond)
    }

    if len(request) > 0 {
        return fmt.Errorf("❌ alguns itens não foram processados após %d tentativas", retries)
    }

    return nil
}
