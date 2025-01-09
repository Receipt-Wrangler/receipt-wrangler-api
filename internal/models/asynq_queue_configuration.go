package models

type AsynqQueueConfiguration struct {
	SystemSettings   SystemSettings
	SystemSettingsId uint
	QueueName        QueueName
	Priority         int
}
