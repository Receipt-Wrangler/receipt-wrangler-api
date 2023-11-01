package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/utils"
)

type UpsertTagCommand struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (tag *UpsertTagCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
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
