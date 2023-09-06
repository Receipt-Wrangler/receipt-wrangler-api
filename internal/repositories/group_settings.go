package repositories

import (
	"gorm.io/gorm"
)

type GroupSettingsRepository struct {
	BaseRepository
}

func NewGroupSettingsRepository(tx *gorm.DB) GroupSettingsRepository {
	repository := GroupSettingsRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}
