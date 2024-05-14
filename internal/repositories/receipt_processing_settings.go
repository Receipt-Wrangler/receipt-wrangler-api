package repositories

import (
	"errors"
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/commands"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
)

type ReceiptProcessingSettingsRepository struct {
	BaseRepository
}

func NewReceiptProcessingSettings(tx *gorm.DB) ReceiptProcessingSettingsRepository {
	repository := ReceiptProcessingSettingsRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository ReceiptProcessingSettingsRepository) GetPagedReceiptProcessingSettings(command commands.PagedRequestCommand) ([]models.ReceiptProcessingSettings, int64, error) {
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

// TODO: in update, need to encrypt key on update

func (repository ReceiptProcessingSettingsRepository) CreateReceiptProcessingSettings(command commands.UpsertReceiptProcessingSettingsCommand) (models.ReceiptProcessingSettings, error) {
	db := repository.GetDB()
	var encryptedKey string
	if len(command.Key) > 0 {
		key, err := utils.EncryptAndEncodeToBase64(config.GetEncryptionKey(), command.Key)
		if err != nil {
			return models.ReceiptProcessingSettings{}, err
		}

		encryptedKey = key
	} else {
		encryptedKey = ""
	}

	settings := models.ReceiptProcessingSettings{
		Name:        command.Name,
		Description: command.Description,
		AiType:      command.AiType,
		Url:         command.Url,
		Key:         encryptedKey,
		Model:       command.Model,
		NumWorkers:  command.NumWorkers,
		OcrEngine:   command.OcrEngine,
		PromptId:    command.PromptId,
	}

	err := db.Create(&settings).Error
	if err != nil {
		return models.ReceiptProcessingSettings{}, err
	}

	return settings, nil
}

func (repository ReceiptProcessingSettingsRepository) GetReceiptProcessingSettingsById(id string) (models.ReceiptProcessingSettings, error) {
	db := repository.GetDB()
	var settings models.ReceiptProcessingSettings

	err := db.First(&settings, id).Error
	if err != nil {
		return models.ReceiptProcessingSettings{}, err
	}

	return settings, nil
}

func (repository ReceiptProcessingSettingsRepository) isValidColumn(orderBy string) bool {
	return orderBy == "name" || orderBy == "description" || orderBy == "ai_type" || orderBy == "ocr_engine" || orderBy == "created_at" || orderBy == "updated_at"
}
