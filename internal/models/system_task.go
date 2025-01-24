package models

import (
	"database/sql/driver"
	"errors"
	"time"
)

type SystemTask struct {
	BaseModel
	Type                   SystemTaskType       `json:"type"`
	Status                 SystemTaskStatus     `json:"status"`
	AssociatedEntityType   AssociatedEntityType `json:"associatedEntityType"`
	AssociatedEntityId     uint                 `json:"associatedEntityId"`
	StartedAt              time.Time            `json:"startedAt"`
	EndedAt                *time.Time           `json:"endedAt"`
	ResultDescription      string               `json:"resultDescription"`
	RanByUser              *User                `json:"-"`
	RanByUserId            *uint                `json:"ranByUserId"`
	GroupId                *uint                `json:"groupId"`
	AssociatedSystemTask   *SystemTask          `json:"associatedSystemTask"`
	AssociatedSystemTaskId *uint                `json:"associatedSystemTaskId"`
	ChildSystemTasks       []*SystemTask        `gorm:"foreignKey:AssociatedSystemTaskId" json:"childSystemTasks"`
	AsynqTaskId            string               `json:"asynqTaskId"`
}

type SystemTaskStatus string

const (
	SYSTEM_TASK_SUCCEEDED SystemTaskStatus = "SUCCEEDED"
	SYSTEM_TASK_FAILED    SystemTaskStatus = "FAILED"
)

func (self *SystemTaskStatus) Scan(value string) error {
	*self = SystemTaskStatus(value)
	return nil
}

func (self SystemTaskStatus) Value() (driver.Value, error) {
	if self != SYSTEM_TASK_SUCCEEDED && self != SYSTEM_TASK_FAILED {
		return nil, errors.New("invalid SystemTaskStatus")
	}
	return string(self), nil
}

type SystemTaskType string

const (
	META_COMBINE_QUICK_SCAN                        SystemTaskType = "META_COMBINE_QUICK_SCAN"
	META_ASSOCIATE_TASKS_TO_RECEIPT                SystemTaskType = "META_ASSOCIATE_TASKS_TO_RECEIPT"
	RECEIPT_UPLOADED                               SystemTaskType = "RECEIPT_UPLOADED"
	OCR_PROCESSING                                 SystemTaskType = "OCR_PROCESSING"
	CHAT_COMPLETION                                SystemTaskType = "CHAT_COMPLETION"
	MAGIC_FILL                                     SystemTaskType = "MAGIC_FILL"
	QUICK_SCAN                                     SystemTaskType = "QUICK_SCAN"
	EMAIL_UPLOAD                                   SystemTaskType = "EMAIL_UPLOAD"
	EMAIL_READ                                     SystemTaskType = "EMAIL_READ"
	SYSTEM_EMAIL_CONNECTIVITY_CHECK                SystemTaskType = "SYSTEM_EMAIL_CONNECTIVITY_CHECK"
	RECEIPT_PROCESSING_SETTINGS_CONNECTIVITY_CHECK SystemTaskType = "RECEIPT_PROCESSING_SETTINGS_CONNECTIVITY_CHECK"
	PROMPT_GENERATED                               SystemTaskType = "PROMPT_GENERATED"
	RECEIPT_UPDATED                                SystemTaskType = "RECEIPT_UPDATED"
)

func (self *SystemTaskType) Scan(value string) error {
	*self = SystemTaskType(value)
	return nil
}

func (self SystemTaskType) Value() (driver.Value, error) {
	if self != SYSTEM_EMAIL_CONNECTIVITY_CHECK &&
		self != RECEIPT_PROCESSING_SETTINGS_CONNECTIVITY_CHECK &&
		self != QUICK_SCAN &&
		self != MAGIC_FILL &&
		self != EMAIL_UPLOAD &&
		self != EMAIL_READ &&
		self != CHAT_COMPLETION &&
		self != OCR_PROCESSING &&
		self != RECEIPT_UPLOADED &&
		self != META_COMBINE_QUICK_SCAN &&
		self != META_ASSOCIATE_TASKS_TO_RECEIPT &&
		self != PROMPT_GENERATED &&
		self != RECEIPT_UPDATED {
		return nil, errors.New("invalid SystemTaskType")
	}
	return string(self), nil
}

type AssociatedEntityType string

const (
	RECEIPT                     AssociatedEntityType = "RECEIPT"
	SYSTEM_EMAIL                AssociatedEntityType = "SYSTEM_EMAIL"
	PROMPT                      AssociatedEntityType = "PROMPT"
	RECEIPT_PROCESSING_SETTINGS AssociatedEntityType = "RECEIPT_PROCESSING_SETTINGS"
	NOOP_ENTITY_TYPE            AssociatedEntityType = "NOOP_ENTITY_TYPE"
)

func (self *AssociatedEntityType) Scan(value string) error {
	*self = AssociatedEntityType(value)
	return nil
}

func (self AssociatedEntityType) Value() (driver.Value, error) {
	if self != SYSTEM_EMAIL &&
		self != NOOP_ENTITY_TYPE &&
		self != RECEIPT_PROCESSING_SETTINGS &&
		self != RECEIPT &&
		self != PROMPT {
		return nil, errors.New("invalid AssociatedEntityType")
	}
	return string(self), nil
}
