package services

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strconv"
)

const INVALID_AI_SETTINGS_ERROR = "invalid ai settings"

type ImportService struct {
	BaseService
}

func NewImportService(tx *gorm.DB) ImportService {
	service := ImportService{
		BaseService: BaseService{
			DB: repositories.GetDB(),
			TX: tx,
		},
	}

	return service
}

func (service ImportService) ImportConfigJson(command commands.ConfigImportCommand) error {
	commandConfig := command.Config
	receiptProcessingSettings := models.ReceiptProcessingSettings{}
	db := service.GetDB()

	txErr := db.Transaction(func(tx *gorm.DB) error {
		err := service.importEmailSettings(tx, commandConfig.EmailSettings)
		if err != nil {
			return err
		}

		if len(commandConfig.AiSettings.AiType) > 0 {
			receiptProcessingSettings, err = service.importAiSettings(tx,
				commandConfig.AiSettings)
			if err != nil {
				return err
			}
		}

		return service.importSystemSettings(tx, commandConfig, receiptProcessingSettings)
	})
	if txErr != nil {
		return txErr
	}

	return nil
}

func (service ImportService) importAiSettings(tx *gorm.DB, aiSettings structs.AiSettings) (models.ReceiptProcessingSettings, error) {
	receiptSettingsRepository := repositories.NewReceiptProcessingSettings(tx)
	var prompt models.Prompt
	var defaultPromptCount int64

	err := tx.Model(models.Prompt{}).Where("name = ?", constants.DefaultPromptName).Count(&defaultPromptCount).Error
	if err != nil {
		return models.ReceiptProcessingSettings{}, err
	}

	if defaultPromptCount > 0 {
		err := tx.Model(models.Prompt{}).Where("name = ?", constants.DefaultPromptName).First(&prompt).Error
		if err != nil {
			return models.ReceiptProcessingSettings{}, err
		}
	} else {
		promptService := NewPromptService(tx)
		defaultPrompt, err := promptService.CreateDefaultPrompt()
		if err != nil {
			return models.ReceiptProcessingSettings{}, err
		}

		prompt = defaultPrompt
	}

	ocrEngine := aiSettings.OcrEngine
	aiType := aiSettings.AiType
	numWorkers := aiSettings.NumWorkers

	if aiSettings.OcrEngine == models.TESSERACT {
		ocrEngine = models.TESSERACT_NEW
	}

	if aiSettings.OcrEngine == models.EASY_OCR {
		ocrEngine = models.EASY_OCR_NEW
	}

	if len(aiSettings.OcrEngine) == 0 {
		ocrEngine = models.TESSERACT_NEW
	}

	if aiSettings.AiType == models.OPEN_AI {
		aiType = models.OPEN_AI_NEW
	}

	if aiSettings.AiType == models.OPEN_AI_CUSTOM {
		aiType = models.OPEN_AI_CUSTOM_NEW
	}

	if aiSettings.AiType == models.GEMINI {
		aiType = models.GEMINI_NEW
	}

	if numWorkers == 0 {
		numWorkers = 1
	}

	var importedSettingsCount int64
	err = tx.Model(models.ReceiptProcessingSettings{}).Where("name LIKE  ?", "Imported Settings%").Count(&importedSettingsCount).Error

	settingsName := "Imported Settings"
	if importedSettingsCount > 0 {
		settingsName = fmt.Sprintf("Imported Settings (%d)", importedSettingsCount+1)

	}

	command := commands.UpsertReceiptProcessingSettingsCommand{
		Name:      settingsName,
		AiType:    aiType,
		Url:       aiSettings.Url,
		Key:       aiSettings.Key,
		Model:     aiSettings.Model,
		OcrEngine: ocrEngine,
		PromptId:  prompt.ID,
	}
	vErrs := command.Validate(false)
	if len(vErrs.Errors) > 0 {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Unable to import invalid AI settings: ", vErrs.Errors)
		return models.ReceiptProcessingSettings{},
			errors.New(INVALID_AI_SETTINGS_ERROR)
	}

	createdSettings, err := receiptSettingsRepository.CreateReceiptProcessingSettings(command)
	if err != nil {
		return models.ReceiptProcessingSettings{}, err
	}

	return createdSettings, nil
}

