package models

import (
	"database/sql/driver"
	"fmt"
)

type GroupRole string

const (
	OWNER  GroupRole = "OWNER"
	VIEWER GroupRole = "VIEWER"
	EDITOR GroupRole = "EDITOR"
)

func (self *GroupRole) Scan(value string) error {
	fmt.Println(value)
	*self = GroupRole(value)
	return nil
}

func (self GroupRole) Value() (driver.Value, error) {
	return string(self), nil
}
