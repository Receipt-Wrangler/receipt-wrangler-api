package models

type GroupSettingsEmail struct {
	BaseModel
	Email           string `json:"email"`
	GroupSettingsId uint   `json:"groupSettingsId"`
}
