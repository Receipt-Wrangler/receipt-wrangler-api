package models

import "encoding/json"

type Widget struct {
	BaseModel
	Name          string          `json:"name"`
	Dashboard     Dashboard       `json:"-"`
	DashboardId   uint            `gorm:"not null;" json:"dashboardId"`
	WidgetType    WidgetType      `gorm:"not null;" json:"widgetType"`
	Configuration json.RawMessage `json:"configuration"`
}
