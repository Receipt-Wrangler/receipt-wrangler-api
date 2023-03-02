package repositories

import (
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
)

func GetGroupMembersByUserId(userId string) ([]models.GroupMember, error) {
	db := db.GetDB()
	var groupMembers []models.GroupMember

	err := db.Model(models.GroupMember{}).Where("user_id = ?", userId).Find(&groupMembers).Error
	if err != nil {
		return nil, err
	}

	return groupMembers, nil
}

func GetGroupMemberByUserIdAndGroupId(userId string, groupId string) (models.GroupMember, error) {
	db := db.GetDB()
	var groupMember models.GroupMember

	err := db.Model(models.GroupMember{}).Where("user_id = ? AND group_id = ?", userId, groupId).Find(&groupMember).Error
	if err != nil {
		return models.GroupMember{}, err
	}

	return groupMember, nil
}
