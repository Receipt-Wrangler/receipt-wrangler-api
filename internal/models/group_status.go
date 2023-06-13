package models

import (
	"database/sql/driver"
)

type GroupStatus string

const (
	GROUP_ACTIVE   GroupStatus = "ACTIVE"
	GROUP_ARCHIVED GroupStatus = "ARCHIVED"
)

func (self *GroupStatus) Scan(value string) error {
	*self = GroupStatus(value)
	return nil
}

func (self GroupStatus) Value() (driver.Value, error) {
	return string(self), nil
}
