package services

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"
	"strings"

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

func (service *ApiKeyService) CreateApiKey(userId uint, command commands.UpsertApiKeyCommand) (string, error) {
	prefix := "key_live"
	version := 1

	id, err := utils.GetRandomString(16)
	if err != nil {
		return "", err
	}

	b64Id := utils.Base64EncodeBytes([]byte(id))

	secret, err := utils.GetRandomString(32)
	if err != nil {
		return "", err
	}

	generatedHmac, err := service.GenerateApiKeyHmac(secret)
	if err != nil {
		return "", err
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

	apiKeyRepository := repositories.NewApiKeyRepository(service.TX)
	_, err = apiKeyRepository.CreateApiKey(apiKey)
	if err != nil {
		return "", err
	}

	return service.BuildV1ApiKey(prefix, version, id, secret), nil
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

func (service *ApiKeyService) BuildV1ApiKey(
	prefix string,
	version int,
	id string,
	secret string,
) string {
	stringVersion := utils.UintToString(uint(version))
	return strings.Join([]string{prefix, stringVersion, id, secret}, ".")
}
