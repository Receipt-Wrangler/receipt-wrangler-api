package models

type Group struct {
	BaseModel
	Name           string      `gorm:"not null"`
	IsDefaultGroup bool        `json:"isDefault"`
	GroupMembers   GroupMember `json:"groupMembers"`
}
