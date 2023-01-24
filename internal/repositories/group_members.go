package repositories

import (
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
)

func GetGroupMembersByUserId(userId uint) ([]models.GroupMember, error) {
	db := db.GetDB()
	var groupMembers []models.GroupMember

	err := db.Model(models.GroupMember{}).Where("user_id = ?", userId).Find(&groupMembers).Error
	if err != nil {
		return nil, err
	}

	return groupMembers, nil
}
