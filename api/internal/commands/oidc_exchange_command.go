package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strings"
)

// OidcExchangeCommand redeems the one-time code the mobile app received on its
// private-use URL scheme for a real session.
type OidcExchangeCommand struct {
	Code string `json:"code"`
	// CodeVerifier is the PKCE verifier the app generated BEFORE opening the
	// browser and never let out of the process. It is what proves this is the same
	// app that started the flow, rather than another app that registered the same
	// URL scheme and intercepted the redirect.
	CodeVerifier string `json:"codeVerifier"`
}

func (command *OidcExchangeCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
	bytes, err := utils.GetBodyData(w, r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &command)
	if err != nil {
		return err
	}

	command.Code = strings.TrimSpace(command.Code)
	command.CodeVerifier = strings.TrimSpace(command.CodeVerifier)

	return nil
}

func (command *OidcExchangeCommand) Validate() structs.ValidatorError {
	errors := make(map[string]string)
	vErr := structs.ValidatorError{}

	if len(command.Code) == 0 {
		errors["code"] = "Code is required"
	}

	if len(command.CodeVerifier) == 0 {
		errors["codeVerifier"] = "Code verifier is required"
	}

	vErr.Errors = errors
	return vErr
}
