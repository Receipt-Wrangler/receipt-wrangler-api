package models

import "time"

type ApiKey struct {
	// Not using base model here, so we can have a custom id
	ID              string     `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	CreatedBy       *uint      `json:"createdBy"`
	CreatedByString string     `json:"createdByString"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Hmac            string     `json:"hmac"` // Key format: <prefix>.<ver>.<id>.<secret>
	Version         int        `json:"version"`
	RevokedAt       *time.Time `json:"revokedAt"`
}
