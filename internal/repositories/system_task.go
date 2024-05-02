package repositories

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"

	"gorm.io/gorm"
)

type SystemTaskRepository struct {
	BaseRepository
}

func NewSystemTaskRepository(tx *gorm.DB) SystemTaskRepository {
	repository := SystemTaskRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository SystemTaskRepository) GetSystemTasks(command commands.GetSystemTaskCommand) ([]models.SystemTask, error) {
	db := repository.GetDB()
	var results []models.SystemTask

	query := db.Model(&models.SystemTask{})

	if command.AssociatedEntityId != 0 {
		query = query.Where("associated_entity_id = ?", command.AssociatedEntityId)
	}

	if len(command.AssociatedEntityType) > 0 {
		query = query.Where("associated_entity_type = ?", command.AssociatedEntityType)
	}

	if command.Count != 0 {
		query = query.Limit(command.Count)
	}

	err := query.Find(&results).Error
	if query.Error != nil {
		return nil, err
	}

	return results, nil
}
