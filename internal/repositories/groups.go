package repositories

import (
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"

	"gorm.io/gorm"
)

func CreateGroup(group models.Group, userId uint) (models.Group, error) {
	db := db.GetDB()
	var returnGroup models.Group
	err := db.Transaction(func(tx *gorm.DB) error {

		txErr := db.Model(&group).Create(&group).Error
		if txErr != nil {
			return txErr
		}

		groupMember := models.GroupMember{
			UserID:    userId,
			GroupID:   group.ID,
			GroupRole: models.OWNER,
		}

		txErr = db.Model(&groupMember).Create(&groupMember).Error
		if txErr != nil {
			return txErr
		}

		return nil
	})
	if err != nil {
		return models.Group{}, err
	}

	err = db.Model(models.Group{}).Where("id = ?", group.ID).Preload("GroupMembers").Find(&returnGroup).Error
	if err != nil {
		return models.Group{}, err
	}

	return returnGroup, nil
}

func UpdateGroup(group models.Group, groupId string) error {
	db := db.GetDB()
	u64Id, err := utils.StringToUint64(groupId)
	if err != nil {
		return err
	}

	group.ID = uint(u64Id)

	err = db.Session(&gorm.Session{FullSaveAssociations: true}).Model(&group).Updates(&group).Error
	if err != nil {
		return err
	}

	return nil
}

func DeleteGroup(groupId string) error {
	db := db.GetDB()

	group, err := GetGroupById(groupId, false)
	if err != nil {
		return err
	}

	err = db.Model(models.Group{}).Delete(&group).Error
	if err != nil {
		return err
	}

	return nil
}

func GetGroupById(id string, preloadGroupMembers bool) (models.Group, error) {
	db := db.GetDB()
	var group models.Group
	var err error

	if preloadGroupMembers {
		err = db.Model(models.Group{}).Where("id = ?", id).Preload("GroupMembers").Find(&group).Error

	} else {
		err = db.Model(models.Group{}).Where("id = ?", id).Find(&group).Error
	}

	if err != nil {
		return models.Group{}, err
	}

	return group, nil
}
