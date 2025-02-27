package repositories

import (
	"fmt"
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
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

func (repository CategoryRepository) CreateCategory(category models.Category) (models.Category, error) {
	db := repository.GetDB()

	err := db.Model(&category).Create(&category).Error
	if err != nil {
		return models.Category{}, err
	}

	return category, nil
}

func (repository CategoryRepository) GetAllPagedCategories(pagedRequestCommand commands.PagedRequestCommand) ([]models.CategoryView, error) {
	db := repository.GetDB()
	var categories []models.CategoryView
	quotedAlias := "\"NumberOfReceipts\""

	if pagedRequestCommand.OrderBy == "numberOfReceipts" {
		pagedRequestCommand.OrderBy = quotedAlias
	}

	query := repository.Sort(db, pagedRequestCommand.OrderBy, pagedRequestCommand.SortDirection)
	query = query.Scopes(repository.Paginate(pagedRequestCommand.Page, pagedRequestCommand.PageSize))
	selectString := fmt.Sprintf("categories.id, categories.name, categories.description,  COUNT(DISTINCT receipt_categories.receipt_id) as %s", quotedAlias)
	query = query.Table("categories").
		Select(selectString).
		Joins("LEFT JOIN receipt_categories ON categories.id = receipt_categories.category_id").
		Group("categories.id, categories.name")

	err := query.Scan(&categories).Error
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (repository CategoryRepository) UpdateCategory(categoryToUpdate models.Category, querySelect string) (models.Category, error) {
	db := repository.GetDB()

	err := db.Model(models.Category{}).Where("id = ?", categoryToUpdate.ID).Updates(map[string]interface{}{"name": categoryToUpdate.Name, "description": categoryToUpdate.Description}).Error
	if err != nil {
		return models.Category{}, err
	}

	return categoryToUpdate, nil
}

func (repository CategoryRepository) DeleteCategory(categoryId uint) error {
	db := repository.GetDB()

	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Delete(&models.ReceiptCategory{}, "category_id = ?", categoryId).Error
		if err != nil {
			return err
		}

		err = tx.Where("id = ?", categoryId).Delete(&models.Category{}).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
