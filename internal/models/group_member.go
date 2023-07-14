package models

import "time"

// Group member
//
// swagger:model
type GroupMember struct {
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// User compound primary key
	//
	// required: true
	UserID uint `gorm:"primaryKey;autoIncrement:false" json:"userId"`

	// Group compound primary key
	//
	// required: true
	GroupID uint `gorm:"primaryKey;autoIncrement:false" json:"groupId"`

	// User's role in group
	//
	// required: true
	GroupRole GroupRole `gorm:"not null" json:"groupRole"`
}
