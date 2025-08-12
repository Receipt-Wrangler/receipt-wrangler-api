package repositories

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
)

type GroupReceiptSettingsRepository struct {
	BaseRepository
}

func NewGroupReceiptSettingsRepository(tx *gorm.DB) GroupReceiptSettingsRepository {
	repository := GroupReceiptSettingsRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository GroupReceiptSettingsRepository) CreateGroupReceiptSettings(groupId uint) (models.GroupReceiptSettings, error) {
	db := repository.GetDB()

	groupReceiptSettingsToCreate := models.GroupReceiptSettings{
		GroupId: groupId,
	}

	err := db.Model(models.GroupReceiptSettings{}).Create(&groupReceiptSettingsToCreate).Error
	if err != nil {
		return models.GroupReceiptSettings{}, err
	}

	return groupReceiptSettingsToCreate, nil
}

func (repository GroupReceiptSettingsRepository) UpdateGroupReceiptSettings(
	groupId string,
	command commands.UpdateGroupReceiptSettingsCommand,
) (models.GroupReceiptSettings, error) {
	db := repository.GetDB()

	var groupReceiptSettings models.GroupReceiptSettings

	err := db.Model(&groupReceiptSettings).Where("group_id = ?", groupId).Preload(clause.Associations).First(&groupReceiptSettings).Error
	if err != nil {
		return models.GroupReceiptSettings{}, err
	}

	groupReceiptSettings.HideImages = command.HideImages
	groupReceiptSettings.HideReceiptCategories = command.HideReceiptCategories
	groupReceiptSettings.HideReceiptTags = command.HideReceiptTags
	groupReceiptSettings.HideItemCategories = command.HideItemCategories
	groupReceiptSettings.HideItemTags = command.HideItemTags
	groupReceiptSettings.HideComments = command.HideComments
	groupReceiptSettings.HideShareCategories = command.HideShareCategories
	groupReceiptSettings.HideShareTags = command.HideShareTags

	err = db.Select("*").Model(*&groupReceiptSettings).Updates(groupReceiptSettings).Error
	if err != nil {
		return models.GroupReceiptSettings{}, err
	}

	return groupReceiptSettings, nil
}
