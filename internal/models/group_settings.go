package models

type GroupSettings struct {
	BaseModel
	GroupId                 uint                 `gorm:"not null;unique" json:"groupId"`
	EmailIntegrationEnabled bool                 `gorm:"not null; default:false" json:"emailIntegrationEnabled"`
	EmailToRead             GroupSettingsEmail   `json:"emailToRead"`
	SubjectLineRegexes      []SubjectLineRegex   `json:"subjectLineRegexes"`
	EmailWhiteList          []GroupSettingsEmail `json:"emailWhiteList"`
}
