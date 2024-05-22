package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type UpsertPromptCommand struct {
	Name        string `gorm:"not null; uniqueIndex" json:"name"`
	Description string `json:"description"`
	Prompt      string `json:"prompt"`
}

func (command *UpsertPromptCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
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

func (command *UpsertPromptCommand) Validate() structs.ValidatorError {
	vErr := structs.ValidatorError{}
	errorMap := make(map[string]string)

	if len(command.Name) == 0 {
		errorMap["name"] = "Name cannot be empty"
	}

	if len(command.Prompt) == 0 {
		errorMap["prompt"] = "Prompt cannot be empty"
	} else {
		regex := utils.GetTriggerRegex()
		templateVariables := regex.FindAllString(command.Prompt, -1)
		for i := 0; i < len(templateVariables); i++ {
			variable := templateVariables[i]
			if variable != string(structs.CATEGORIES) && variable != string(structs.TAGS) && variable != string(structs.OCR_TEXT) && variable != string(structs.CURRENT_YEAR) {
				errorMap["prompt"] = "Invalid template variables found"
			}
		}
	}

	vErr.Errors = errorMap
	return vErr
}
