package models

type Dashboard struct {
	BaseModel
	Name    string   `gorm:"not null;" json:"name"`
	User    User     `json:"-"`
	UserID  uint     `gorm:"not null;" json:"userId"`
	Group   Group    `json:"-"`
	GroupID uint     `json:"groupId"`
	Widgets []Widget `json:"widgets"`
}
