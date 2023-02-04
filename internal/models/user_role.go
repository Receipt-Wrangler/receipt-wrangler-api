package models

import "database/sql/driver"

type UserRole string

const (
	ADMIN  UserRole = "OWNER"
	USER UserRole = "USER"
)

func (self *UserRole) Scan(value interface{}) error {
	*self = UserRole(value.([]byte))
	return nil
}

func (self UserRole) Value() (driver.Value, error) {
	return string(self), nil
}
