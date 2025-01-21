package models

import (
	"database/sql/driver"
	"errors"
)

type WidgetType string

const (
	GROUP_SUMMARY     WidgetType = "GROUP_SUMMARY"
	FILTERED_RECEIPTS WidgetType = "FILTERED_RECEIPTS"
	GROUP_ACTIVITY    WidgetType = "GROUP_ACTIVITY"
)

func (widgetType *WidgetType) Scan(value string) error {
	*widgetType = WidgetType(value)
	return nil
}

func (widgetType WidgetType) Value() (driver.Value, error) {
	if widgetType != GROUP_SUMMARY &&
		widgetType != FILTERED_RECEIPTS &&
		widgetType != GROUP_ACTIVITY {
		return nil, errors.New("invalid widget type")
	}
	return string(widgetType), nil
}
