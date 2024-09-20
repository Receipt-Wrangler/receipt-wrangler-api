package models

type UserShortcut struct {
	BaseModel
	UserPrefernces   UserPrefernces `json:"-"`
	UserPreferncesId uint           `json:"userPreferncesId"`
	Name             string         `json:"name"`
	Url              string         `json:"url"`
	Icon             string         `json:"icon"`
}
