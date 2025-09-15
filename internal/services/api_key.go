package services

import (
	"errors"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strings"

	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/golang-jwt/jwt/v5"
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
	prefix := constants.V1Prefix
	version := 1

	id, err := utils.GetRandomString(32)
	if err != nil {
		return "", err
	}

	b64Id := utils.Base64EncodeBytes([]byte(id))

	secret, err := utils.GetRandomString(64)
	if err != nil {
		return "", err
	}

	b64secret := utils.Base64EncodeBytes([]byte(secret))

	b64hmac, err := service.GenerateApiKeyHmac(secret)
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
		Hmac:        b64hmac,
		Version:     version,
	}

	apiKeyRepository := repositories.NewApiKeyRepository(service.TX)
	_, err = apiKeyRepository.CreateApiKey(apiKey)
	if err != nil {
		return "", err
	}

	return service.BuildV1ApiKey(prefix, version, b64Id, b64secret), nil
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

func (service *ApiKeyService) ValidateV1ApiKey(apiKey string) (models.ApiKey, error) {
	parts := strings.Split(apiKey, ".")
	if len(parts) != constants.V1PartLength {
		return models.ApiKey{}, errors.New("invalid api key structure")
	}

	b64id := parts[2]

	apiKeyRepository := repositories.NewApiKeyRepository(service.TX)
	apiKeyData, err := apiKeyRepository.GetApiKeyById(b64id)
	if err != nil {
		return models.ApiKey{}, err
	}

	secret := parts[3]
	decodedSecret, err := utils.Base64Decode(secret)
	if err != nil {
		return models.ApiKey{}, err
	}

	b64hmac, err := service.GenerateApiKeyHmac(string(decodedSecret))
	if err != nil {
		return models.ApiKey{}, err
	}

	if b64hmac != apiKeyData.Hmac {
		return models.ApiKey{}, errors.New("invalid api key secret")
	}

	return apiKeyData, nil
}

func (service *ApiKeyService) GetClaimsFromApiKey(key models.ApiKey) (validator.ValidatedClaims, error) {
	userRepository := repositories.NewUserRepository(service.TX)
	user, err := userRepository.GetUserById(*key.UserID)
	if err != nil {
		return validator.ValidatedClaims{}, err
	}

	claims := structs.Claims{
		DefaultAvatarColor: user.DefaultAvatarColor,
		UserId:             user.ID,
		Username:           user.Username,
		Displayname:        user.DisplayName,
		UserRole:           user.UserRole,
		ApiKeyScope:        models.ApiKeyScope(key.Scope),
		RegisteredClaims:   jwt.RegisteredClaims{},
	}

	result := validator.ValidatedClaims{
		CustomClaims:     &claims,
		RegisteredClaims: validator.RegisteredClaims{},
	}

	return result, nil
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

func (service *ApiKeyService) GetPagedApiKeys(command commands.PagedApiKeyRequestCommand, userId string) ([]models.ApiKeyView, int64, error) {
	apiKeyRepository := repositories.NewApiKeyRepository(service.TX)

	apiKeys, count, err := apiKeyRepository.GetPagedApiKeys(command, userId)
	if err != nil {
		return nil, 0, err
	}

	apiKeyViews := make([]models.ApiKeyView, len(apiKeys))
	for i, apiKey := range apiKeys {
		apiKeyViews[i] = models.ApiKeyView{
			ID:              apiKey.ID,
			CreatedAt:       apiKey.CreatedAt,
			UpdatedAt:       apiKey.UpdatedAt,
			CreatedBy:       apiKey.CreatedBy,
			CreatedByString: apiKey.CreatedByString,
			Name:            apiKey.Name,
			Description:     apiKey.Description,
			UserID:          apiKey.UserID,
			Scope:           apiKey.Scope,
			LastUsedAt:      apiKey.LastUsedAt,
			RevokedAt:       apiKey.RevokedAt,
		}
	}

	return apiKeyViews, count, nil
}
