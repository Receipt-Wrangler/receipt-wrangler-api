package models

import (
	"os"
	"receipt-wrangler/api/internal/simpleutils"

	"gorm.io/gorm"
)

// Group in the system
//
// swagger:model
type Group struct {
	BaseModel

	// Name of the group
	//
	// required: true
	Name string `gorm:"not null" json:"name"`

	// Is default group (not used yet)
	IsDefaultGroup bool `json:"isDefault"`

	// Members of the group
	GroupMembers []GroupMember `json:"groupMembers"`

	// Current status fo the group
	//
	// required: true
	Status GroupStatus `gorm:"default:'ACTIVE'; not null" json:"status"`
}

func (groupToUpdate *Group) BeforeUpdate(tx *gorm.DB) (err error) {
	if groupToUpdate.ID > 0 {
		var dbGroup Group

		err := tx.Table("groups").Where("id = ?", groupToUpdate.ID).Select("id", "name").Find(&dbGroup).Error
		if err != nil {
			return err
		}

		if groupToUpdate.Name != dbGroup.Name {
			oldGroupId := simpleutils.UintToString(dbGroup.ID)
			newGroupId := simpleutils.UintToString(groupToUpdate.ID)

			oldGroupPath, err := simpleutils.BuildGroupPathString(oldGroupId, dbGroup.Name)
			if err != nil {
				return err
			}

			newGroupPath, err := simpleutils.BuildGroupPathString(newGroupId, groupToUpdate.Name)
			if err != nil {
				return err
			}

			err = os.Rename(oldGroupPath, newGroupPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (deletedGroup *Group) AfterDelete(tx *gorm.DB) (err error) {
	if deletedGroup.ID > 0 {
		dataPath, err := simpleutils.BuildGroupPathString(simpleutils.UintToString(deletedGroup.ID), deletedGroup.Name)
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
