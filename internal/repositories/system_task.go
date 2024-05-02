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

// TODO: Validate the column name
func (repository SystemTaskRepository) GetPagedSystemTasks(command commands.GetSystemTaskCommand) ([]models.SystemTask, error) {
	db := repository.GetDB()
	var results []models.SystemTask

	query := db.Model(&models.SystemTask{})

	if command.AssociatedEntityId != 0 {
		query = query.Where("associated_entity_id = ?", command.AssociatedEntityId)
	}

	if len(command.AssociatedEntityType) > 0 {
		query = query.Where("associated_entity_type = ?", command.AssociatedEntityType)
	}

	query = repository.Sort(query, command.OrderBy, command.SortDirection)
	query = query.Scopes(repository.Paginate(command.Page, command.PageSize))

	err := query.Find(&results).Error
	if query.Error != nil {
		return nil, err
	}

	return results, nil
}

func (repository SystemTaskRepository) CreateSystemTask(command commands.UpsertSystemTaskCommand) (models.SystemTask, error) {
	db := repository.GetDB()

	systemTask := models.SystemTask{
		Type:                 command.Type,
		Status:               command.Status,
		AssociatedEntityType: command.AssociatedEntityType,
		AssociatedEntityId:   command.AssociatedEntityId,
		StartedAt:            command.StartedAt,
		EndedAt:              command.EndedAt,
		ResultDescription:    command.ResultDescription,
	}

	err := db.Create(&systemTask).Error
	if err != nil {
		return models.SystemTask{}, err
	}

	return systemTask, nil
}
