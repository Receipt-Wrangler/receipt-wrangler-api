package models

import (
	"database/sql/driver"
	"errors"
)

type ItemStatus string

const (
	ITEM_OPEN     ItemStatus = "OPEN"
	ITEM_RESOLVED ItemStatus = "RESOLVED"
	ITEM_DRAFT	  ItemStatus = "DRAFT"
)

func (self *ItemStatus) Scan(value string) error {
	*self = ItemStatus(value)
	return nil
}

func (self ItemStatus) Value() (driver.Value, error) {
	if self != ITEM_OPEN && self != ITEM_RESOLVED && self != ITEM_DRAFT {
		return nil, errors.New("invalid itemStatus")
	}
	return string(self), nil
}
