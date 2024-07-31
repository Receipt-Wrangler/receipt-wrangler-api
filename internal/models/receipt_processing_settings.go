package models

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/utils"
)

type ReceiptProcessingSettings struct {
	BaseModel
	Name          string       `gorm:"not null; uniqueIndex" json:"name"`
	Description   string       `json:"description"`
	AiType        AiClientType `json:"aiType"`
	Url           string       `json:"url"`
	Key           string       `json:"-"`
	Model         string       `json:"model"`
	OcrEngine     *OcrEngine   `json:"ocrEngine"`
	Prompt        Prompt       `json:"prompt"`
	PromptId      uint         `json:"promptId"`
	IsVisionModel bool         `json:"isVisionModel"`
}

func (ReceiptProcessingSettings *ReceiptProcessingSettings) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
	bytes, err := utils.GetBodyData(w, r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &ReceiptProcessingSettings)
	if err != nil {
		return err
	}

	return nil
}
