package repositories

import (
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"

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
