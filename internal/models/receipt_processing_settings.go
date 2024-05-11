package models

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/utils"
)

type ReceiptProcessingSettings struct {
	BaseModel
	Name        string       `gorm:"not null; uniqueIndex" json:"name"`
	Description string       `json:"description"`
	AiType      AiClientType `json:"type"`
	Url         string       `json:"url"`
	Key         string       `json:"key"`
	Model       string       `json:"model"`
	NumWorkers  int          `json:"numWorkers"`
	OcrEngine   OcrEngine    `json:"ocrEngine"`
	Prompt      Prompt       `json:"prompt"`
	PromptId    uint         `json:"promptId"`
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
