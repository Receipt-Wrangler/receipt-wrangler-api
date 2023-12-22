package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/structs"
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

func (command UpsertDashboardCommand) Validate() structs.ValidatorError {
	vErr := structs.ValidatorError{
		Errors: make(map[string]string),
	}

	if len(command.Name) == 0 {
		vErr.Errors["name"] = "Name is required"
	}

	if len(command.GroupId) == 0 {
		vErr.Errors["groupId"] = "Group Id is required"
	}

	return vErr
}

// TODO: cast config somewhere
func (command *UpsertDashboardCommand) LoadDataFromRequestAndValidate(w http.ResponseWriter, r *http.Request) (structs.ValidatorError, error) {
	err := command.LoadDataFromRequest(w, r)
	if err != nil {
		return structs.ValidatorError{}, err
	}

	vErr := command.Validate()
	if len(vErr.Errors) > 0 {
		return vErr, nil
	}

	return structs.ValidatorError{}, nil
}
