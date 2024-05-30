package commands

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type ImportType string

const (
	ImportConfig ImportType = "IMPORT_CONFIG"
)

func (importType *ImportType) Scan(value string) error {
	*importType = ImportType(value)
	return nil
}

func (importType ImportType) Value() (driver.Value, error) {
	if len(importType) == 0 {
		return "", nil
	}

	if importType != ImportConfig {
		return nil, errors.New("invalid import type")
	}
	return string(importType), nil
}

type ImportCommand struct {
	ImportType ImportType `json:"importType"`
	ConfigImportCommand
}

func (command *ImportCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
	bytes, err := utils.GetBodyData(w, r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &command)
	if err != nil {
		return err
	}

	if command.ImportType == ImportConfig {
		command.ConfigImportCommand = ConfigImportCommand{}

		err = command.ConfigImportCommand.LoadDataFromRequest(w, r)
		if err != nil {
			return err
		}
	}

	return nil
}

func (command *ImportCommand) Validate() structs.ValidatorError {
	vErr := structs.ValidatorError{
		Errors: make(map[string]string),
	}

	if len(command.ImportType) == 0 {
		vErr.Errors["importType"] = "Import Type is required."
	}

	if command.ImportType == ImportConfig {
		configImportCommandErr := command.ConfigImportCommand.Validate()
		for key, value := range configImportCommandErr.Errors {
			vErr.Errors[key] = value
		}
	}

	return vErr
}
