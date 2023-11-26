package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/utils"
)

type UpsertDashboardCommand struct {
	Name    string                `json:"name"`
	GroupId string                `json:"groupId"`  
	Widgets []UpsertWidgetCommand `json:"widgets"`
}

func (command *UpsertDashboardCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
	bytes, err := utils.GetBodyData(w, r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &command)
	if err != nil {
		return err
	}

	return nil
}
