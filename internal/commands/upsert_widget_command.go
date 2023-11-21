package commands

import "receipt-wrangler/api/internal/models"

type UpsertWidgetCommand struct {
	Name       string            `json:"name"`
	WidgetType models.WidgetType `json:"widgetType"`
}
