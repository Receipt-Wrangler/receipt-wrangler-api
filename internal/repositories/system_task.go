package repositories

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
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

	if !isColumnNameValid(command.OrderBy) {
		return nil, 0, errors.New("invalid column name")
	}

	systemTaskTypesToGet := []models.SystemTaskType{
		models.QUICK_SCAN,
		models.RECEIPT_UPLOADED,
		models.RECEIPT_UPDATED,
		models.EMAIL_UPLOAD,
	}

	query := db.Model(&models.SystemTask{}).
		Omit("can_be_restarted").
		Where("type IN ?", systemTaskTypesToGet).
		Where("group_id IN ?", command.GroupIds).
		Not(db.Where("type = ? AND ran_by_user_id IS NULL", models.RECEIPT_UPLOADED))

	query.Count(&count)

	query = repository.Sort(query, command.OrderBy, command.SortDirection)
	query = query.Scopes(repository.Paginate(command.Page, command.PageSize))

	query.Find(&results)

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
		ReceiptId:              command.ReceiptId,
		GroupId:                command.GroupId,
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

func (repository SystemTaskRepository) AssociateSystemTaskToReceipt(receiptId uint, systemTaskId uint) error {
	db := repository.GetDB()
	return db.Model(&models.SystemTask{}).Where("id = ?", systemTaskId).Update("receipt_id", receiptId).Error
}
