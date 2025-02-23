package services

import (
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
			Name: "test",
		},
	}

	bytes, err := service.BuildReceiptCsv(receipts)
	if err != nil {
		utils.PrintTestError(t, string(bytes), expected)
	}

	if string(bytes) != expected {
		utils.PrintTestError(t, string(bytes), expected)
	}
}
