package repositories

import (
	"errors"
	"fmt"
	"receipt-wrangler/api/internal/commands"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"

	"gorm.io/gorm"
)

type SystemEmailRepository struct {
	BaseRepository
}

func NewSystemEmailRepository(tx *gorm.DB) SystemEmailRepository {
	repository := SystemEmailRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository SystemEmailRepository) GetPagedSystemEmails(command commands.PagedRequestCommand) ([]models.SystemEmail, error) {
	db := repository.GetDB()
	var systemEmails []models.SystemEmail
	query := db.Model(models.SystemEmail{})
	if !isValidColumn(command.OrderBy) {
		return nil, errors.New(fmt.Sprint("Invalid column name: ", command.OrderBy))
	}
	query = repository.Sort(query, command.OrderBy, command.SortDirection)
	query = query.Scopes(repository.Paginate(command.Page, command.PageSize))

	err := db.Model(models.SystemEmail{}).Find(&systemEmails).Error
	if err != nil {
		return nil, err
	}

	return systemEmails, nil
}

func (repository SystemEmailRepository) GetSystemEmailById(id string) (models.SystemEmail, error) {
	db := repository.GetDB()
	var systemEmail models.SystemEmail

	err := db.Model(models.SystemEmail{}).Where("id = ?", id).First(&systemEmail).Error
	if err != nil {
		return models.SystemEmail{}, err
	}

	return systemEmail, nil
}

func (repository SystemEmailRepository) AddSystemEmail(command commands.UpsertSystemEmailCommand) (models.SystemEmail, error) {
	db := repository.GetDB()

	rawEncryptedPassword, err := utils.EncryptData(config.GetEncryptionKey(), []byte(command.Password))
	if err != nil {
		return models.SystemEmail{}, err
	}

	systemEmail := models.SystemEmail{
		Host:     command.Host,
		Port:     command.Port,
		Username: command.Username,
		Password: utils.EncodeToBase64(rawEncryptedPassword),
	}

	err = db.Create(&systemEmail).Error
	if err != nil {
		return models.SystemEmail{}, err
	}

	return systemEmail, nil
}

func (repository SystemEmailRepository) UpdateSystemEmail(id string, command commands.UpsertSystemEmailCommand, updatePassword bool) (models.SystemEmail, error) {
	db := repository.GetDB()

	currentSystemEmail, err := repository.GetSystemEmailById(id)
	if err != nil {
		return models.SystemEmail{}, err
	}

	action := db.Model(&currentSystemEmail)

	if updatePassword {
		rawEncryptedPassword, err := utils.EncryptData(config.GetEncryptionKey(), []byte(command.Password))
		if err != nil {
			return models.SystemEmail{}, err
		}
		command.Password = utils.EncodeToBase64(rawEncryptedPassword)
	} else {
		action.Omit("password")
	}

	err = action.Updates(command).Error
	if err != nil {
		return models.SystemEmail{}, err
	}

	return currentSystemEmail, nil
}

func (repository SystemEmailRepository) DeleteSystemEmail(id string) error {
	db := repository.GetDB()

	txErr := db.Transaction(func(tx *gorm.DB) error {
		taskRepository := NewSystemTaskRepository(tx)
		repository.SetTransaction(tx)

		err := taskRepository.DeleteSystemTaskByAssociatedEntityId(id)
		if err != nil {
			return err
		}

		err = tx.Delete(&models.SystemEmail{}, id).Error
		if err != nil {
			return err
		}

		repository.ClearTransaction()
		return nil
	})

	if txErr != nil {
		return txErr
	}

	return nil
}

func (repository SystemEmailRepository) GetSystemTasksForSystemEmail(id string) ([]models.SystemTask, error) {
	db := repository.GetDB()
	var systemTasks []models.SystemTask

	err := db.Model(models.SystemTask{}).Where("associated_entity_id = ? AND associated_entity_type = ?", id, models.SYSTEM_EMAIL).Find(&systemTasks).Error
	if err != nil {
		return nil, err
	}

	return systemTasks, nil
}

func isValidColumn(columnName string) bool {
	columnNames := []string{"username", "host", "created_at", "updated_at"}
	for _, name := range columnNames {
		if name == columnName {
			return true
		}
	}

	return false
}
