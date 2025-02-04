package models

type TaskQueueConfiguration struct {
	BaseModel
	Name             QueueName      `json:"name" gorm:"unique"`
	Priority         int            `json:"priority"`
	SystemSettings   SystemSettings `json:"-"`
	SystemSettingsId uint           `json:"systemSettingsId"`
}

func GetDefaultQuickScanQueueConfiguration() TaskQueueConfiguration {
	return TaskQueueConfiguration{
		Name:     QuickScanQueue,
		Priority: 4,
	}
}

func GetDefaultEmailReceiptProcessingQueueConfiguration() TaskQueueConfiguration {
	return TaskQueueConfiguration{
		Name:     EmailReceiptProcessingQueue,
		Priority: 3,
	}
}

func GetDefaultEmailPollingQueueConfiguration() TaskQueueConfiguration {
	return TaskQueueConfiguration{
		Name:     EmailPollingQueue,
		Priority: 2,
	}
}

func GetDefaultEmailReceiptImageCleanupQueueConfiguration() TaskQueueConfiguration {
	return TaskQueueConfiguration{
		Name:     EmailReceiptImageCleanupQueue,
		Priority: 1,
	}
}

func GetDefaultSystemCleanupQueueConfiguration() TaskQueueConfiguration {
	return TaskQueueConfiguration{
		Name:     SystemCleanUpQueue,
		Priority: 5,
	}
}

func GetAllDefaultQueueConfigurations() []TaskQueueConfiguration {
	return []TaskQueueConfiguration{
		GetDefaultQuickScanQueueConfiguration(),
		GetDefaultEmailReceiptProcessingQueueConfiguration(),
		GetDefaultEmailPollingQueueConfiguration(),
		GetDefaultEmailReceiptImageCleanupQueueConfiguration(),
		GetDefaultSystemCleanupQueueConfiguration(),
	}
}
