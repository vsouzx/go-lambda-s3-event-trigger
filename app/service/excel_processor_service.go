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
	file, err := es.getFile(excelBytes)
    if err != nil {
        return fmt.Errorf("Erro ao abrir excel: " + err.Error())
    }
    defer file.Close()
    
    excelRows, err := es.getExcelRows(file)
	if err != nil {
		return fmt.Errorf("Erro ao obter linhas do excel: %w", err.Error())
	}
    defer excelRows.Close()

	batchChan := make(chan []dto.Acesso, 5)
	var wg sync.WaitGroup

	es.createWorkersToReadBatchesFromChanelAndSendToDynamo(batchChan, &wg)
	es.readExcelAndSendBatchesToChanel(excelRows, batchChan)

	close(batchChan)
	wg.Wait()

	fmt.Printf("Finalizado processamento do Excel\n")
	return nil
}

func (es *ExcelProcessorService) createWorkersToReadBatchesFromChanelAndSendToDynamo(batchChan <-chan []dto.Acesso, wg *sync.WaitGroup) {
	tableName := os.Getenv("DYNAMO_TABLE")
	workerCount, _ := strconv.Atoi(os.Getenv("WORKERS"))

	for i := range workerCount {
		wg.Add(1)
		go func(id int) {
			fmt.Println("Worker ", id, " iniciado")
			defer wg.Done()
			for batch := range batchChan {
				fmt.Printf("[Worker %d] Processando lote de tamanho %d\n", id, len(batch))
				if err := es.repository.BatchInsert(context.Background(), tableName, batch, id); err != nil {
					fmt.Printf("[Worker %d] Erro ao inserir lote: %v\n", id, err)
				}
			}
		}(i)
	}
}

func (es *ExcelProcessorService) getFile(excelBytes []byte) (*excelize.File, error) {
	f, err := excelize.OpenReader(bytes.NewReader(excelBytes))
	if err != nil {
		return nil, fmt.Errorf("Erro ao abrir excel: %s", err.Error())
	}

    return f, nil
}

func (es *ExcelProcessorService) getExcelRows(file *excelize.File) (*excelize.Rows, error) {
	sheet := file.GetSheetList()[0]
	rows, err := file.Rows(sheet)
	if err != nil {
		return nil, fmt.Errorf("Erro ao iterar linhas: %s", err.Error())
	}

	return rows, nil
}

func (es *ExcelProcessorService) readExcelAndSendBatchesToChanel(rows *excelize.Rows, batchChan chan<- []dto.Acesso) error {
	fmt.Println("Iniciando leitura do Excel e envio de lotes para o canal")
	batchSize, _ := strconv.Atoi(os.Getenv("BATCH_SIZE"))
	batch := make([]dto.Acesso, 0, batchSize)
	line := 0

	for rows.Next() {
		cols, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("erro lendo colunas: %w", err)
		}

		if line == 0 {
			fmt.Println("continue")
			line++
			continue
		}
		fmt.Println("passou continue")

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

	return nil
}
