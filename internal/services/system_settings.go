package services

import (
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
)

type SystemSettingsService struct {
	BaseService
}

func NewSystemSettingsService(tx *gorm.DB) SystemSettingsService {
	service := SystemSettingsService{BaseService: BaseService{
		DB: repositories.GetDB(),
		TX: tx,
	}}
	return service
}

func (service SystemSettingsService) GetFeatureConfig() (structs.FeatureConfig, error) {
	systemSettingsRepository := repositories.NewSystemSettingsRepository(service.TX)
	featureConfig := structs.FeatureConfig{}

	systemSettings, err := systemSettingsRepository.GetSystemSettings()
	if err != nil {
		return structs.FeatureConfig{}, err
	}

	aiPoweredReceipts := systemSettings.ReceiptProcessingSettingsId != nil

	featureConfig.EnableLocalSignUp = systemSettings.EnableLocalSignUp
	featureConfig.AiPoweredReceipts = aiPoweredReceipts

	return featureConfig, nil
}
