package services

import (
	"errors"
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
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

func (repository ImportService) ImportConfigJson(command commands.ConfigImportCommand) error {
	commandConfig := command.Config
	txErr := repository.DB.Transaction(func(tx *gorm.DB) error {
		err := repository.importEmailSettings(tx, commandConfig.EmailSettings)
		if err != nil {
			return err
		}

		receiptProcessingSettings, err := repository.importAiSettings(tx,
			commandConfig.AiSettings)
		if err != nil {
			if err.Error() != INVALID_AI_SETTINGS_ERROR {
				return err
			}
		}

		return repository.importSystemSettings(tx, commandConfig, receiptProcessingSettings)
	})
	if txErr != nil {
		return txErr
	}

	return nil
}

func (repository ImportService) importAiSettings(tx *gorm.DB, aiSettings structs.AiSettings) (models.ReceiptProcessingSettings, error) {
	receiptSettingsRepository := repositories.NewReceiptProcessingSettings(tx)
	var prompt models.Prompt
	var defaultPromptCount int64

	err := tx.Model(models.Prompt{}).Where("name = ?", constants.DEFAULT_PROMPT_NAME).Count(&defaultPromptCount).Error
	if err != nil {
		return models.ReceiptProcessingSettings{}, err
	}

	if defaultPromptCount > 0 {
		tx.Model(models.Prompt{}).Where("name = ?", constants.DEFAULT_PROMPT_NAME).First(&prompt)
	} else {
		defaultPrompt, err := CreateDefaultPrompt()
		if err != nil {
			return models.ReceiptProcessingSettings{}, err
		}

		prompt = defaultPrompt
	}

	command := commands.UpsertReceiptProcessingSettingsCommand{
		Name:       "Imported Settings",
		AiType:     aiSettings.AiType,
		Url:        aiSettings.Url,
		Key:        aiSettings.Key,
		Model:      aiSettings.Model,
		NumWorkers: aiSettings.NumWorkers,
		OcrEngine:  aiSettings.OcrEngine,
		PromptId:   prompt.ID,
	}
	vErrs := command.Validate()
	if len(vErrs.Errors) > 0 {
		logging.GetLogger().Println("Unable to import invalid AI settings: ", vErrs.Errors)
		return models.ReceiptProcessingSettings{},
			errors.New(INVALID_AI_SETTINGS_ERROR)
	}

	createdSettings, err := receiptSettingsRepository.CreateReceiptProcessingSettings(command)
	if err != nil {
		return models.ReceiptProcessingSettings{}, err
	}

	return createdSettings, nil
}

func (repository ImportService) importEmailSettings(tx *gorm.DB, settings []structs.EmailSettings) error {
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
			logging.GetLogger().Println("Unable to import invalid email settings: ", vErr.Errors)
			return errors.New("invalid email settings")
		}

		systemEmail, err := systemEmailRepository.AddSystemEmail(systemEmailCommand)
		if err != nil {
			return err
		}

		createdSystemEmails = append(createdSystemEmails, systemEmail)
	}

	return repository.updateGroupSettings(tx, createdSystemEmails)
}

func (repository ImportService) updateGroupSettings(tx *gorm.DB, createdSystemEmails []models.SystemEmail) error {
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

				idString := simpleutils.UintToString(groupSetting.ID)
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

func (repository ImportService) importSystemSettings(tx *gorm.DB, config structs.Config, receiptProcessingSettings models.ReceiptProcessingSettings) error {
	systemSettingsRepository := repositories.NewSystemSettingsRepository(tx)
	_, err := systemSettingsRepository.GetSystemSettings()
	if err != nil {
		return err
	}

	emailPollingInterval := config.EmailPollingInterval
	if emailPollingInterval < 0 {
		emailPollingInterval = 1800
	}

	command := commands.UpsertSystemSettingsCommand{
		EnableLocalSignUp:           config.Features.EnableLocalSignUp,
		DebugOcr:                    config.Debug.DebugOcr,
		EmailPollingInterval:        emailPollingInterval,
		ReceiptProcessingSettingsId: &receiptProcessingSettings.ID,
	}

	_, err = systemSettingsRepository.UpdateSystemSettings(command)
	if err != nil {
		return err
	}

	return nil
}
