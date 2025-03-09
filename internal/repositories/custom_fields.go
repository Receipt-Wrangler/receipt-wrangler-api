package repositories

import (
	"errors"
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
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

func (repository CustomFieldRepository) GetPagedCustomFields(pagedRequestCommand commands.PagedRequestCommand) ([]models.CustomField, int64, error) {
	db := repository.GetDB()
	var customFields []models.CustomField

	err := repository.validateOrderBy(pagedRequestCommand.OrderBy)
	if err != nil {
		return customFields, 0, err
	}

	query := repository.Sort(db, pagedRequestCommand.OrderBy, pagedRequestCommand.SortDirection)
	query = query.Scopes(repository.Paginate(pagedRequestCommand.Page, pagedRequestCommand.PageSize))

	err = query.Model(&models.CustomField{}).Scan(&customFields).Error
	if err != nil {
		return nil, 0, err
	}

	var count int64
	err = db.Model(&models.CustomField{}).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	return customFields, count, nil
}

func (repository CustomFieldRepository) validateOrderBy(orderBy string) error {
	if orderBy != "name" && orderBy != "type" && orderBy != "description" {
		return errors.New("invalid orderBy")
	}

	return nil
}
