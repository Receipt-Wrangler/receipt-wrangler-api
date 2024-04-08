package commands

import (
	"github.com/shopspring/decimal"
	"receipt-wrangler/api/internal/models"
	"time"
)

type UpsertReceiptCommand struct {
	Name         string               `json:"name"`
	Amount       decimal.Decimal      `json:"amount"`
	Date         time.Time            `json:"date"`
	GroupId      uint                 `json:"groupId"`
	PaidByUserID uint                 `json:"paidByUserId"`
	Status       models.ReceiptStatus `json:"status"`
	Categories   []models.Category    `json:"categories"`
	Tags         []UpsertTagCommand   `json:"tags"`
	Items        []UpsertItemCommand  `json:"receiptItems"`
}
