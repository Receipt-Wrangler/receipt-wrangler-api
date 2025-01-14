package models

import (
	"database/sql/driver"
	"errors"
)

type ReceiptStatus string

const (
	OPEN            ReceiptStatus = "OPEN"
	NEEDS_ATTENTION ReceiptStatus = "NEEDS_ATTENTION"
	RESOLVED        ReceiptStatus = "RESOLVED"
	DRAFT           ReceiptStatus = "DRAFT"
)

func (self *ReceiptStatus) Scan(value string) error {
	*self = ReceiptStatus(value)
	return nil
}

func (self ReceiptStatus) Value() (driver.Value, error) {
	if self != OPEN && self != NEEDS_ATTENTION && self != RESOLVED && self != DRAFT && self != "" {
		return nil, errors.New("invalid receiptStatus")
	}
	return string(self), nil
}

func ReceiptStatuses() []interface{} {
	return []interface{}{OPEN, NEEDS_ATTENTION, RESOLVED, DRAFT}
}
