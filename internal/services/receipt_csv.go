package services

import (
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"
)

type ReceiptCsvService struct {
	BaseService
	CsvService
}

func NewReceiptCsvService(tx *gorm.DB) ReceiptCsvService {
	service := ReceiptCsvService{
		BaseService: BaseService{
			DB: repositories.GetDB(),
			TX: tx,
		},
		CsvService: NewCsvService(),
	}
	return service
}

func (service *ReceiptCsvService) BuildReceiptCsv(receipts []models.Receipt) ([]byte, error) {
	headers := []string{
		"id",
		"name",
	}
	rowData := make([][]string, 0, len(receipts))

	for _, receipt := range receipts {
		newRow := []string{
			utils.UintToString(receipt.ID),
			receipt.Name,
		}
		rowData = append(rowData, newRow)
	}

	buffer, err := service.CsvService.BuildCsv(headers, rowData)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
