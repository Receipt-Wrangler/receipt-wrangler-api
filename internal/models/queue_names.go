package models

type QueueName string

const (
	QuickScanQueue                QueueName = "quick_scan"
	EmailPollingQueue             QueueName = "email_polling"
	EmailReceiptProcessingQueue   QueueName = "email_receipt_processing"
	EmailReceiptImageCleanupQueue QueueName = "email_receipt_image_cleanup"
)

func GetQueueNames() []QueueName {
	return []QueueName{
		QuickScanQueue,
		EmailPollingQueue,
		EmailReceiptProcessingQueue,
		EmailReceiptImageCleanupQueue,
	}
}

func GetDefaultQueueConfigurationMap() map[QueueName]AsynqQueueConfiguration {
	return map[QueueName]AsynqQueueConfiguration{
		QuickScanQueue:                GetDefaultQuickScanQueueConfiguration(),
		EmailPollingQueue:             GetDefaultEmailPollingQueueConfiguration(),
		EmailReceiptProcessingQueue:   GetDefaultEmailReceiptProcessingQueueConfiguration(),
		EmailReceiptImageCleanupQueue: GetDefaultEmailReceiptImageCleanupQueueConfiguration(),
	}
}
