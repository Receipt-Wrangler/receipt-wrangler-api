package models

type GroupSettings struct {
	BaseModel
	GroupId                     uint                          `gorm:"not null;unique" json:"groupId"`
	EmailIntegrationEnabled     bool                          `gorm:"not null; default:false" json:"emailIntegrationEnabled"`
	EmailToRead                 string                        `json:"emailToRead"`
	SubjectLineRegexes          []SubjectLineRegex            `json:"subjectLineRegexes"`
	EmailWhiteList              []GroupSettingsWhiteListEmail `json:"emailWhiteList"`
	EmailDefaultReceiptStatus   ReceiptStatus                 `json:"emailDefaultReceiptStatus"`
	EmailDefaultReceiptPaidBy   *User                         `json:"-"`
	EmailDefaultReceiptPaidById *uint                         `json:"emailDefaultReceiptPaidById"`
}
