package models

type SystemSettings struct {
	BaseModel
	EnableLocalSignUp    bool         `json:"enableLocalSignUp" default:"false"`
	AiPoweredReceipts    bool         `json:"aiPoweredReceipts" default:"false"`
	AiType               AiClientType `json:"aiType"`
	AiUrl                string       `json:"aiUrl"`
	AiKey                string       `json:"aiKey"`
	AiModel              string       `json:"aiModel"`
	NumWorkers           int          `json:"numWorkers" default:"1"`
	OcrEngine            OcrEngine    `json:"ocrEngine"`
	EmailPollingInterval int          `json:"emailPollingInterval" default:"1800"`
}
