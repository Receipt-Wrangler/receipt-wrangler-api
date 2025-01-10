package models

import (
	"os"
	"receipt-wrangler/api/internal/utils"

	"gorm.io/gorm"
)

type Group struct {
	BaseModel
	Name                 string               `gorm:"not null" json:"name"`
	IsDefaultGroup       bool                 `json:"isDefault"`
	GroupMembers         []GroupMember        `json:"groupMembers"`
	GroupSettings        GroupSettings        `json:"groupSettings"`
	GroupReceiptSettings GroupReceiptSettings `json:"groupReceiptSettings"`
	Status               GroupStatus          `gorm:"default:'ACTIVE'; not null" json:"status"`
	IsAllGroup           bool                 `json:"isAllGroup" gorm:"default:false"`
}

func (groupToUpdate *Group) BeforeUpdate(tx *gorm.DB) (err error) {
	if groupToUpdate.ID > 0 {
		var dbGroup Group

		err := tx.Table("groups").Where("id = ?", groupToUpdate.ID).Select("id", "name").Find(&dbGroup).Error
		if err != nil {
			return err
		}

		if groupToUpdate.Name != dbGroup.Name {
			oldGroupId := utils.UintToString(dbGroup.ID)
			newGroupId := utils.UintToString(groupToUpdate.ID)

			oldGroupPath, err := utils.BuildGroupPathString(oldGroupId, dbGroup.Name)
			if err != nil {
				return err
			}

			newGroupPath, err := utils.BuildGroupPathString(newGroupId, groupToUpdate.Name)
			if err != nil {
				return err
			}

			os.Rename(oldGroupPath, newGroupPath)
		}
	}

	return nil
}

func (deletedGroup *Group) AfterDelete(tx *gorm.DB) (err error) {
	if deletedGroup.ID > 0 {
		dataPath, err := utils.BuildGroupPathString(utils.UintToString(deletedGroup.ID), deletedGroup.Name)
		if err != nil {
			return err
		}

		err = os.RemoveAll(dataPath)
		if err != nil {
			return err
		}
	}

	return nil
}
