package models

import (
	"database/sql/driver"
	"errors"
)

type ShareStatus string

const (
	SHARE_OPEN     ShareStatus = "OPEN"
	SHARE_RESOLVED ShareStatus = "RESOLVED"
	SHARE_DRAFT    ShareStatus = "DRAFT"
)

func (self *ShareStatus) Scan(value string) error {
	*self = ShareStatus(value)
	return nil
}

func (self ShareStatus) Value() (driver.Value, error) {
	if self != SHARE_OPEN && self != SHARE_RESOLVED && self != SHARE_DRAFT {
		return nil, errors.New("invalid itemStatus")
	}
	return string(self), nil
}
