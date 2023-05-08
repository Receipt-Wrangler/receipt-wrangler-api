package repositories

import (
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
)

// Gets groupMembers that the user has access to
func GetGroupMembersByUserId(userId string) ([]models.GroupMember, error) {
	db := db.GetDB()
	var groupMembers []models.GroupMember

	err := db.Model(models.GroupMember{}).Where("user_id = ?", userId).Find(&groupMembers).Error
	if err != nil {
		return nil, err
	}

	return groupMembers, nil
}

// Gets group ids that the user has access to
func GetGroupIdsByUserId(userId string) ([]uint, error) {
	groupMembers, err := GetGroupMembersByUserId(userId)
	if err != nil {
		return nil, err
	}
	result := make([]uint, len(groupMembers))

	for i := 0; i < len(groupMembers); i++ {
		result[i] = groupMembers[i].GroupID
	}

	return result, nil
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
