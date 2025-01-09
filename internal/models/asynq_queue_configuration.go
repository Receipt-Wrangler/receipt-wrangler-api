package models

type AsynqQueueConfiguration struct {
	BaseModel
	SystemSettings   SystemSettings `json:"-"`
	SystemSettingsId uint           `json:"-"`
	Name             QueueName      `json:"name"`
	Priority         int            `json:"priority"`
}
