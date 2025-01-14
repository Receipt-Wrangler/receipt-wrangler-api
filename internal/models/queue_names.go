package models

import (
	"database/sql/driver"
	"errors"
)

type QueueName string

const (
	QuickScanQueue                QueueName = "quick_scan"
	EmailPollingQueue             QueueName = "email_polling"
	EmailReceiptProcessingQueue   QueueName = "email_receipt_processing"
	EmailReceiptImageCleanupQueue QueueName = "email_receipt_image_cleanup"
)

func (name *QueueName) Scan(value string) error {
	*name = QueueName(value)
	return nil
}

func (name QueueName) Value() (driver.Value, error) {
	if name != QuickScanQueue &&
		name != EmailPollingQueue &&
		name != EmailReceiptProcessingQueue &&
		name != EmailReceiptImageCleanupQueue {
		return nil, errors.New("invalid queue name")
	}

	return string(name), nil
}

func GetQueueNames() []QueueName {
	return []QueueName{
		QuickScanQueue,
		EmailPollingQueue,
		EmailReceiptProcessingQueue,
		EmailReceiptImageCleanupQueue,
	}
}

func GetDefaultQueueConfigurationMap() map[QueueName]TaskQueueConfiguration {
	return map[QueueName]TaskQueueConfiguration{
		QuickScanQueue:                GetDefaultQuickScanQueueConfiguration(),
		EmailPollingQueue:             GetDefaultEmailPollingQueueConfiguration(),
		EmailReceiptProcessingQueue:   GetDefaultEmailReceiptProcessingQueueConfiguration(),
		EmailReceiptImageCleanupQueue: GetDefaultEmailReceiptImageCleanupQueueConfiguration(),
	}
}
