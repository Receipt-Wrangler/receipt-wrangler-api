package commands

import "receipt-wrangler/api/internal/models"

type UpsertTaskQueueConfigurationCommand struct {
	Id       *uint            `json:"id"`
	Name     models.QueueName `json:"name"`
	Priority int              `json:"priority"`
}
