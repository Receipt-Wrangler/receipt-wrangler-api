package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type UpsertCategoryCommand struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (tag *UpsertCategoryCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
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

func ValidateCategory(command *UpsertCategoryCommand) structs.ValidatorError {
	errors := make(map[string]string)
	vErr := structs.ValidatorError{}

	if len(command.Name) == 0 {
		errors["name"] = "Name is required"
	}

	vErr.Errors = errors
	return vErr
}
