package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type UpsertSystemEmailCommand struct {
	Host        string `json:"host"`
	Port        string `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	UseStartTLS bool   `json:"useStartTLS"`
}

func (command *UpsertSystemEmailCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
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

func (command *UpsertSystemEmailCommand) Validate(isCreate bool) structs.ValidatorError {
	errors := make(map[string]string)
	vErr := structs.ValidatorError{}

	if len(command.Host) == 0 {
		errors["host"] = "Host is required"
	}

	if len(command.Port) == 0 {
		errors["port"] = "Port is required"
	}

	if len(command.Username) == 0 {
		errors["username"] = "Username is required"
	}

	if isCreate {
		if len(command.Password) == 0 {
			errors["password"] = "Password is required"
		}

	}

	vErr.Errors = errors
	return vErr
}
