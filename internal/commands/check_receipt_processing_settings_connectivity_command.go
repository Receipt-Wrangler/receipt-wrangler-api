package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type CheckReceiptProcessingSettingsCommand struct {
	ID uint `json:"id"`
	UpsertReceiptProcessingSettingsCommand
}

func (command *CheckReceiptProcessingSettingsCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
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

func (command *CheckReceiptProcessingSettingsCommand) Validate() structs.ValidatorError {
	vErrs := structs.ValidatorError{}
	errors := map[string]string{}
	vErrs.Errors = errors

	idIsEmpty := command.ID == 0
	settingsEmpty := command.UpsertReceiptProcessingSettingsCommand.IsEmpty()

	if idIsEmpty && settingsEmpty {
		vErrs.Errors["command"] = "Command and ID cannot be empty."
		return vErrs
	}

	if !settingsEmpty {
		settingsErrors := command.UpsertReceiptProcessingSettingsCommand.Validate(false)
		for k, v := range settingsErrors.Errors {
			vErrs.Errors[k] = v
		}
	}

	if !idIsEmpty {
		if command.ID < 1 {
			errors["id"] = "id must be greater than 0"
		}
	}

	return vErrs
}
