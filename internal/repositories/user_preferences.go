package repositories

import (
	"gorm.io/gorm/clause"
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

	err := db.
		Model(models.UserPrefernces{}).
		Where("user_id = ?", userId).
		Preload(clause.Associations).
		Find(&userPreferences).
		Error
	if err != nil {
		return models.UserPrefernces{}, err
	}

	if userPreferences.ID == 0 {
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

	return repository.GetUserPreferencesOrCreate(userPreferences.UserId)
}

func (repository UserPreferncesRepository) UpdateUserPreferences(userId uint, userPreferences models.UserPrefernces) (models.UserPrefernces, error) {
	db := repository.GetDB()
	var userPreferencesToUpdate models.UserPrefernces

	err := db.Model(models.UserPrefernces{}).Where("user_id = ?", userId).First(&userPreferencesToUpdate).Error
	if err != nil {
		return models.UserPrefernces{}, err
	}

	userPreferencesToUpdate.ShowLargeImagePreviews = userPreferences.ShowLargeImagePreviews
	userPreferencesToUpdate.QuickScanDefaultGroupId = userPreferences.QuickScanDefaultGroupId
	userPreferencesToUpdate.QuickScanDefaultPaidById = userPreferences.QuickScanDefaultPaidById
	userPreferencesToUpdate.QuickScanDefaultStatus = userPreferences.QuickScanDefaultStatus
	userPreferencesToUpdate.UserShortcuts = userPreferences.UserShortcuts

	err = db.Transaction(func(tx *gorm.DB) error {
		err = db.
			Model(&userPreferencesToUpdate).
			Select("*").
			Updates(&userPreferencesToUpdate).Error
		if err != nil {
			return err
		}

		err = db.
			Model(&userPreferencesToUpdate).
			Association("UserShortcuts").
			Replace(&userPreferencesToUpdate.UserShortcuts)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return models.UserPrefernces{}, err
	}

	return userPreferencesToUpdate, nil
}

func (repository UserPreferncesRepository) DeleteUserPreferences(userId uint) error {
	db := repository.GetDB()

	var userPreferences models.UserPrefernces

	err := db.
		Model(models.UserPrefernces{}).
		Where("user_id = ?", userId).
		First(&userPreferences).Error
	if err != nil {
		return err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		txErr := tx.
			Model(models.UserShortcut{}).
			Where("user_prefernces_id = ?", userPreferences.ID).
			Delete(&models.UserShortcut{}).Error
		if txErr != nil {
			return txErr
		}

		txErr = tx.
			Model(models.UserPrefernces{}).
			Where("user_id = ?", userId).
			Delete(&userPreferences).Error
		if txErr != nil {
			return txErr
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
