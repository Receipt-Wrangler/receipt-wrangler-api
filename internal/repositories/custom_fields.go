package repositories

import (
	"gorm.io/gorm"
)

type CustomFieldRepository struct {
	BaseRepository
}

func NewCustomFieldRepository(tx *gorm.DB) CustomFieldRepository {
	repository := CustomFieldRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}
