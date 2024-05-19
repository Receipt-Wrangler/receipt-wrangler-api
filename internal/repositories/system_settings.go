package repositories

import (
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
)

type SystemSettingsRepository struct {
	BaseRepository
}

func NewSystemSettingsRepository(tx *gorm.DB) SystemSettingsRepository {
	repository := SystemSettingsRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository SystemSettingsRepository) GetSystemSettings() (models.SystemSettings, error) {
	db := repository.GetDB()
	var systemSettings models.SystemSettings
	var count int64

	err := db.Model(&models.SystemSettings{}).Count(&count).Error
	if err != nil {
		return models.SystemSettings{}, err
	}

	if count == 0 {
		err = db.Model(&models.SystemSettings{}).Create(&models.SystemSettings{}).Error
		if err != nil {
			return models.SystemSettings{}, err
		}
	}

	err = db.Model(&models.SystemSettings{}).First(&systemSettings).Error
	if err != nil {
		return models.SystemSettings{}, err
	}

	return systemSettings, nil
}

func (repository SystemSettingsRepository) UpdateSystemSettings(command commands.UpsertSystemSettingsCommand) (models.SystemSettings, error) {
	db := repository.GetDB()

	var existingSettings models.SystemSettings

	db.Model(&models.SystemSettings{}).First(&existingSettings)

	existingSettings.EnableLocalSignUp = command.EnableLocalSignUp
	existingSettings.EmailPollingInterval = command.EmailPollingInterval
	existingSettings.ReceiptProcessingSettingsId = command.ReceiptProcessingSettingsId
	existingSettings.FallbackReceiptProcessingSettingsId = command.FallbackReceiptProcessingSettingsId

	err := db.Model(&models.SystemSettings{}).Select("*").Where("id = ?", existingSettings.ID).Updates(&existingSettings).Error
	if err != nil {
		return models.SystemSettings{}, err
	}

	return existingSettings, nil
}
