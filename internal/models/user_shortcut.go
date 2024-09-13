package models

type UserShortcut struct {
	BaseModel
	UserPreferences   Receipt `json:"-"`
	UserPreferencesId uint    `json:"userPreferencesId"`
	Url               string  `json:"url"`
	Icon              string  `json:"icon"`
}
