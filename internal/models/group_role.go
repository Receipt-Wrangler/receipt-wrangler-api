package models

import "database/sql/driver"

type GroupRole string

const (
	OWNER  GroupRole = "OWNER"
	VIEWER GroupRole = "VIEWER"
	EDITOR GroupRole = "EDITOR"
)

func (self *GroupRole) Scan(value interface{}) error {
	*self = GroupRole(value.([]byte))
	return nil
}

func (self GroupRole) Value() (driver.Value, error) {
	return string(self), nil
}
