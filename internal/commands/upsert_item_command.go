package commands

import (
	"github.com/shopspring/decimal"
	"receipt-wrangler/api/internal/models"
)

type UpsertItemCommand struct {
	Amount          decimal.Decimal   `json:"amount"`
	ChargedToUserId uint              `json:"chargedToUserId"`
	IsTaxed         bool              `json:"isTaxed"`
	Name            string            `json:"name"`
	ReceiptId       uint              `json:"receiptId"`
	Status          models.ItemStatus `json:"status"`
}
