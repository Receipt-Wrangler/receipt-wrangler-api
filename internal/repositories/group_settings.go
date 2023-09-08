package repositories

import (
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
