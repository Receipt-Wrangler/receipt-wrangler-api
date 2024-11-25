package commands

import "receipt-wrangler/api/internal/models"

type UpsertGroupMemberCommand struct {
	UserID    uint             `gorm:"primaryKey;autoIncrement:false" json:"userId"`
	GroupID   uint             `gorm:"primaryKey;autoIncrement:false" json:"groupId"`
	GroupRole models.GroupRole `gorm:"not null" json:"groupRole"`
}
