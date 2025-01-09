package models

type QueueName string

const (
	QuickScanQueue                QueueName = "quick_scan"
	EmailPollingQueue             QueueName = "email_polling"
	EmailReceiptProcessingQueue   QueueName = "email_receipt_processing"
	EmailReceiptImageCleanupQueue QueueName = "email_receipt_image_cleanup"
)
