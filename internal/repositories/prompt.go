package repositories

import (
	"errors"
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
)

type PromptRepository struct {
	BaseRepository
}

func NewPromptRepository(tx *gorm.DB) PromptRepository {
	repository := PromptRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository PromptRepository) GetPagedPrompts(command commands.PagedRequestCommand) ([]models.Prompt, int64, error) {
	db := repository.GetDB()
	var results []models.Prompt
	var count int64

	query := db.Model(&models.Prompt{})

	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	validColumn := repository.isValidColumn(command.OrderBy)
	if !validColumn {
		return nil, 0, errors.New("invalid column: " + command.OrderBy)
	}

	query = repository.Sort(query, command.OrderBy, command.SortDirection)
	query = query.Scopes(repository.Paginate(command.Page, command.PageSize))

	err = query.Find(&results).Error
	if err != nil {
		return nil, 0, err
	}

	return results, count, nil
}

func (repository PromptRepository) isValidColumn(orderBy string) bool {
	return orderBy == "name" || orderBy == "description" || orderBy == "created_at" || orderBy == "updated_at"
}
