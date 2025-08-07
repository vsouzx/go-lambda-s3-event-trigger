package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/vsouzx/go-lambda-s3-event-trigger/dto"
	"github.com/vsouzx/go-lambda-s3-event-trigger/repository"
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
	buffer := bufio.NewReader(bytes.NewReader(excelBytes))
	reader := csv.NewReader(buffer)
	reader.FieldsPerRecord = 3
	reader.Comma = ';'

	bufferSize, _ := strconv.Atoi(os.Getenv("BUFFER_SIZE"))

	batchChan := make(chan []dto.Acesso, bufferSize)
	var wg sync.WaitGroup

	es.createWorkersToReadBatchesFromChanelAndSendToDynamo(batchChan, &wg)
	es.readExcelAndSendBatchesToChanel(reader, batchChan)

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
				if err := es.repository.BatchInsert(context.Background(), tableName, batch, id); err != nil {
					fmt.Printf("[Worker %d] Erro ao inserir lote: %v\n", id, err)
				}
			}
		}(i)
	}
}

func (es *ExcelProcessorService) readExcelAndSendBatchesToChanel(reader *csv.Reader, batchChan chan<- []dto.Acesso) error {
	batchSize, _ := strconv.Atoi(os.Getenv("BATCH_SIZE"))
	batch := make([]dto.Acesso, 0, batchSize)
	line := 0

	for {
		cols, err := reader.Read()
		if err == io.EOF {
			fmt.Println("Fim do arquivo CSV")
			break
		}

		if line == 0 {
			line++
			continue
		}

		if len(cols) < 3 {
			fmt.Println()
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
