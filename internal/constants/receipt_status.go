package constants

import "receipt-wrangler/api/internal/models"

func ReceiptStatuses() []interface{} {
	return []interface{}{models.OPEN, models.NEEDS_ATTENTION, models.RESOLVED}
}
