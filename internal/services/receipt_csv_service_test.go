package services

import (
	"github.com/shopspring/decimal"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"testing"
)

func TestShouldBuildReceiptCsv(t *testing.T) {
	expected :=
		"id,name\n" +
			"1,test\n"

	service := NewReceiptCsvService()
	receipts := []models.Receipt{
		models.Receipt{
			BaseModel: models.BaseModel{
				ID: 1,
			},
			Name:   "test",
			Amount: decimal.NewFromFloat(123.45),
		},
	}

	result, err := service.BuildReceiptCsv(receipts)
	if err != nil {
		utils.PrintTestError(t, result, expected)
	}

	bytes := result.ReceiptCsvBytes
	if string(bytes) != expected {
		utils.PrintTestError(t, string(bytes), expected)
	}
}
