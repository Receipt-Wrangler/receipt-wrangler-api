package repositories

import (
	"receipt-wrangler/api/internal/models"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	BaseRepository
}

func NewCategoryRepository(tx *gorm.DB) CategoryRepository {
	repository := CategoryRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository CategoryRepository) GetAllCategories(querySelect string) ([]models.Category, error) {
	db := repository.GetDB()
	var categories []models.Category

	err := db.Table("categories").Select(querySelect).Find(&categories).Error
	if err != nil {
		return nil, err
	}

	return categories, nil
}
