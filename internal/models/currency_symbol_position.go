package models

import (
	"database/sql/driver"
	"errors"
)

type CurrencySymbolPosition string

const (
	START CurrencySymbolPosition = "START"
	END   CurrencySymbolPosition = "END"
)

func (currencySymbolPosition *CurrencySymbolPosition) Scan(value string) error {
	*currencySymbolPosition = CurrencySymbolPosition(value)
	return nil
}

func (currencySymbolPosition CurrencySymbolPosition) Value() (driver.Value, error) {
	if len(currencySymbolPosition) == 0 {
		return "", nil
	}

	if currencySymbolPosition != START && currencySymbolPosition != END {
		return nil, errors.New("invalid currency symbol position")
	}
	return string(currencySymbolPosition), nil
}
