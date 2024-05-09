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

func (repository PromptRepository) GetPromptById(id string) (models.Prompt, error) {
	db := repository.GetDB()
	var prompt models.Prompt

	err := db.Model(models.Prompt{}).Where("id = ?", id).First(&prompt).Error
	if err != nil {
		return models.Prompt{}, err
	}

	return prompt, nil
}

func (repository PromptRepository) UpdatePromptById(id string, command commands.UpsertPromptCommand) (models.Prompt, error) {
	db := repository.GetDB()
	prompt, err := repository.GetPromptById(id)
	if err != nil {
		return models.Prompt{}, err
	}

	err = db.Model(&prompt).Updates(command).Error
	if err != nil {
		return models.Prompt{}, err
	}

	return prompt, nil
}

func (repository PromptRepository) isValidColumn(orderBy string) bool {
	return orderBy == "name" || orderBy == "description" || orderBy == "created_at" || orderBy == "updated_at"
}

func (repository PromptRepository) CreatePrompt(command commands.UpsertPromptCommand) (models.Prompt, error) {
	db := repository.GetDB()
	prompt := models.Prompt{
		Name:        command.Name,
		Description: command.Description,
		Prompt:      command.Prompt,
	}

	err := db.Create(&prompt).Error
	if err != nil {
		return models.Prompt{}, err
	}

	return prompt, nil
}
