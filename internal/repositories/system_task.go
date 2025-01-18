package repositories

import (
	"errors"
	"gorm.io/gorm/clause"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/wranglerasynq"

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

func (repository SystemTaskRepository) GetPagedSystemTasks(command commands.GetSystemTaskCommand) ([]models.SystemTask, int64, error) {
	db := repository.GetDB()
	var results []models.SystemTask
	var count int64

	if !isColumnNameValid(command.OrderBy) {
		return nil, 0, errors.New("invalid column name")
	}

	filteredSystemTaskTypes := []models.SystemTaskType{
		models.RECEIPT_UPLOADED,
		models.CHAT_COMPLETION,
		models.OCR_PROCESSING,
	}

	query := db.Model(&models.SystemTask{}).Where("type NOT IN ?", filteredSystemTaskTypes)

	if command.AssociatedEntityId != 0 {
		query = query.Where("associated_entity_id = ?", command.AssociatedEntityId)
	}

	if len(command.AssociatedEntityType) > 0 {
		query = query.Where("associated_entity_type = ?", command.AssociatedEntityType)
	}

	query.Count(&count)

	query = repository.Sort(query, command.OrderBy, command.SortDirection)
	query = query.Scopes(repository.Paginate(command.Page, command.PageSize))

	err := query.Preload(clause.Associations).Preload("ChildSystemTasks.ChildSystemTasks").Find(&results).Error
	if query.Error != nil {
		return nil, 0, err
	}

	return results, count, nil
}

func (repository SystemTaskRepository) GetPagedActivities(command commands.PagedActivityRequestCommand) (
	[]structs.Activity,
	int64,
	error,
) {
	db := repository.GetDB()
	var results []structs.Activity
	var count int64

	// TODO: implement rerun
	if !isColumnNameValid(command.OrderBy) {
		return nil, 0, errors.New("invalid column name")
	}

	filteredSystemTaskTypes := []models.SystemTaskType{
		models.MAGIC_FILL,
		models.SYSTEM_EMAIL_CONNECTIVITY_CHECK,
		models.RECEIPT_PROCESSING_SETTINGS_CONNECTIVITY_CHECK,
		models.META_ASSOCIATE_TASKS_TO_RECEIPT,
	}

	query := db.Model(&models.SystemTask{}).
		Distinct().
		Joins("LEFT JOIN users ON system_tasks.ran_by_user_id = users.id").
		Joins("LEFT JOIN group_members ON users.id = group_members.user_id").
		Joins("LEFT JOIN receipts ON system_tasks.associated_entity_type = 'RECEIPT' "+
			"AND system_tasks.associated_entity_id = receipts.id").
		Where("(group_members.group_id IN ?) OR receipts.group_id IN ?", command.GroupIds, command.GroupIds).
		Where("system_tasks.type NOT IN ?", filteredSystemTaskTypes)

	query.Count(&count)

	query = repository.Sort(query, command.OrderBy, command.SortDirection)
	query = query.Scopes(repository.Paginate(command.Page, command.PageSize))

	query.Find(&results)

	inspector, err := wranglerasynq.GetAsynqInspector()
	if err != nil {
		return nil, 0, err
	}
	archivedTasks, err := inspector.ListArchivedTasks(string(models.QuickScanQueue))

	systemTaskRepository := NewSystemTaskRepository(nil)

	for _, activity := range results {
		if activity.Type == models.QUICK_SCAN && activity.AssociatedSystemTaskId != nil {
			associatedSystemTask, err := systemTaskRepository.GetSystemTaskById(*activity.AssociatedSystemTaskId)
			if err != nil {
				return nil, 0, err
			}
			for i := 0; i < len(archivedTasks); i++ {
				task := archivedTasks[i]
				if task.ID == associatedSystemTask.AsynqTaskId {
					activity.CanBeRestarted = true
					break
				}
			}
		}
	}

	return results, count, nil
}

func isColumnNameValid(columnName string) bool {
	return columnName == "type" || columnName == "status" || columnName == "associated_entity_type" || columnName == "associated_entity_id" || columnName == "started_at" || columnName == "ended_at" || columnName == "result_description" || columnName == "ran_by_user_id"
}

func (repository SystemTaskRepository) CreateSystemTask(command commands.UpsertSystemTaskCommand) (models.SystemTask, error) {
	db := repository.GetDB()

	systemTask := models.SystemTask{
		Type:                   command.Type,
		Status:                 command.Status,
		AssociatedEntityType:   command.AssociatedEntityType,
		AssociatedEntityId:     command.AssociatedEntityId,
		StartedAt:              command.StartedAt,
		EndedAt:                command.EndedAt,
		ResultDescription:      command.ResultDescription,
		RanByUserId:            command.RanByUserId,
		AssociatedSystemTaskId: command.AssociatedSystemTaskId,
		AsynqTaskId:            command.AsynqTaskId,
	}

	err := db.Create(&systemTask).Error
	if err != nil {
		return models.SystemTask{}, err
	}

	if command.AssociatedSystemTaskId != nil && systemTask.Status == models.SYSTEM_TASK_FAILED {
		var parentSystemTask models.SystemTask
		db.Model(&models.SystemTask{}).Where("id = ?", command.AssociatedSystemTaskId).Find(&parentSystemTask)

		if parentSystemTask.Status == models.SYSTEM_TASK_SUCCEEDED {
			db.Model(&parentSystemTask).Update("status", models.SYSTEM_TASK_FAILED)
		}

	}

	return systemTask, nil
}

func (repository SystemTaskRepository) DeleteSystemTaskByAssociatedEntityId(
	associatedEntityId string,
	emailType models.AssociatedEntityType,
) error {
	db := repository.GetDB()
	err := db.Where("associated_entity_id = ? and associated_entity_type = ?", associatedEntityId, emailType).Delete(&models.SystemTask{}).Error
	if err != nil {
		return err
	}

	return nil
}

func (repository SystemTaskRepository) GetSystemTaskById(id uint) (models.SystemTask, error) {
	db := repository.GetDB()
	var systemTask models.SystemTask

	err := db.Model(&models.SystemTask{}).Where("id = ?", id).First(&systemTask).Error
	if err != nil {
		return models.SystemTask{}, err
	}

	return systemTask, nil
}
