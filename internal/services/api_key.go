package services

import (
	"errors"
	"fmt"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strings"
	"time"

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

	b64Id := utils.Base64URLEncode([]byte(id))

	secret, err := utils.GetRandomString(64)
	if err != nil {
		return "", err
	}

	b64secret := utils.Base64Encode([]byte(secret))

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
		CreatedBy:   &userId,
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
	return utils.Base64Encode(hmac), nil
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
	decodedSecret, err := utils.Base64URLDecode(secret)
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

func (service *ApiKeyService) UpdateApiKeyLastUsedDate(id string) error {
	apiKeyRepository := repositories.NewApiKeyRepository(service.TX)
	return apiKeyRepository.UpdateApiKeyLastUsedDate(id)
}

func (service *ApiKeyService) UpdateApiKey(apiKeyId string, userId uint, command commands.UpsertApiKeyCommand) error {
	apiKeyRepository := repositories.NewApiKeyRepository(service.TX)

	// First verify the API key exists and belongs to the user
	existingKey, err := apiKeyRepository.GetApiKeyById(apiKeyId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("API key not found")
		}
		return err
	}

	if existingKey.UserID == nil || *existingKey.UserID != userId {
		return errors.New("API key not found")
	}

	// Update the API key
	return apiKeyRepository.UpdateApiKey(apiKeyId, userId, command.Name, command.Description, command.Scope)
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
		}
	}

	return apiKeyViews, count, nil
}

func (service *ApiKeyService) DeleteApiKey(apiKeyId string, userId uint, isAdmin bool) error {
	startTime := time.Now()
	apiKeyRepository := repositories.NewApiKeyRepository(service.TX)

	// First verify the API key exists
	existingKey, err := apiKeyRepository.GetApiKeyById(apiKeyId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("API key not found")
		}
		return err
	}

	// Check authorization: admin can delete any key, non-admin can only delete their own
	if !isAdmin {
		if existingKey.UserID == nil || *existingKey.UserID != userId {
			return errors.New("API key not found")
		}
	}

	// Delete the API key
	deleteErr := apiKeyRepository.DeleteApiKey(apiKeyId)

	// Create system task to track the deletion
	endTime := time.Now()
	systemTaskService := NewSystemTaskService(service.TX)

	resultDescription := fmt.Sprintf("Deleted API key '%s' (ID: %s)", existingKey.Name, apiKeyId)
	if existingKey.UserID != nil {
		resultDescription += fmt.Sprintf(" owned by user %d", *existingKey.UserID)
	}

	systemTaskCommand := commands.UpsertSystemTaskCommand{
		Type:                 models.API_KEY_DELETED,
		Status:               models.SYSTEM_TASK_SUCCEEDED,
		AssociatedEntityType: models.API_KEY,
		AssociatedEntityId:   0, // Not used for API keys
		ApiKeyId:             &apiKeyId,
		StartedAt:            startTime,
		EndedAt:              &endTime,
		ResultDescription:    resultDescription,
		RanByUserId:          &userId,
	}

	if deleteErr != nil {
		systemTaskCommand.Status = models.SYSTEM_TASK_FAILED
		systemTaskCommand.ResultDescription = deleteErr.Error()
	}

	_, taskErr := systemTaskService.CreateSystemTaskFromError(systemTaskCommand, deleteErr)
	if taskErr != nil {
		logging.LogStd(logging.LOG_LEVEL_INFO, taskErr.Error())
	}

	return deleteErr
}
