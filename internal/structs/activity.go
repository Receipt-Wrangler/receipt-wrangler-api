package structs

import (
	"receipt-wrangler/api/internal/models"
	"time"
)

type Activity struct {
	Id                     uint                    `json:"id"`
	Type                   models.SystemTaskType   `json:"type"`
	Status                 models.SystemTaskStatus `json:"status"`
	StartedAt              time.Time               `json:"startedAt"`
	EndedAt                *time.Time              `json:"endedAt"`
	RanByUserId            *uint                   `json:"ranByUserId"`
	ReceiptId              *uint                   `json:"receiptId"`
	GroupId                *uint                   `json:"groupId"`
	CanBeRestarted         bool                    `json:"canBeRestarted"`
	AssociatedSystemTaskId *uint                   `json:"-"`
}
