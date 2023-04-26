package structs

import "receipt-wrangler/api/internal/models"

type BulkResolve struct {
	Comment    string
	Status     models.ReceiptStatus
	ReceiptIds []uint
}
