package commands

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type PagedApiKeyRequestCommand struct {
	PagedRequestCommand
	ApiKeyFilter `json:"filter"`
}

func (command *PagedApiKeyRequestCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
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

func (command *PagedApiKeyRequestCommand) Validate(r *http.Request) structs.ValidatorError {
	vErrs := command.PagedRequestCommand.Validate()
	token := structs.GetClaims(r)

	if command.ApiKeyFilter.AssociatedApiKeys == ASSOCIATED_API_KEYS_ALL &&
		token.UserRole != models.ADMIN {
		vErrs.Errors["associatedApiKeys"] = "Must be an admin to view all API keys"
	}

	if command.ApiKeyFilter.AssociatedApiKeys == "" {
		vErrs.Errors["associatedApiKeys"] = "Associated API keys is required"
	}

	return vErrs
}

type ApiKeyFilter struct {
	AssociatedApiKeys AssociatedApiKeys `json:"associatedApiKeys"`
}

type AssociatedApiKeys string

const (
	ASSOCIATED_API_KEYS_MINE AssociatedApiKeys = "MINE"
	ASSOCIATED_API_KEYS_ALL  AssociatedApiKeys = "ALL"
)

func (associatedApiKeys *AssociatedApiKeys) Scan(value string) error {
	*associatedApiKeys = AssociatedApiKeys(value)
	return nil
}

func (associatedApiKeys AssociatedApiKeys) Value() (driver.Value, error) {
	if associatedApiKeys != ASSOCIATED_API_KEYS_MINE && associatedApiKeys != ASSOCIATED_API_KEYS_ALL {
		return nil, errors.New("invalid associatedApiKeys")
	}
	return string(associatedApiKeys), nil
}