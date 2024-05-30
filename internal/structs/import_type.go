package structs

import (
	"database/sql/driver"
	"errors"
)

type ImportType string

const (
	ImportConfig ImportType = "IMPORT_CONFIG"
)

func (importType *ImportType) Scan(value string) error {
	*importType = ImportType(value)
	return nil
}

func (importType ImportType) Value() (driver.Value, error) {
	if len(importType) == 0 {
		return "", nil
	}

	if importType != ImportConfig {
		return nil, errors.New("invalid import type")
	}
	return string(importType), nil
}
