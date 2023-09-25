package models

import "time"

type BaseModel struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	CreatedBy       *uint     `json:"createdBy"`
	CreatedByString string    `json:"createdByString"`
}
