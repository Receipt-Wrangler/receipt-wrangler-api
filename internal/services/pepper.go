package services

import (
	"gorm.io/gorm"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"
)

type PepperService struct {
	BaseService
}

func NewPepperService(tx *gorm.DB) ImportService {
	service := ImportService{
		BaseService: BaseService{
			DB: repositories.GetDB(),
			TX: tx,
		},
	}

	return service
}

func (service PepperService) InitPepper() error {
	var pepperCount = int64(0)
	err := service.GetDB().Model(models.Pepper{}).Count(&pepperCount).Error
	if err != nil {
		return err
	}

	if pepperCount > 0 {
		return nil
	}

	encryptedPepper, err := service.CreatePepper()
	if err != nil {
		return err
	}

	pepper := models.Pepper{
		Ciphertext: encryptedPepper,
		Algorithm:  "AES-256-GCM",
	}

	pepperRepository := repositories.NewPepperRepository(service.TX)
	return pepperRepository.CreatePepper(pepper)
}

func (service PepperService) CreatePepper() (string, error) {
	pepper, err := utils.GetRandomString(32)
	if err != nil {
		return "", err
	}

	return utils.EncryptAndEncodeToBase64(config.GetEncryptionKey(), pepper)
}
