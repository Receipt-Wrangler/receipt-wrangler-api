package services

import (
	"github.com/shopspring/decimal"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"testing"
	"time"
)

func TestShouldBuildReceiptCsv(t *testing.T) {
	expected :=
		"Id,Added At,Receipt Date,Name,Paid By,Amount,Status,Categories,Tags,Resolved Date\n" +
			"1,2025-01-01,2025-01-01,test,Jim,123.45,OPEN,\"Groceries,Food\",\"Bill,Essential\",2025-01-01\n"

	date := time.Date(
		2025, 1, 1, 0, 0, 0, 0, time.UTC)
	service := NewReceiptCsvService()
	receipts := []models.Receipt{
		models.Receipt{
			BaseModel: models.BaseModel{
				ID:        1,
				CreatedAt: date,
			},
			Date:       date,
			Name:       "test",
			PaidByUser: models.User{DisplayName: "Jim"},
			Amount:     decimal.NewFromFloat(123.45),
			Status:     models.OPEN,
			Categories: []models.Category{
				models.Category{Name: "Groceries"},
				models.Category{Name: "Food"},
			},
			Tags: []models.Tag{
				models.Tag{Name: "Bill"},
				models.Tag{Name: "Essential"},
			},
			ResolvedDate: &date,
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
