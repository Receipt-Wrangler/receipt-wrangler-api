package repositories

import (
	"errors"
	"receipt-wrangler/api/internal/commands"
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

func (repository ApiKeyRepository) GetPagedApiKeys(command commands.PagedApiKeyRequestCommand, userId string) ([]models.ApiKey, int64, error) {
	db := repository.GetDB()
	var results []models.ApiKey
	var count int64

	query := db.Model(&models.ApiKey{})

	if !repository.isValidColumn(command.OrderBy) {
		return nil, 0, errors.New("invalid column")
	}

	if command.ApiKeyFilter.AssociatedApiKeys == commands.ASSOCIATED_API_KEYS_MINE {
		query = query.Where("user_id = ?", userId)
	}

	query.Count(&count)

	query = repository.Sort(query, command.OrderBy, command.SortDirection)
	query = query.Scopes(repository.Paginate(command.Page, command.PageSize))

	err := query.Find(&results).Error
	if err != nil {
		return nil, 0, err
	}

	return results, count, nil
}

func (repository ApiKeyRepository) isValidColumn(orderBy string) bool {
	return orderBy == "name" ||
		orderBy == "description" ||
		orderBy == "created_at" ||
		orderBy == "revoked_at" ||
		orderBy == "updated_at" ||
		orderBy == "last_used_at"
}
