package services

import (
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"

	"gorm.io/gorm"
)

type PepperService struct {
	BaseService
}

func NewPepperService(tx *gorm.DB) PepperService {
	service := PepperService{
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

	_, encryptedPepper, err := service.CreatePepper()
	if err != nil {
		return err
	}

	pepper := models.Pepper{
		Ciphertext: encryptedPepper,
		Algorithm:  "AES-256-GCM",
	}

	pepperRepository := repositories.NewPepperRepository(service.TX)
	err = pepperRepository.CreatePepper(pepper)
	if err != nil {
		return err
	}

	return nil
}

func (service PepperService) CreatePepper() (string, string, error) {
	cleartextPepper, err := utils.GetRandomString(32)
	if err != nil {
		return "", "", err
	}

	encryptedPepper, err := utils.EncryptAndEncodeToBase64(config.GetEncryptionKey(), cleartextPepper)
	if err != nil {
		return "", "", err
	}

	return cleartextPepper, encryptedPepper, nil
}

func (service PepperService) GetDecryptedPepper() (string, error) {
	pepperRepository := repositories.NewPepperRepository(service.TX)
	pepper, err := pepperRepository.GetPepper()
	if err != nil {
		return "", err
	}

	cleartext, err := utils.DecryptB64EncodedData(config.GetEncryptionKey(), pepper.Ciphertext)
	if err != nil {
		return "", err
	}

	return cleartext, nil
}
