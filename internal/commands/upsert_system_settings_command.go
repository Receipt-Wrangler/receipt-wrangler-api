package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type UpsertSystemSettingsCommand struct {
	EnableLocalSignUp                   bool  `json:"enableLocalSignUp"`
	EmailPollingInterval                int   `json:"emailPollingInterval"`
	ReceiptProcessingSettingsId         *uint `json:"receiptProcessingSettingsId"`
	FallbackReceiptProcessingSettingsId *uint `json:"fallbackReceiptProcessingSettingsId"`
}

func (command *UpsertSystemSettingsCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
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

func (command *UpsertSystemSettingsCommand) Validate() structs.ValidatorError {
	vErr := structs.ValidatorError{}
	errorMap := make(map[string]string)
	vErr.Errors = errorMap

	if command.EmailPollingInterval < 0 {
		errorMap["emailPollingInterval"] = "Email polling interval must be greater than 0"
	}

	if command.ReceiptProcessingSettingsId != nil && *command.ReceiptProcessingSettingsId <= 0 {
		errorMap["receiptProcessingSettingsId"] = "Invalid receipt processing settings ID"
	}

	if command.FallbackReceiptProcessingSettingsId != nil && *command.FallbackReceiptProcessingSettingsId <= 0 {
		errorMap["fallbackReceiptProcessingSettingsId"] = "Invalid fallback receipt processing settings ID"
	}

	if command.ReceiptProcessingSettingsId == nil && command.FallbackReceiptProcessingSettingsId != nil {
		errorMap["fallbackReceiptProcessingSettingsId"] = "Fallback receipt processing settings ID cannot be set without receipt processing settings ID"
	}

	if command.ReceiptProcessingSettingsId != nil &&
		command.FallbackReceiptProcessingSettingsId != nil &&
		*command.ReceiptProcessingSettingsId ==
			*command.FallbackReceiptProcessingSettingsId {
		errorMap["fallbackReceiptProcessingSettingsId"] = "Fallback receipt processing settings ID cannot be the same as receipt processing settings ID"
	}

	return vErr
}
