package models

type SystemSettings struct {
	BaseModel
	EnableLocalSignUp                     bool                      `json:"enableLocalSignUp" gorm:"default:false"`
	DebugOcr                              bool                      `json:"debugOcr" gorm:"default:false"`
	NumWorkers                            int                       `json:"numWorkers"`
	EmailPollingInterval                  int                       `json:"emailPollingInterval" gorm:"default:1800"`
	CurrencyDisplay                       string                    `json:"currencyDisplay" gorm:"default:$"`
	CurrencyThousandthsSeparator          CurrencySeparator         `json:"currencyThousandthsSeparator" gorm:"default:,"`
	CurrencyDecimalSeparator              CurrencySeparator         `json:"currencyDecimalSeparator" gorm:"default:."`
	CurrencySymbolPosition                CurrencySymbolPosition    `json:"currencySymbolPosition" gorm:"default:START"`
	CurrencyHideDecimalPlaces             bool                      `json:"currencyHideDecimalPlaces" gorm:"default:false"`
	ReceiptProcessingSettings             ReceiptProcessingSettings `json:"-"`
	ReceiptProcessingSettingsId           *uint                     `json:"receiptProcessingSettingsId"`
	FallbackReceiptProcessingSettings     ReceiptProcessingSettings `json:"-"`
	FallbackReceiptProcessingSettingsId   *uint                     `json:"fallbackReceiptProcessingSettingsId"`
	AsynqConcurrency                      int                       `json:"asynqConcurrency" gorm:"default:10"`
	AsynqQuickScanPriority                int                       `json:"asynqQuickScanPriority" gorm:"default:4"`
	AsynqEmailReceiptProcessingPriority   int                       `json:"asynqEmailReceiptProcessingPriority" gorm:"default:3"`
	AsynqEmailPollingPriority             int                       `json:"asynqEmailPollingPriority" gorm:"default:2"`
	AsynqEmailReceiptImageCleanupPriority int                       `json:"asynqEmailReceiptImageCleanupPriority" gorm:"default:1"`
}
