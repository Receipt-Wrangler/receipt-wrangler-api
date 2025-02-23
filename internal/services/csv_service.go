package services

import (
	"bytes"
	"encoding/csv"
)

type CsvService struct {
}

func NewCsvService() CsvService {
	service := CsvService{}
	return service
}

func (service *CsvService) BuildCsv(headers []string, rowData [][]string) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	writer := csv.NewWriter(buffer)

	err := writer.Write(headers)
	if err != nil {
		return nil, err
	}

	err = writer.WriteAll(rowData)
	if err != nil {
		return nil, err
	}

	return buffer, nil
}
