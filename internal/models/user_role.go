package models

import "database/sql/driver"

type UserRole string

const (
	ADMIN UserRole = "ADMIN"
	USER  UserRole = "USER"
)

func (self *UserRole) Scan(value string) error {
	*self = UserRole(value)
	return nil
}

func (self UserRole) Value() (driver.Value, error) {
	return string(self), nil
}
