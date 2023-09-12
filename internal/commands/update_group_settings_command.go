package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type UpdateGroupSettingsCommand struct {
	EmailToRead        string                               `json:"emailToRead"`
	SubjectLineRegexes []models.SubjectLineRegex            `json:"subjectLineRegexes"`
	EmailWhiteList     []models.GroupSettingsWhiteListEmail `json:"emailWhiteList"`
}

func (command *UpdateGroupSettingsCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
	updateGroupSettingsCommand := UpdateGroupSettingsCommand{}

	bytes, err := utils.GetBodyData(w, r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &updateGroupSettingsCommand)
	if err != nil {
		return err
	}

	command.EmailToRead = updateGroupSettingsCommand.EmailToRead
	command.SubjectLineRegexes = updateGroupSettingsCommand.SubjectLineRegexes
	command.EmailWhiteList = updateGroupSettingsCommand.EmailWhiteList

	return nil
}

func (command UpdateGroupSettingsCommand) Validate() structs.ValidatorError {
	vErr := structs.ValidatorError{
		Errors: make(map[string]string),
	}

	// if len(command.EmailToRead.Email.Email) == 0 {
	// 	vErr.Errors["emailToRead"] = "Email to read is required."
	// }

	return vErr
}

func (command *UpdateGroupSettingsCommand) LoadDataFromRequestAndValidate(w http.ResponseWriter, r *http.Request) (structs.ValidatorError, error) {
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
