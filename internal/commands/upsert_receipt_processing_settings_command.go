package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type UpsertReceiptProcessingSettingsCommand struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	AiType      models.AiClientType `json:"aiType"`
	Url         string              `json:"url"`
	Key         string              `json:"key"`
	Model       string              `json:"model"`
	NumWorkers  int                 `json:"numWorkers"`
	OcrEngine   models.OcrEngine    `json:"ocrEngine"`
	PromptId    uint                `json:"promptId"`
}

func (command *UpsertReceiptProcessingSettingsCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
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

func (command *UpsertReceiptProcessingSettingsCommand) Validate() structs.ValidatorError {
	vErrs := structs.ValidatorError{}
	errors := map[string]string{}
	vErrs.Errors = errors

	if len(command.Name) == 0 {
		errors["name"] = "name is required"
		return vErrs
	}

	if len(command.OcrEngine) == 0 {
		errors["ocrEngine"] = "ocrEngine is required"
		return vErrs
	}

	if command.PromptId < 1 {
		errors["promptId"] = "promptId must be greater than 0"
	}

	if command.NumWorkers < 1 {
		errors["numWorkers"] = "numWorkers must be greater than 0"
	}

	if len(command.AiType) == 0 {
		errors["type"] = "type is required"
		return vErrs
	}

	if command.AiType == models.OPEN_AI || command.AiType == models.GEMINI {
		if command.Key == "" {
			errors["key"] = "key is required"
		}

		if command.Url != "" {
			errors["url"] = "url is not required"
		}
	}

	if command.AiType == models.OPEN_AI_CUSTOM {
		if len(command.Url) == 0 {
			errors["url"] = "url is required"
		}
	}

	return vErrs
}

func (command *UpsertReceiptProcessingSettingsCommand) IsEmpty() bool {
	return command.Name == "" && command.Description == "" && command.AiType == "" && command.Url == "" && command.Key == "" && command.Model == "" && command.NumWorkers == 0 && command.OcrEngine == "" && command.PromptId == 0
}
