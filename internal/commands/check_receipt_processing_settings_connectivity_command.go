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
	vErrs := command.UpsertReceiptProcessingSettingsCommand.Validate()

	idIsEmpty := command.ID == 0
	commandIsEmpty := command.IsEmpty()

	if idIsEmpty && commandIsEmpty {
		vErrs.Errors["command"] = "Command and ID cannot be empty."
	}

	return vErrs
}
