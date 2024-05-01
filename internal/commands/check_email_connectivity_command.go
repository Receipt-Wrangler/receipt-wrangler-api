package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type CheckEmailConnectivityCommand struct {
	ID uint `json:"id"`
	UpsertSystemEmailCommand
}

func (command *CheckEmailConnectivityCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
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

func (command *CheckEmailConnectivityCommand) Validate() structs.ValidatorError {
	errors := make(map[string]string)
	vErr := structs.ValidatorError{}

	idIsEmpty := command.ID == 0
	hostIsEmpty := len(command.Host) == 0
	portIsEmpty := len(command.Port) == 0
	usernameIsEmpty := len(command.Username) == 0
	passwordIsEmpty := len(command.Password) == 0

	if idIsEmpty && (hostIsEmpty || portIsEmpty || usernameIsEmpty || passwordIsEmpty) {
		errors["command"] = "If ID is not provided, full credentials must be provided"
	}

	if idIsEmpty && hostIsEmpty && portIsEmpty && usernameIsEmpty && passwordIsEmpty {
		errors["command"] = "Command cannot be empty."
	}

	vErr.Errors = errors
	return vErr
}
