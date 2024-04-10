package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type UpsertTagCommand struct {
	Id          *uint  `json:"id"`
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

func (tag *UpsertTagCommand) Validate() structs.ValidatorError {
	errors := make(map[string]string)
	vErr := structs.ValidatorError{}

	if len(tag.Name) == 0 {
		errors["name"] = "Name is required"
	}

	vErr.Errors = errors
	return vErr
}
