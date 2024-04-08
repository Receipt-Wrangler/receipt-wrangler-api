package commands

import (
	"github.com/shopspring/decimal"
	"receipt-wrangler/api/internal/models"
	"time"
)

type ReceiptUpsertCommand struct {
	Name         string               `json:"name"`
	Amount       decimal.Decimal      `json:"amount"`
	Date         time.Time            `json:"date"`
	GroupId      uint                 `json:"groupId"`
	PaidByUserID uint                 `json:"paidByUserId"`
	Status       models.ReceiptStatus `json:"status"`
	Categories   []models.Category    `json:"categories"`
	Tags         []models.Tag         `json:"tags"`
	Items        []models.Item        `json:"receiptItems"`
}
