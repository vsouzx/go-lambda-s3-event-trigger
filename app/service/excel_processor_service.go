package service

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/vsouzx/go-lambda-s3-event-trigger/dto"
	"github.com/vsouzx/go-lambda-s3-event-trigger/repository"
	"github.com/xuri/excelize/v2"
)

type ExcelProcessorService struct {
	repository *repository.Repository
}

func NewExcelProcessorService(repository *repository.Repository) *ExcelProcessorService {
	return &ExcelProcessorService{
		repository: repository,
	}
}

func (es *ExcelProcessorService) ProcessExcelFile(excelBytes []byte) error {
	tableName := os.Getenv("DYNAMO_TABLE")
	workerCount, _ := strconv.Atoi(os.Getenv("WORKERS"))
	batchSize, _ := strconv.Atoi(os.Getenv("BATCH_SIZE"))

	f, err := excelize.OpenReader(bytes.NewReader(excelBytes))
	if err != nil {
		return fmt.Errorf("erro ao abrir excel: %w", err)
	}
	defer f.Close()

	sheet := f.GetSheetList()[0]
	rows, err := f.Rows(sheet)
	if err != nil {
		return fmt.Errorf("erro ao iterar linhas: %w", err)
	}
	defer rows.Close()

	// Canal para enviar lotes para workers
	batchChan := make(chan []dto.Acesso, 5)
	var wg sync.WaitGroup

	// Workers que vão processar os lotes
	for i := range workerCount {
		fmt.Println("Starting worker:", i)
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for batch := range batchChan {
				if err := es.repository.BatchInsert(context.Background(), tableName, batch, id); err != nil {
					fmt.Printf("[Worker %d] Erro ao inserir lote: %v\n", id, err)
				}
			}
		}(i)
	}

	// Leitura do Excel e envio para o canal
	line := 0
	batch := make([]dto.Acesso, 0, batchSize)

	for rows.Next() {
		cols, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("erro lendo colunas: %w", err)
		}

		line++
		if line == 1 {
			continue
		}
		if len(cols) < 3 {
			continue
		}

		batch = append(batch, dto.Acesso{
			FuncionalChefe:       cols[0],
			FuncionalColaborador: cols[1],
			Departamento:         cols[2],
		})

		if len(batch) == batchSize {
			b := make([]dto.Acesso, len(batch))
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

	fmt.Printf("Finalizado processamento do Excel com %d linhas\n", line)
	return nil
}
