package models

import (
	"database/sql/driver"
)

type GroupRole string

const (
	OWNER  GroupRole = "OWNER"
	VIEWER GroupRole = "VIEWER"
	EDITOR GroupRole = "EDITOR"
)

func (self *GroupRole) Scan(value string) error {
	*self = GroupRole(value)
	return nil
}

func (self GroupRole) Value() (driver.Value, error) {
	return string(self), nil
}
