package repositories

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
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
		err = db.Model(&models.SystemSettings{}).Create(&models.SystemSettings{
			BaseModel: models.BaseModel{
				ID: 1,
			},
		}).Error
		if err != nil {
			return models.SystemSettings{}, err
		}
	}

	err = db.Model(&models.SystemSettings{}).Preload(clause.Associations).Preload("AsynqQueueConfigurations").First(&systemSettings).Error
	if err != nil {
		return models.SystemSettings{}, err
	}

	// NOTE: Eventually this can get deleted. This is to fix associations not working if ID Is 0
	if systemSettings.ID == 0 {
		err = db.Model(models.SystemSettings{}).Where("id = 0").Update("id", 1).Error
		if err != nil {
			return models.SystemSettings{}, err
		}
	}

	if len(systemSettings.AsynqQueueConfigurations) == 0 {
		systemSettings.AsynqQueueConfigurations = models.GetAllDefaultQueueConfigurations()
	}

	return systemSettings, nil
}

func (repository SystemSettingsRepository) GetSystemReceiptProcessingSettings() (structs.SystemReceiptProcessingSettings, error) {
	systemSettings, err := repository.GetSystemSettings()
	if err != nil {
		return structs.SystemReceiptProcessingSettings{}, err
	}

	if systemSettings.ReceiptProcessingSettings.ID == 0 {
		return structs.SystemReceiptProcessingSettings{}, errors.New("ReceiptProcessingSettings do not exist")
	}

	return structs.SystemReceiptProcessingSettings{
		ReceiptProcessingSettings:         systemSettings.ReceiptProcessingSettings,
		FallbackReceiptProcessingSettings: systemSettings.FallbackReceiptProcessingSettings,
	}, nil
}

func (repository SystemSettingsRepository) UpdateSystemSettings(command commands.UpsertSystemSettingsCommand) (models.SystemSettings, error) {
	db := repository.GetDB()

	existingSettings, err := repository.GetSystemSettings()
	if err != nil {
		return models.SystemSettings{}, err
	}

	updatedSettings, err := command.ToSystemSettings()
	if err != nil {
		return models.SystemSettings{}, err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		txErr := db.Model(&models.SystemSettings{}).Select("*").Where("id = ?", existingSettings.ID).Updates(&updatedSettings).Error
		if txErr != nil {
			return txErr
		}

		txErr = tx.Model(&updatedSettings).
			Where("id = ?", existingSettings.ID).
			Association("AsynqQueueConfigurations").
			Replace(&updatedSettings.AsynqQueueConfigurations)
		if txErr != nil {
			return txErr
		}

		return nil
	})

	if err != nil {
		return models.SystemSettings{}, nil
	}

	return updatedSettings, nil
}
