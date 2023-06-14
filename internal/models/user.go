package models

type User struct {
	BaseModel
	DefaultAvatarColor string   `json:"defaultAvatarColor"`
	DisplayName        string   `json:"displayName"`
	IsDummyUser        bool     `json:"isDummyUser"`
	Password           string   `gorm:"not null"`
	Username           string   `gorm:"not null; uniqueIndex"`
	UserRole           UserRole `gorm:"default:'USER'" json:"userRole"`
}
