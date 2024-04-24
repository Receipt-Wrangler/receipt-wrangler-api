package models

type SystemSettings struct {
	BaseModel
	EnableLocalSignUp    bool         `json:"enableLocalSignUp;" gorm:"default:false"`
	AiPoweredReceipts    bool         `json:"aiPoweredReceipts;" gorm:"default:false"`
	AiType               AiClientType `json:"aiType"`
	AiUrl                string       `json:"aiUrl"`
	AiKey                string       `json:"aiKey"`
	AiModel              string       `json:"aiModel"`
	NumWorkers           int          `json:"numWorkers;" gorm:"default:3"`
	OcrEngine            OcrEngine    `json:"ocrEngine"`
	EmailPollingInterval int          `json:"emailPollingInterval;" gorm:"default:1800"`
}
