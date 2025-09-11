package repositories

import (
	"receipt-wrangler/api/internal/models"

	"gorm.io/gorm"
)

type ApiKeyRepository struct {
	BaseRepository
}

func NewApiKeyRepository(tx *gorm.DB) ApiKeyRepository {
	repository := ApiKeyRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository ApiKeyRepository) CreateApiKey(apiKey models.ApiKey) (models.ApiKey, error) {
	err := repository.GetDB().Create(&apiKey).Error
	return apiKey, err
}

func (repository ApiKeyRepository) GetApiKeyById(id string) (models.ApiKey, error) {
	var apiKey models.ApiKey
	err := repository.GetDB().Where("id = ?", id).First(&apiKey).Error
	return apiKey, err
}
