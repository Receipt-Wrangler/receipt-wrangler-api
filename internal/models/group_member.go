package models

import "time"

type GroupMember struct {
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Name      string    `gorm:"not null" json:"name"`
	UserID    uint64    `gorm:"primaryKey;autoIncrement:false" json:"userId"`
	GroupID   uint64    `gorm:"primaryKey;autoIncrement:false" json:"groupId"`
}