func (service ImportService) importEmailSettings(tx *gorm.DB, settings []structs.EmailSettings) error {
	systemEmailRepository := repositories.NewSystemEmailRepository(tx)

	if len(settings) == 0 {
		return nil
	}

	createdSystemEmails := make([]models.SystemEmail, 0)
	for _, emailSetting := range settings {
		portNumber := strconv.FormatInt(int64(emailSetting.Port), 10)

		systemEmailCommand := commands.UpsertSystemEmailCommand{
			Host:     emailSetting.Host,
			Port:     portNumber,
			Username: emailSetting.Username,
			Password: emailSetting.Password,
		}
		vErr := systemEmailCommand.Validate(true)
		if len(vErr.Errors) > 0 {
			logging.LogStd(logging.LOG_LEVEL_ERROR, "Unable to import invalid email settings: ", vErr.Errors)
			return errors.New("invalid email settings")
		}

		systemEmail, err := systemEmailRepository.AddSystemEmail(systemEmailCommand)
		if err != nil {
			return err
		}

		createdSystemEmails = append(createdSystemEmails, systemEmail)
	}

	return service.updateGroupSettings(tx, createdSystemEmails)
}

func (service ImportService) updateGroupSettings(tx *gorm.DB, createdSystemEmails []models.SystemEmail) error {
	groupSettingRepository := repositories.NewGroupSettingsRepository(tx)

	groupSettings := make([]models.GroupSettings, 0)
	err := tx.Model(models.GroupSettings{}).Find(&groupSettings).Error
	if err != nil {
		return err
	}

	for _, groupSetting := range groupSettings {
		for _, createdSystemEmail := range createdSystemEmails {
			if groupSetting.EmailToRead == createdSystemEmail.Host {
				command := commands.UpdateGroupSettingsCommand{
					SystemEmailId:               &createdSystemEmail.ID,
					EmailIntegrationEnabled:     groupSetting.EmailIntegrationEnabled,
					SubjectLineRegexes:          groupSetting.SubjectLineRegexes,
					EmailWhiteList:              groupSetting.EmailWhiteList,
					EmailDefaultReceiptStatus:   groupSetting.EmailDefaultReceiptStatus,
					EmailDefaultReceiptPaidById: groupSetting.EmailDefaultReceiptPaidById,
				}

				idString := utils.UintToString(groupSetting.ID)
				_, err := groupSettingRepository.UpdateGroupSettings(idString, command)
				if err != nil {
					return err
				}
			}
		}
	}

	err = tx.Model(models.GroupSettings{}).Where("system_email_id = ?", 0).Update("email_integration_enabled", false).Error
	if err != nil {
		return err
	}

	return nil
}

func (service ImportService) importSystemSettings(tx *gorm.DB, config structs.Config, receiptProcessingSettings models.ReceiptProcessingSettings) error {
	systemSettingsRepository := repositories.NewSystemSettingsRepository(tx)
	_, err := systemSettingsRepository.GetSystemSettings()
	if err != nil {
		return err
	}

	numWorkersToUse := config.AiSettings.NumWorkers
	if config.AiSettings.NumWorkers < 0 {
		numWorkersToUse = 1
	}

	emailPollingInterval := config.EmailPollingInterval
	if emailPollingInterval < 0 {
		emailPollingInterval = 1800
	}

	command := commands.UpsertSystemSettingsCommand{
		EnableLocalSignUp:    config.Features.EnableLocalSignUp,
		DebugOcr:             config.Debug.DebugOcr,
		EmailPollingInterval: emailPollingInterval,
		NumWorkers:           numWorkersToUse,
	}

	if receiptProcessingSettings.ID != 0 {
		command.ReceiptProcessingSettingsId = &receiptProcessingSettings.ID
	}

	_, err = systemSettingsRepository.UpdateSystemSettings(command)
	if err != nil {
		return err
	}

	return nil
}
