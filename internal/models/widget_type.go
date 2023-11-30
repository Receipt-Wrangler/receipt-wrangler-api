package models

import (
	"database/sql/driver"
	"errors"
)

type WidgetType string

const (
	GROUP_SUMMARY WidgetType = "GROUP_SUMMARY"
)

func (widgetType *WidgetType) Scan(value string) error {
	*widgetType = WidgetType(value)
	return nil
}

func (widgetType WidgetType) Value() (driver.Value, error) {
	if widgetType != GROUP_SUMMARY {
		return nil, errors.New("invalid widget type")
	}
	return string(widgetType), nil
}
