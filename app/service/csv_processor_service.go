package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/vsouzx/go-lambda-s3-event-trigger/repository"
)

type CsvProcessorService struct {
	repository *repository.Repository
}

func NewCsvProcessorService(repository *repository.Repository) *CsvProcessorService {
	return &CsvProcessorService{
		repository: repository,
	}
}

func (es *CsvProcessorService) ProcessCSVFile(csvPath string) error {
	tableName := os.Getenv("DYNAMO_TABLE")
	workerCount, _ := strconv.Atoi(os.Getenv("WORKERS"))
	fmt.Println("workerCount:", workerCount)
	batchSize, _ := strconv.Atoi(os.Getenv("BATCH_SIZE"))
	fmt.Println("batch size:", batchSize)

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
