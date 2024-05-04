package models

import (
	"database/sql/driver"
	"errors"
	"time"
)

type SystemTask struct {
	BaseModel
	Type                 SystemTaskType       `json:"type"`
	Status               SystemTaskStatus     `json:"status"`
	AssociatedEntityType AssociatedEntityType `json:"associatedEntityType"`
	AssociatedEntityId   uint                 `json:"associatedEntityId"`
	StartedAt            time.Time            `json:"startedAt"`
	EndedAt              *time.Time           `json:"endedAt"`
	ResultDescription    string               `json:"resultDescription"`
	RanByUserId          *uint                `json:"ranByUserId"`
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
	SYSTEM_EMAIL_CONNECTIVITY_CHECK SystemTaskType = "SYSTEM_EMAIL_CONNECTIVITY_CHECK"
)

func (self *SystemTaskType) Scan(value string) error {
	*self = SystemTaskType(value)
	return nil
}

func (self SystemTaskType) Value() (driver.Value, error) {
	if self != SYSTEM_EMAIL_CONNECTIVITY_CHECK {
		return nil, errors.New("invalid SystemTaskType")
	}
	return string(self), nil
}

type AssociatedEntityType string

const (
	SYSTEM_EMAIL AssociatedEntityType = "SYSTEM_EMAIL"
)

func (self *AssociatedEntityType) Scan(value string) error {
	*self = AssociatedEntityType(value)
	return nil
}

func (self AssociatedEntityType) Value() (driver.Value, error) {
	if self != SYSTEM_EMAIL {
		return nil, errors.New("invalid AssociatedEntityType")
	}
	return string(self), nil
}