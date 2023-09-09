package models

type GroupSettingsEmail struct {
	BaseModel
	Email
	GroupSettingsId uint `json:"groupSettingsId"`
}
