package models

type GroupSettings struct {
	BaseModel
	GroupId                     uint                          `gorm:"not null;unique" json:"groupId"`
	EmailIntegrationEnabled     bool                          `gorm:"not null; default:false" json:"emailIntegrationEnabled"`
	EmailBodyProcessingEnabled  bool                          `gorm:"not null; default:false" json:"emailBodyProcessingEnabled"`
	SystemEmailId               *uint                         `json:"systemEmailId"`
	SystemEmail                 SystemEmail                   `json:"systemEmail"`
	EmailToRead                 string                        `json:"emailToRead"`
	SubjectLineRegexes          []SubjectLineRegex            `json:"subjectLineRegexes"`
	EmailWhiteList              []GroupSettingsWhiteListEmail `json:"emailWhiteList"`
	EmailDefaultReceiptStatus   ReceiptStatus                 `json:"emailDefaultReceiptStatus"`
	EmailDefaultReceiptPaidBy   *User                         `json:"-"`
	EmailDefaultReceiptPaidById *uint                         `json:"emailDefaultReceiptPaidById"`
	PromptId                    *uint                         `json:"promptId"`
	Prompt                      *Prompt                       `json:"prompt"`
	FallbackPromptId            *uint                         `json:"fallbackPromptId"`
	FallbackPrompt              *Prompt                       `json:"fallbackPrompt"`
	// EmailRequireSignedSender makes the poller accept an email only when it
	// carries a valid DKIM signature from its From domain. AutoMigrate backfills
	// existing rows with the default. Because the column has a default, GORM
	// omits a false value on Create, so no create path may rely on setting it
	// false; updates write it explicitly.
	EmailRequireSignedSender bool `gorm:"not null; default:true" json:"emailRequireSignedSender"`
}
