package models

type SystemSettings struct {
	BaseModel
	EnableLocalSignUp                   bool                      `json:"enableLocalSignUp" gorm:"default:false"`
	DebugOcr                            bool                      `json:"debugOcr" gorm:"default:false"`
	NumWorkers                          int                       `json:"numWorkers"`
	EmailPollingInterval                int                       `json:"emailPollingInterval" gorm:"default:1800"`
	CurrencyDisplay                     string                    `json:"currencyDisplay" gorm:"default:$"`
	CurrencyThousandthsSeparator        string                    `json:"currencyThousandthsSeparator"`
	CurrencyDecimalSeparator            string                    `json:"currencyDecimalSeparator"`
	CurrencySymbolPosition              string                    `json:"currencySymbolPosition"`
	ReceiptProcessingSettings           ReceiptProcessingSettings `json:"-"`
	ReceiptProcessingSettingsId         *uint                     `json:"receiptProcessingSettingsId"`
	FallbackReceiptProcessingSettings   ReceiptProcessingSettings `json:"-"`
	FallbackReceiptProcessingSettingsId *uint                     `json:"fallbackReceiptProcessingSettingsId"`
}
