package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type UpsertSystemSettingsCommand struct {
	EnableLocalSignUp                   bool                                  `json:"enableLocalSignUp"`
	DebugOcr                            bool                                  `json:"debugOcr"`
	CurrencyDisplay                     string                                `json:"currencyDisplay"`
	CurrencyThousandthsSeparator        models.CurrencySeparator              `json:"currencyThousandthsSeparator"`
	CurrencyDecimalSeparator            models.CurrencySeparator              `json:"currencyDecimalSeparator"`
	CurrencySymbolPosition              models.CurrencySymbolPosition         `json:"currencySymbolPosition"`
	CurrencyHideDecimalPlaces           bool                                  `json:"currencyHideDecimalPlaces"`
	NumWorkers                          int                                   `json:"numWorkers"`
	EmailPollingInterval                int                                   `json:"emailPollingInterval"`
	ReceiptProcessingSettingsId         *uint                                 `json:"receiptProcessingSettingsId"`
	FallbackReceiptProcessingSettingsId *uint                                 `json:"fallbackReceiptProcessingSettingsId"`
	TaskConcurrency                     int                                   `json:"taskConcurrency"`
	TaskQueueConfigurations             []UpsertTaskQueueConfigurationCommand `json:"taskQueueConfigurations"`
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

	if len(command.CurrencySymbolPosition) == 0 {
		errorMap["currencySymbolPosition"] = "Currency symbol position is required"
	}

	if len(command.CurrencyThousandthsSeparator) == 0 {
		errorMap["currencyThousandthsSeparator"] = "Currency thousandths separator is required"
	}

	if len(command.CurrencyDecimalSeparator) == 0 {
		errorMap["currencyDecimalSeparator"] = "Currency decimal separator is required"
	}

	if command.TaskConcurrency < 0 {
		errorMap["taskConcurrency"] = "Task concurrency must be greater than or equal to 0"
	}

	queueNames := models.GetQueueNames()
	if len(command.TaskQueueConfigurations) != len(queueNames) {
		errorMap["taskQueueConfigurations"] = "Task queue configurations must be provided for all queues"
	}

	return vErr
}

func (command *UpsertSystemSettingsCommand) ToSystemSettings(id uint) (models.SystemSettings, error) {
	var systemSettings models.SystemSettings

	bytes, err := json.Marshal(command)
	if err != nil {
		return systemSettings, err
	}

	err = json.Unmarshal(bytes, &systemSettings)
	if err != nil {
		return systemSettings, err
	}

	systemSettings.ID = id
	for _, config := range systemSettings.TaskQueueConfigurations {
		config.SystemSettingsId = id
	}

	return systemSettings, nil
}
