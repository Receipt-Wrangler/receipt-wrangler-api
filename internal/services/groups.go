package services

import (
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/simpleutils"

	"gorm.io/gorm"
)

func GetGroupsForUser(userId string) ([]models.Group, error) {
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

func DeleteGroup(groupId string) error {
	db := db.GetDB()
	var receipts []models.Receipt

	group, err := repositories.GetGroupById(groupId, false)
	if err != nil {
		return err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		txErr := db.Model(models.Receipt{}).Where("group_id = ?", groupId).Find(&receipts).Error
		if txErr != nil {
			return txErr
		}

		for i := 0; i < len(receipts); i++ {
			txErr = DeleteReceipt(simpleutils.UintToString(receipts[i].ID))
			if txErr != nil {
				return txErr
			}
		}

		txErr = db.Where("group_id = ?", groupId).Delete(&models.GroupMember{}).Error
		if txErr != nil {
			return txErr
		}

		txErr = db.Model(&group).Delete(&group).Error
		if txErr != nil {
			return txErr
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
