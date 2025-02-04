package models

import "time"

type RefreshToken struct {
	BaseModel
	UserId    uint      `gorm:"not null"`
	Token     string    `gorm:"not null"`
	IsUsed    bool      `gorm:"default:false"`
	ExpiresAt time.Time `json:"expiryDate"`
}
