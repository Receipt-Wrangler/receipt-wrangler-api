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
	SystemEmailId               *uint                                `json:"SystemEmailId"`
	EmailIntegrationEnabled     bool                                 `json:"emailIntegrationEnabled"`
	SubjectLineRegexes          []models.SubjectLineRegex            `json:"subjectLineRegexes"`
	EmailWhiteList              []models.GroupSettingsWhiteListEmail `json:"emailWhiteList"`
	EmailDefaultReceiptStatus   models.ReceiptStatus                 `json:"emailDefaultReceiptStatus"`
	EmailDefaultReceiptPaidBy   *models.User                         `json:"-"`
	EmailDefaultReceiptPaidById *uint                                `json:"emailDefaultReceiptPaidById"`
	PromptId                    *uint                                `json:"promptId"`
	FallbackPromptId            *uint                                `json:"fallbackPromptId"`
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
	command.SystemEmailId = updateGroupSettingsCommand.SystemEmailId
	command.SubjectLineRegexes = updateGroupSettingsCommand.SubjectLineRegexes
	command.EmailWhiteList = updateGroupSettingsCommand.EmailWhiteList
	command.EmailDefaultReceiptStatus = updateGroupSettingsCommand.EmailDefaultReceiptStatus
	command.EmailDefaultReceiptPaidById = updateGroupSettingsCommand.EmailDefaultReceiptPaidById
	command.PromptId = updateGroupSettingsCommand.PromptId
	command.FallbackPromptId = updateGroupSettingsCommand.FallbackPromptId

	return nil
}

func (command UpdateGroupSettingsCommand) Validate() structs.ValidatorError {
	vErr := structs.ValidatorError{
		Errors: make(map[string]string),
	}

	if command.SystemEmailId == nil && command.EmailIntegrationEnabled {
		vErr.Errors["systemEmailId"] = "System email is required"
	}

	if command.EmailDefaultReceiptStatus == "" && command.EmailIntegrationEnabled {
		vErr.Errors["emailDefaultReceiptStatus"] = "Default receipt status is required when email integration is enabled"
	}

	if (command.EmailDefaultReceiptPaidById == nil || *command.EmailDefaultReceiptPaidById == 0) && command.EmailIntegrationEnabled {
		vErr.Errors["emailDefaultReceiptPaidById"] = "Default receipt paid by is required when email integration is enabled"
	}

	if command.PromptId != nil && *command.PromptId < 1 {
		vErr.Errors["promptId"] = "PromptId must be greater than 0"
	}

	if command.FallbackPromptId != nil && *command.FallbackPromptId < 1 {
		vErr.Errors["fallbackPromptId"] = "FallbackPromptId must be greater than 0"
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
