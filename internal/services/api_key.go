package services

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"

	"gorm.io/gorm"
)

type ApiKeyService struct {
	BaseService
}

func NewApiKeyService(tx *gorm.DB) ApiKeyService {
	service := ApiKeyService{
		BaseService: BaseService{
			DB: repositories.GetDB(),
			TX: tx,
		},
	}

	return service
}

// return full api key for use
func (service *ApiKeyService) CreateApiKey(userId uint, command commands.UpsertApiKeyCommand) error {
	prefix := "key_live"
	version := 1

	id, err := utils.GetRandomString(16)
	if err != nil {
		return err
	}

	b64Id, err := utils.Base64EncodeBytes([]byte(id))
	if err != nil {
		return err
	}

	secret, err := utils.GetRandomString(32)
	if err != nil {
		return err
	}

	generatedHmac, err := service.GenerateApiKeyHmac(secret)
	if err != nil {
		return err
	}

	apiKey := models.ApiKey{
		ID:          b64Id,
		UserID:      &userId,
		Name:        command.Name,
		Description: command.Description,
		Scope:       command.Scope,
		Prefix:      prefix,
		Hmac:        generatedHmac,
		Version:     version,
	}
	return nil
}

func (service *ApiKeyService) GenerateApiKeyHmac(secret string) (string, error) {
	pepperService := NewPepperService(service.TX)

	clearTextPepper, err := pepperService.GetDecryptedPepper()
	if err != nil {
		return "", err
	}

	hmac := utils.GenerateHmac([]byte(clearTextPepper), []byte(secret))
	return utils.Base64EncodeBytes(hmac), nil
}
