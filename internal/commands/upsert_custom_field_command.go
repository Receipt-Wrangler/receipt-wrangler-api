package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type UpsertCustomFieldCommand struct {
	Name        string                           `json:"name"`
	Type        models.CustomFieldType           `json:"type"`
	Description string                           `json:"description"`
	Options     []UpsertCustomFieldOptionCommand `json:"options"`
}

func (command *UpsertCustomFieldCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
	bytes, err := utils.GetBodyData(w, r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &command)
	if err != nil {
		return err
	}

	if command.Type != models.SELECT && len(command.Options) > 0 {
		command.Options = []UpsertCustomFieldOptionCommand{}
	}

	return nil
}

func (command *UpsertCustomFieldCommand) Validate() structs.ValidatorError {
	errors := make(map[string]string)
	vErr := structs.ValidatorError{}

	if len(command.Name) == 0 {
		errors["name"] = "Name is required"
	}

	if len(command.Type) == 0 {
		errors["type"] = "Type is required"
	}

	if command.Type == models.SELECT && len(command.Options) == 0 {
		errors["options"] = "Options are required"
	}

	vErr.Errors = errors
	return vErr
}
