package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type UpdateGroupSettingsCommand struct {
	EmailToRead             string                               `json:"emailToRead"`
	EmailIntegrationEnabled bool                                 `json:"emailIntegrationEnabled"`
	SubjectLineRegexes      []models.SubjectLineRegex            `json:"subjectLineRegexes"`
	EmailWhiteList          []models.GroupSettingsWhiteListEmail `json:"emailWhiteList"`
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

	command.EmailIntegrationEnabled = updateGroupSettingsCommand.EmailIntegrationEnabled
	command.EmailToRead = updateGroupSettingsCommand.EmailToRead
	command.SubjectLineRegexes = updateGroupSettingsCommand.SubjectLineRegexes
	command.EmailWhiteList = updateGroupSettingsCommand.EmailWhiteList

	return nil
}

func (command UpdateGroupSettingsCommand) Validate() structs.ValidatorError {
	vErr := structs.ValidatorError{
		Errors: make(map[string]string),
	}

	if command.EmailToRead == "" && command.EmailIntegrationEnabled {
		vErr.Errors["emailToRead"] = "Email to read is required when email integration is enabled"
	}

	_, err := mail.ParseAddress(command.EmailToRead)
	if err != nil {
		vErr.Errors["emailToRead"] = "Email to read is invalid"
	}

	for index, email := range command.EmailWhiteList {
		_, err := mail.ParseAddress(email.Email)
		if err != nil {
			errorKey := fmt.Sprintf("emailWhiteList.%d.email", index)
			vErr.Errors[errorKey] = "Email is an  invalid email"
		}
	}

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
