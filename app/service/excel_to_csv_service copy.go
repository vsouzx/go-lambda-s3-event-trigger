package service

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/xuri/excelize/v2"
)

type ExcelToCsvService struct {
}

func NewExcelToCsvService() *ExcelToCsvService {
	return &ExcelToCsvService{}
}

func (es *ExcelToCsvService) ConvertExcelToCSV(excelBytes []byte, outputPath string) error {
	f, err := excelize.OpenReader(bytes.NewReader(excelBytes))
	if err != nil {
		return fmt.Errorf("erro ao abrir excel: %w", err)
	}
	defer f.Close()

	sheet := f.GetSheetList()[0]
	rows, err := f.Rows(sheet)
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

	for rows.Next() {
		cols, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("erro lendo colunas: %w", err)
		}

		if err := writer.Write(cols); err != nil {
			return fmt.Errorf("erro escrevendo linha csv: %w", err)
		}
	}
	return nil
}
