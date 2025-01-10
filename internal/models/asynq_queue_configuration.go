package models

type AsynqQueueConfiguration struct {
	BaseModel
	SystemSettings   SystemSettings `json:"-"`
	SystemSettingsId uint           `json:"-"`
	Name             QueueName      `json:"name"`
	Priority         int            `json:"priority"`
}

func GetDefaultQuickScanQueueConfiguration() AsynqQueueConfiguration {
	return AsynqQueueConfiguration{
		Name:     QuickScanQueue,
		Priority: 4,
	}
}

func GetDefaultEmailReceiptProcessingQueueConfiguration() AsynqQueueConfiguration {
	return AsynqQueueConfiguration{
		Name:     EmailReceiptProcessingQueue,
		Priority: 3,
	}
}

func GetDefaultEmailPollingQueueConfiguration() AsynqQueueConfiguration {
	return AsynqQueueConfiguration{
		Name:     EmailPollingQueue,
		Priority: 2,
	}
}

func GetDefaultEmailReceiptImageCleanupQueueConfiguration() AsynqQueueConfiguration {
	return AsynqQueueConfiguration{
		Name:     EmailReceiptImageCleanupQueue,
		Priority: 1,
	}
}

func GetAllDefaultQueueConfigurations() []AsynqQueueConfiguration {
	return []AsynqQueueConfiguration{
		GetDefaultQuickScanQueueConfiguration(),
		GetDefaultEmailReceiptProcessingQueueConfiguration(),
		GetDefaultEmailPollingQueueConfiguration(),
		GetDefaultEmailReceiptImageCleanupQueueConfiguration(),
	}
}