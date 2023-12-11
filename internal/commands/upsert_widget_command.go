package commands

import (
	"encoding/json"
	"receipt-wrangler/api/internal/models"
)

type UpsertWidgetCommand struct {
	Name          string            `json:"name"`
	WidgetType    models.WidgetType `json:"widgetType"`
	Configuration json.RawMessage   `json:"configuration"`
}
