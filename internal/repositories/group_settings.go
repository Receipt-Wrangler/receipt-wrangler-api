package repositories

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (repository GroupSettingsRepository) UpdateGroupSettings(groupId string, groupSettings commands.UpdateGroupSettingsCommand) (models.GroupSettings, error) {
	db := repository.GetDB()
	var result models.GroupSettings

	err := db.Model(&models.GroupSettings{}).Where("group_id = ?", groupId).Clauses(clause.OnConflict{UpdateAll: true}).Create(&groupSettings).Error
	if err != nil {
		return models.GroupSettings{}, err
	}

	err = db.Model(&models.GroupSettings{}).Where("group_id = ?", groupId).First(&result).Error
	if err != nil {
		return models.GroupSettings{}, err
	}

	return result, nil
}
