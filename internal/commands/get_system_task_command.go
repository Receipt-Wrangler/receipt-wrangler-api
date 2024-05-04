package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type GetSystemTaskCommand struct {
	PagedRequestCommand
	AssociatedEntityId   uint                        `json:"associatedEntityId"`
	AssociatedEntityType models.AssociatedEntityType `json:"associatedEntityType"`
}

func (command *GetSystemTaskCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
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

func (command *GetSystemTaskCommand) Validate() structs.ValidatorError {
	vErrs := command.PagedRequestCommand.Validate()

	associatedEntityIdIsEmpty := command.AssociatedEntityId == 0
	associatedEntityTypeIsEmpty := len(command.AssociatedEntityType) == 0

	if associatedEntityIdIsEmpty && associatedEntityTypeIsEmpty {
		vErrs.Errors["command"] = "Command cannot be empty."
	}

	if !associatedEntityIdIsEmpty && associatedEntityTypeIsEmpty {
		vErrs.Errors["associatedEntityType"] = "Associated entity type cannot be empty."
	}

	return vErrs
}
