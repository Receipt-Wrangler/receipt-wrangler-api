package models

import (
	"gorm.io/gorm"
)

type Group struct {
	BaseModel
	Name           string        `gorm:"not null" json:"name"`
	IsDefaultGroup bool          `json:"isDefault"`
	GroupMembers   []GroupMember `json:"groupMembers"`
}

func (groupToUpdate *Group) BeforeUpdate(tx *gorm.DB) (err error) {
	if groupToUpdate.ID > 0 {
		var dbGroup Group

		err := tx.Table("groups").Where("id = ?", groupToUpdate.ID).Select("Name").Error
		if err != nil {
			return err
		}

		if groupToUpdate.Name != dbGroup.Name {
			// oldGroupPath, err := utils.BuildGroupPath(dbGroup.ID, dbGroup.Name)
		}
	}

	return nil
}
