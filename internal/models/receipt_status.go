package models

import (
	"database/sql/driver"
)

type ReceiptStatus string

const (
	OPEN            ReceiptStatus = "OPEN"
	NEEDS_ATTENTION ReceiptStatus = "NEEDSATTENTION"
	RESOLVED        ReceiptStatus = "RESOLVED"
)

func (self *ReceiptStatus) Scan(value string) error {
	*self = ReceiptStatus(value)
	return nil
}

func (self ReceiptStatus) Value() (driver.Value, error) {
	return string(self), nil
}
