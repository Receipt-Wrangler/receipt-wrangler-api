package models

import (
	"database/sql/driver"
	"errors"
)

type CustomFieldType string

const (
	TEXT     CustomFieldType = "TEXT"
	DATE     CustomFieldType = "DATE"
	SELECT   CustomFieldType = "SELECT"
	CURRENCY CustomFieldType = "CURRENCY"
)

func (fieldType *CustomFieldType) Scan(value string) error {
	*fieldType = CustomFieldType(value)
	return nil
}

func (fieldType CustomFieldType) Value() (driver.Value, error) {
	if len(fieldType) == 0 {
		return "", nil
	}

	if fieldType != TEXT && fieldType != DATE && fieldType != SELECT && fieldType != CURRENCY {
		return "", errors.New("invalid custom field type")
	}
	return string(fieldType), nil
}
