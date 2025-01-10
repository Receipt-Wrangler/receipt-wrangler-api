package commands

import "receipt-wrangler/api/internal/models"

type UpsertAsynqQueueConfigurationCommand struct {
	Id       *uint            `json:"id"`
	Name     models.QueueName `json:"name"`
	Priority int              `json:"priority"`
}
