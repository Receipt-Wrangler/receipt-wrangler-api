package commands

import "receipt-wrangler/api/internal/models"

type BulkStatusUpdateCommand struct {
	Comment    string
	Status     models.ReceiptStatus
	ReceiptIds []uint
}
