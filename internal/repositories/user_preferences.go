package repositories

import (
	"receipt-wrangler/api/internal/models"

	"gorm.io/gorm"
)

type UserPreferncesRepository struct {
	BaseRepository
}

func NewUserPreferencesRepository(tx *gorm.DB) UserPreferncesRepository {
	repository := UserPreferncesRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository UserPreferncesRepository) GetUserPreferencesOrCreate(userId string) (models.UserPrefernces, error) {
	db := repository.GetDB()
	var userPreferences models.UserPrefernces

	err := db.Model(models.UserPrefernces{}).Where("user_id = ?", userId).Find(&userPreferences).Error
	if err != nil {
		return models.UserPrefernces{}, err
	}

	return userPreferences, nil
}
