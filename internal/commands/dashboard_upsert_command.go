package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
)

type UpsertDashboardCommand struct {
	Name    string          `json:"name"`
	Widgets []models.Widget `json:"widgets"`
}

func (tag *UpsertDashboardCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
	bytes, err := utils.GetBodyData(w, r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &tag)
	if err != nil {
		return err
	}

	return nil
}
