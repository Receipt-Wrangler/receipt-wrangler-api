package models

type Notification struct {
	BaseModel
	Type   NotificationType `gorm:"not null" json:"type"`
	Title  string           `gorm:"not null;" json:"title"`
	Body   string           `gorm:"not null;" json:"body"`
	UserId uint             `json:"userId"`
	User   User             `json:"-"`
}
