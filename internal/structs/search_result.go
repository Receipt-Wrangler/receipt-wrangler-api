package structs

import (
	"github.com/shopspring/decimal"
	"receipt-wrangler/api/internal/models"
	"time"
)

type SearchResult struct {
	ID            uint                 `json:"id"`
	Name          string               `json:"name"`
	Type          string               `json:"type"`
	GroupID       uint                 `json:"groupId"`
	Date          time.Time            `json:"date"`
	CreatedAt     time.Time            `json:"createdAt"`
	Amount        decimal.Decimal      `json:"amount"`
	ReceiptStatus models.ReceiptStatus `json:"receiptStatus"`
	PaidByUserId  uint                 `json:"paidByUserId"`
}
