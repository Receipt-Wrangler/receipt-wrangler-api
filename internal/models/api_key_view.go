package models

import "time"

type ApiKeyView struct {
	ID              string     `json:"id"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	CreatedBy       *uint      `json:"createdBy"`
	CreatedByString string     `json:"createdByString"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	UserID          *uint      `json:"userId"`
	Scope           string     `json:"scope"`
	LastUsedAt      *time.Time `json:"lastUsedAt"`
}
