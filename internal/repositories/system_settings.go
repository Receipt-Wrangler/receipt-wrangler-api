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
		err = db.Model(&models.SystemSettings{}).Create(&models.SystemSettings{}).Error
		if err != nil {
			return models.SystemSettings{}, err
		}
	}

	err = db.Model(&models.SystemSettings{}).Preload(clause.Associations).First(&systemSettings).Error
	if err != nil {
		return models.SystemSettings{}, err
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

	var existingSettings models.SystemSettings

	db.Model(&models.SystemSettings{}).First(&existingSettings)

	existingSettings.EnableLocalSignUp = command.EnableLocalSignUp
	existingSettings.DebugOcr = command.DebugOcr
	existingSettings.NumWorkers = command.NumWorkers
	existingSettings.CurrencyDisplay = command.CurrencyDisplay
	existingSettings.CurrencyThousandthsSeparator = command.CurrencyThousandthsSeparator
	existingSettings.CurrencyDecimalSeparator = command.CurrencyDecimalSeparator
	existingSettings.CurrencySymbolPosition = command.CurrencySymbolPosition
	existingSettings.CurrencyHideDecimalPlaces = command.CurrencyHideDecimalPlaces
	existingSettings.EmailPollingInterval = command.EmailPollingInterval
	existingSettings.ReceiptProcessingSettingsId = command.ReceiptProcessingSettingsId
	existingSettings.FallbackReceiptProcessingSettingsId = command.FallbackReceiptProcessingSettingsId
	existingSettings.AsynqConcurrency = command.AsynqConcurrency
	existingSettings.AsynqQuickScanPriority = command.AsynqQuickScanPriority
	existingSettings.AsynqEmailReceiptProcessingPriority = command.AsynqEmailReceiptProcessingPriority
	existingSettings.AsynqEmailPollingPriority = command.AsynqEmailPollingPriority
	existingSettings.AsynqEmailReceiptImageCleanupPriority = command.AsynqEmailReceiptImageCleanupPriority

	err := db.Model(&models.SystemSettings{}).Select("*").Where("id = ?", existingSettings.ID).Updates(&existingSettings).Error
	if err != nil {
		return models.SystemSettings{}, err
	}

	return existingSettings, nil
}
