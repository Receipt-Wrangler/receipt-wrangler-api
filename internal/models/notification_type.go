package models

import (
	"database/sql/driver"
)

type NotificationType string

const (
	NOTIFICATION_TYPE_NORMAL NotificationType = "NORMAL"
	NOTIFICATION_TYPE_URGENT NotificationType = "URGENT"
)

func (self *NotificationType) Scan(value string) error {
	*self = NotificationType(value)
	return nil
}

func (self NotificationType) Value() (driver.Value, error) {
	return string(self), nil
}
