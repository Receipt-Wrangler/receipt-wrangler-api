package commands

import "receipt-wrangler/api/internal/models"

type UpsertDashboardCommand struct {
	Name    string          `json:"name"`
	Widgets []models.Widget `json:"widgets"`
}
