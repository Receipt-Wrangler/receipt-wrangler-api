package models

import (
	"database/sql/driver"
)

type ItemStatus string

const (
	ITEM_OPEN     ItemStatus = "OPEN"
	ITEM_RESOLVED ItemStatus = "RESOLVED"
)

func (self *ItemStatus) Scan(value string) error {
	*self = ItemStatus(value)
	return nil
}

func (self ItemStatus) Value() (driver.Value, error) {
	return string(self), nil
}
