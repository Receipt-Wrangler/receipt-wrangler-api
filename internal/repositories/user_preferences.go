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

func (repository UserPreferncesRepository) GetUserPreferencesOrCreate(userId uint) (models.UserPrefernces, error) {
	db := repository.GetDB()
	var userPreferences models.UserPrefernces

	err := db.Model(models.UserPrefernces{}).Where("user_id = ?", userId).Find(&userPreferences).Error
	if err != nil {
		return models.UserPrefernces{}, err
	}

	if userPreferences.ID == 0 {
		if err != nil {
			return models.UserPrefernces{}, err
		}

		userPreferencesToCreate := models.UserPrefernces{
			UserId: userId,
		}

		userPreferences, err = repository.CreateUserPreferences(userPreferencesToCreate)
		if err != nil {
			return models.UserPrefernces{}, err
		}
	}

	return userPreferences, nil
}

func (repository UserPreferncesRepository) CreateUserPreferences(userPreferences models.UserPrefernces) (models.UserPrefernces, error) {
	db := repository.GetDB()

	err := db.Model(models.UserPrefernces{}).Create(&userPreferences).Error
	if err != nil {
		return models.UserPrefernces{}, err
	}

	return userPreferences, nil
}

func (repository UserPreferncesRepository) UpdateUserPreferences(userId uint, userPreferences models.UserPrefernces) (models.UserPrefernces, error) {
	db := repository.GetDB()
	update := map[string]interface{}{"quickScanDefaultGroupId": userPreferences.QuickScanDefaultGroupId, "quickScanDefaultPaidById": userPreferences.QuickScanDefaultPaidById, "quickScanDefaultStatus": userPreferences.QuickScanDefaultStatus}

	err := db.Model(models.UserPrefernces{}).Where("user_id = ?", userId).Updates(&update).Error
	if err != nil {
		return models.UserPrefernces{}, err
	}

	return userPreferences, nil
}
