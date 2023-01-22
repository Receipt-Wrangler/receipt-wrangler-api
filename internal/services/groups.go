package services

import (
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
)

func GetGroupsForUser(userId uint) ([]models.Group, error) {
	db := db.GetDB()
	var groups []models.Group

	groupMembers, err := repositories.GetGroupMembersByUserId(userId)
	if err != nil {
		return nil, err
	}

	groupIds := make([]uint, len(groupMembers))
	for i := 0; i < len(groupMembers); i++ {
		groupIds[i] = groupMembers[i].GroupID
	}

	err = db.Model(models.Group{}).Where("id IN ?", groupIds).Preload("GroupMembers").Find(&groups).Error
	if err != nil {
		return nil, err
	}

	return groups, nil
}
