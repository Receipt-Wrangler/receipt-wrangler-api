package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type GetSystemTaskCommand struct {
	AssociatedEntityId   uint                        `json:"associatedEntityId"`
	AssociatedEntityType models.AssociatedEntityType `json:"associatedEntityType"`
	Count                int                         `json:"count"`
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
	errors := make(map[string]string)
	vErr := structs.ValidatorError{}

	associatedEntityIdIsEmpty := command.AssociatedEntityId == 0
	associatedEntityTypeIsEmpty := len(command.AssociatedEntityType) == 0
	countIsEmpty := command.Count == 0

	if associatedEntityIdIsEmpty && associatedEntityTypeIsEmpty && countIsEmpty {
		errors["command"] = "Command cannot be empty."
	}

	if !associatedEntityIdIsEmpty && associatedEntityTypeIsEmpty {
		errors["associatedEntityType"] = "Associated entity type cannot be empty."
	}

	vErr.Errors = errors
	return vErr
}
