package repositories

import (
	"fmt"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"

	"gorm.io/gorm"
)

type GroupSettingsRepository struct {
	BaseRepository
}

func NewGroupSettingsRepository(tx *gorm.DB) GroupSettingsRepository {
	repository := GroupSettingsRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository GroupSettingsRepository) CreateGroupSettings(groupSettings models.GroupSettings) (models.GroupSettings, error) {
	db := repository.GetDB()

	err := db.Model(&groupSettings).Create(&groupSettings).Error
	if err != nil {
		return models.GroupSettings{}, err
	}

	return groupSettings, nil
}

func (repository GroupSettingsRepository) UpdateGroupSettings(groupId string, command commands.UpdateGroupSettingsCommand) (models.GroupSettings, error) {
	db := repository.GetDB()
	var groupSettings models.GroupSettings

	err := db.Model(&models.GroupSettings{}).Where("group_id = ?", groupId).First(&groupSettings).Error
	if err != nil {
		return models.GroupSettings{}, err
	}

	err = db.Transaction(func(tx *gorm.DB) error {

		err = tx.Model(&groupSettings).Where("id = ?", groupSettings.ID).Association("EmailToRead").Replace(&command.EmailToRead)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		err = tx.Model(&groupSettings).Where("id = ?", groupSettings.ID).Association("SubjectLineRegexes").Replace(&command.SubjectLineRegexes)
		if err != nil {
			return err
		}

		err = tx.Model(&groupSettings).Where("id = ?", groupSettings.ID).Association("EmailWhiteList").Replace(&command.EmailWhiteList)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return models.GroupSettings{}, err
	}

	return groupSettings, nil
}
