package models

import "time"

type GroupMember struct {
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	UserID    uint      `gorm:"primaryKey;autoIncrement:false" json:"userId"`
	GroupID   uint      `gorm:"primaryKey;autoIncrement:false" json:"groupId"`
	GroupRole GroupRole `gorm:"not null" json:"groupRole"`
}
