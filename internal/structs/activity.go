package structs

import (
	"receipt-wrangler/api/internal/models"
	"time"
)

type Activity struct {
	Id                uint                    `json:"id"`
	Type              models.SystemTaskType   `json:"type"`
	Status            models.SystemTaskStatus `json:"status"`
	StartedAt         time.Time               `json:"startedAt"`
	EndedAt           *time.Time              `json:"endedAt"`
	ResultDescription string                  `json:"resultDescription"`
	RanByUserId       *uint                   `json:"ranByUserId"`
	CanBeRestarted    bool                    `json:"canBeRestarted"`
}
