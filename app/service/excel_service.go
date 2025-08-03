package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
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
			continue // pula cabeçalho
		}
		if len(record) < 3 {
			continue
		}

		funcionalCp := record[0]
		funcionalColaborador := record[1]
		departamento := record[2]

		fmt.Printf("Linha %d: %s | %s | %s\n", line, funcionalCp, funcionalColaborador, departamento)

		item := map[string]string{
			"funcionalChefe":       funcionalCp,
			"funcionalColaborador": funcionalColaborador,
			"departamento":         departamento,
		}

		av, err := attributevalue.MarshalMap(item)
		if err != nil {
			log.Fatalf("Erro ao converter item: %v", err)
		}

		_, err = es.dynamoClient.PutItem(context.Background(), &dynamodb.PutItemInput{
			TableName: &tableName,
			Item:      av,
		})

		if err != nil {
			log.Fatalf("Erro ao inserir item: %v", err)
		} else {
			fmt.Println("Item inserido com sucesso no DynamoDB!")
		}
	}
	return nil
}
