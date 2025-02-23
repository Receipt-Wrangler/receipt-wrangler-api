package services

import (
	"bytes"
	"encoding/csv"
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"
)

type ReceiptCsvService struct {
	BaseService
}

func NewReceiptCsvService(tx *gorm.DB) ReceiptCsvService {
	service := ReceiptCsvService{BaseService: BaseService{
		DB: repositories.GetDB(),
		TX: tx,
	}}
	return service
}

func (s *ReceiptCsvService) GetReceiptCsv(receipts []models.Receipt) (*bytes.Buffer, error) {
	headers := []string{
		"id",
		"name",
	}
	rowData := make([][]string, 0, len(receipts)+1)
	rowData = append(rowData, headers)

	for _, receipt := range receipts {
		newRow := []string{
			utils.UintToString(receipt.ID),
			receipt.Name,
		}
		rowData = append(rowData, newRow)
	}

	buffer := new(bytes.Buffer)
	writer := csv.NewWriter(buffer)
	err := writer.WriteAll(rowData)
	if err != nil {
		return nil, err
	}

	return buffer, nil
}
