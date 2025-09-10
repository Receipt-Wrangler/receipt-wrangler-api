package services

import (
	"receipt-wrangler/api/internal/repositories"

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

func (service *ApiKeyService) CreateApiKey() error {
	return nil
}
