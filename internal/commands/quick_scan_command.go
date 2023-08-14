package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type QuickScanCommand struct {
	models.FileData
	GroupId uint               `json:"groupId"`
	Group   models.Group       `json:"-"`
	Status  models.GroupStatus `json:"groupStatus"`
}

func (command *QuickScanCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
	quickScanCommand := QuickScanCommand{}

	bytes, err := utils.GetBodyData(w, r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &quickScanCommand)
	if err != nil {
		return err
	}

	command.FileData = quickScanCommand.FileData
	command.GroupId = quickScanCommand.GroupId
	command.Status = quickScanCommand.Status

	return nil
}

func (command QuickScanCommand) Validate() structs.ValidatorError {
	vErr := structs.ValidatorError{
		Errors: make(map[string]string),
	}

	if command.GroupId == 0 {
		vErr.Errors["groupId"] = "Group Id is required."
	}

	if len(command.Status) == 0 {
		vErr.Errors["status"] = "Status is required."
	}

	if len(command.Name) == 0 {
		vErr.Errors["filename"] = "Filename is required."
	}

	if len(command.ImageData) == 0 {
		vErr.Errors["imageData"] = "Image data is required."
	}

	return vErr
}

func (command *QuickScanCommand) LoadDataFromRequestAndValidate(w http.ResponseWriter, r *http.Request) (structs.ValidatorError, error) {
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
