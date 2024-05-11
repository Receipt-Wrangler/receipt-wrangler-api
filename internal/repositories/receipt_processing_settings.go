package repositories

import (
	"errors"
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
)

type ReceiptProcessingSettings struct {
	BaseRepository
}

func NewReceiptProcessingSettings(tx *gorm.DB) ReceiptProcessingSettings {
	repository := ReceiptProcessingSettings{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository ReceiptProcessingSettings) GetPagedReceiptProcessingSettings(command commands.PagedRequestCommand) ([]models.ReceiptProcessingSettings, int64, error) {
	db := repository.GetDB()
	var results []models.ReceiptProcessingSettings
	var count int64

	validColumn := repository.isValidColumn(command.OrderBy)
	if !validColumn {
		return nil, 0, errors.New("invalid column: " + command.OrderBy)
	}

	query := db.Model(&models.ReceiptProcessingSettings{})

	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	query = repository.Sort(query, command.OrderBy, command.SortDirection)
	query = query.Scopes(repository.Paginate(command.Page, command.PageSize))

	err = query.Find(&results).Error
	if err != nil {
		return nil, 0, err
	}

	return results, count, nil
}

func (repository ReceiptProcessingSettings) isValidColumn(orderBy string) bool {
	return orderBy == "name" || orderBy == "description" || orderBy == "ai_type" || orderBy == "ocr_engine" || orderBy == "created_at" || orderBy == "updated_at"
}
