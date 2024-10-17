package models

import (
	"database/sql/driver"
	"errors"
)

type CurrencySeparator string

const (
	COMMA CurrencySeparator = ","
	DOT   CurrencySeparator = "."
)

func (currencySeparator *CurrencySeparator) Scan(value string) error {
	*currencySeparator = CurrencySeparator(value)
	return nil
}

func (currencySeparator CurrencySeparator) Value() (driver.Value, error) {
	if len(currencySeparator) == 0 {
		return "", nil
	}

	if currencySeparator != COMMA && currencySeparator != DOT && currencySeparator != "" {
		return nil, errors.New("invalid currency symbol position")
	}
	return string(currencySeparator), nil
}
