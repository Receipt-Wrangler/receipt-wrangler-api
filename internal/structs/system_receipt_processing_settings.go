package structs

import "receipt-wrangler/api/internal/models"

type SystemReceiptProcessingSettings struct {
	ReceiptProcessingSettings         models.ReceiptProcessingSettings
	FallbackReceiptProcessingSettings models.ReceiptProcessingSettings
}
