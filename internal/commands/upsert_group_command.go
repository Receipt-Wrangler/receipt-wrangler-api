package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type UpsertGroupCommand struct {
	Name         string               `gorm:"not null" json:"name"`
	GroupMembers []models.GroupMember `json:"groupMembers"`
	Status       models.GroupStatus   `gorm:"default:'ACTIVE'; not null" json:"status"`
}

func (command *UpsertGroupCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
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

func (command *UpsertGroupCommand) Validate() structs.ValidatorError {
	vErr := structs.ValidatorError{}
	errorMap := make(map[string]string)

	if len(command.Name) == 0 {
		errorMap["name"] = "Name is required"
	}

	if len(command.Status) == 0 {
		errorMap["status"] = "Status is required"
	}

	if len(command.GroupMembers) == 0 {
		errorMap["groupMembers"] = "Group Members are required"
	}

	vErr.Errors = errorMap
	return vErr
}
