package repositories

import (
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
