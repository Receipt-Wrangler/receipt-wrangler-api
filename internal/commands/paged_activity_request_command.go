package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type PagedActivityRequestCommand struct {
	PagedRequestCommand
	GroupIds []uint `json:"groupIds"`
}

func (command *PagedActivityRequestCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
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

func (command *PagedActivityRequestCommand) Validate() structs.ValidatorError {
	vErrs := command.PagedRequestCommand.Validate()

	if len(command.GroupIds) == 0 {
		vErrs.Errors["groupIds"] = "Must provide at least one group id"
	}

	return vErrs
}
