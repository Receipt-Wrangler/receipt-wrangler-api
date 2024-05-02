package commands

import (
	"receipt-wrangler/api/internal/models"
	"time"
)

type UpsertSystemTaskCommand struct {
	Type                 models.SystemTaskType       `json:"type"`
	Status               models.SystemTaskStatus     `json:"status"`
	AssociatedEntityType models.AssociatedEntityType `json:"associatedEntityType"`
	AssociatedEntityId   uint                        `json:"associatedEntityId"`
	StartedAt            time.Time                   `json:"startedAt"`
	EndedAt              *time.Time                  `json:"endedAt"`
	ResultDescription    string                      `json:"resultDescription"`
}
