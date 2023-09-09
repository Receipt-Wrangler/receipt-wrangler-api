package models

type SubjectLineRegex struct {
	BaseModel
	Regex           string `gorm:"not null" json:"regex"`
	GroupSettingsId uint   `json:"groupSettingsId"`
}
