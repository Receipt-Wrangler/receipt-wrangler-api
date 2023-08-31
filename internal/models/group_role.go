package models

import (
	"errors"
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

func (self GroupRole) Value() (string, error) {
	value := string(self)
	if value != "OWNER" && value != "VIEWER" && value != "EDITOR" {
		return "", errors.New("invalid GroupRole value")
	}
	return string(self), nil
}
