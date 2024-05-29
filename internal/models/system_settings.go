package models

type SystemSettings struct {
	BaseModel
	EnableLocalSignUp                   bool                      `json:"enableLocalSignUp" gorm:"default:false"`
	DebugOcr                            bool                      `json:"debugOcr" gorm:"default:false"`
	EmailPollingInterval                int                       `json:"emailPollingInterval" gorm:"default:1800"`
	ReceiptProcessingSettings           ReceiptProcessingSettings `json:"-"`
	ReceiptProcessingSettingsId         *uint                     `json:"receiptProcessingSettingsId"`
	FallbackReceiptProcessingSettings   ReceiptProcessingSettings `json:"-"`
	FallbackReceiptProcessingSettingsId *uint                     `json:"fallbackReceiptProcessingSettingsId"`
}
