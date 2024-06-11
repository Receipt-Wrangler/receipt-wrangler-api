package structs

import "receipt-wrangler/api/internal/models"

type ReceiptProcessingSystemTasks struct {
	SystemTask         models.SystemTask
	FallbackSystemTask models.SystemTask
}
