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

func TestShouldBuildShareCsv(t *testing.T) {
	expected :=
		"Id,Receipt Id,Receipt Name,Receipt Date,Name,Charged to User,Amount,Status,Categories,Tags\n" +
			"1,2,Test Receipt,2025-01-01,Test Share,John,25.5,OPEN,\"Groceries,Food\",\"Essential,Bill\"\n" +
			"2,3,Another Receipt,2025-01-02,Another Share,Jane,15.75,RESOLVED,Electronics,Gadget\n"

	date1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	service := NewReceiptCsvService()
	shares := []models.Share{
		{
			BaseModel: models.BaseModel{ID: 1},
			ReceiptId: 2,
			Receipt: models.Receipt{
				Name: "Test Receipt",
				Date: date1,
			},
			Name:          "Test Share",
			ChargedToUser: models.User{DisplayName: "John"},
			Amount:        decimal.NewFromFloat(25.50),
			Status:        models.SHARE_OPEN,
			Categories: []models.Category{
				{Name: "Groceries"},
				{Name: "Food"},
			},
			Tags: []models.Tag{
				{Name: "Essential"},
				{Name: "Bill"},
			},
		},
		{
			BaseModel: models.BaseModel{ID: 2},
			ReceiptId: 3,
			Receipt: models.Receipt{
				Name: "Another Receipt",
				Date: date2,
			},
			Name:          "Another Share",
			ChargedToUser: models.User{DisplayName: "Jane"},
			Amount:        decimal.NewFromFloat(15.75),
			Status:        models.SHARE_RESOLVED,
			Categories: []models.Category{
				{Name: "Electronics"},
			},
			Tags: []models.Tag{
				{Name: "Gadget"},
			},
		},
	}

	result, err := service.BuildShareCsv(shares)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	if string(result) != expected {
		utils.PrintTestError(t, string(result), expected)
	}
}
