package models

type GroupSettingsWhiteListEmail struct {
	BaseModel
	Email           string `json:"email"`
	GroupSettingsId uint   `json:"groupSettingsId"`
}
