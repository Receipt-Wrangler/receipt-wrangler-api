package models

type SystemSettings struct {
	BaseModel
	EnableLocalSignUp                   bool                      `json:"enableLocalSignUp" gorm:"default:false"`
	DebugOcr                            bool                      `json:"debugOcr" gorm:"default:false"`
	NumWorkers                          int                       `json:"numWorkers"`
	EmailPollingInterval                int                       `json:"emailPollingInterval" gorm:"default:1800"`
	CurrencyLocale                      string                    `json:"currencyLocale" gorm:"default:en-US"`
	CurrencyCode                        string                    `json:"currencyCode" gorm:"default:USD"`
	ShowCurrencySymbol                  bool                      `json:"showCurrencySymbol"`
	CurrencyDisplay                     string                    `json:"currencyDisplay" gorm:"default:$"`
	ReceiptProcessingSettings           ReceiptProcessingSettings `json:"-"`
	ReceiptProcessingSettingsId         *uint                     `json:"receiptProcessingSettingsId"`
	FallbackReceiptProcessingSettings   ReceiptProcessingSettings `json:"-"`
	FallbackReceiptProcessingSettingsId *uint                     `json:"fallbackReceiptProcessingSettingsId"`
}
