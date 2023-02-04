package models

type User struct {
	BaseModel
	Username    string `gorm:"not null; uniqueIndex"`
	Password    string `gorm:"not null"`
	DisplayName string `json:"displayName"`
	UserRole UserRole `gorm:"default:'USER'" json:"userRole"` 
}
