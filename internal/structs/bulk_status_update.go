package structs

import "receipt-wrangler/api/internal/models"

type BulkStatusUpdate struct {
	Comment    string
	Status     models.ReceiptStatus
	ReceiptIds []uint
}
