package services

import (
	"encoding/json"
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/ai"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"time"
)

type ReceiptProcessingService struct {
	BaseService
	ReceiptProcessingSettings         models.ReceiptProcessingSettings
	FallbackReceiptProcessingSettings models.ReceiptProcessingSettings
}

func NewReceiptProcessingService(tx *gorm.DB, receiptProcessingSettingsId string, fallbackReceiptProcessingSettingsId string) (ReceiptProcessingService, error) {
	service := ReceiptProcessingService{BaseService: BaseService{
		DB: repositories.GetDB(),
		TX: tx,
	}}

	receiptProcessingSettingsRepository := repositories.NewReceiptProcessingSettings(nil)
	receiptProcessingSettings, err := receiptProcessingSettingsRepository.GetReceiptProcessingSettingsById(receiptProcessingSettingsId)
	if err != nil {
		return service, err
	}
	service.ReceiptProcessingSettings = receiptProcessingSettings

	if len(fallbackReceiptProcessingSettingsId) > 0 && fallbackReceiptProcessingSettingsId != "0" {
		fallbackReceiptProcessingSettings, err := receiptProcessingSettingsRepository.GetReceiptProcessingSettingsById(fallbackReceiptProcessingSettingsId)
		if err != nil {
			return service, err
		}
		service.FallbackReceiptProcessingSettings = fallbackReceiptProcessingSettings
	}

	return service, nil
}

func (service ReceiptProcessingService) ReadReceiptImage(imagePath string) (models.Receipt, error) {
	var receipt models.Receipt

	receipt, err := service.processImage(imagePath, service.ReceiptProcessingSettings)
	if err != nil {
		if service.FallbackReceiptProcessingSettings.ID > 0 {
			receipt, err = service.processImage(imagePath, service.FallbackReceiptProcessingSettings)
			if err != nil {
				return models.Receipt{}, err
			}
		}
	}

	return receipt, err
}

func (service ReceiptProcessingService) processImage(imagePath string, receiptProcessingSettings models.ReceiptProcessingSettings) (models.Receipt, error) {
	aiMessages := []structs.AiClientMessage{}
	receipt := models.Receipt{}

	ocrService := NewOcrService(service.TX, receiptProcessingSettings)
	ocrText, err := ocrService.ReadImage(imagePath)
	if err != nil {
		return models.Receipt{}, err
	}

	prompt, err := service.buildPrompt(receiptProcessingSettings, ocrText)
	if err != nil {
		return models.Receipt{}, err
	}

	aiMessages = append(aiMessages, structs.AiClientMessage{
		Role:    "user",
		Content: prompt,
	})

	aiClient := ai.AiClientNew{
		ReceiptProcessingSettings: receiptProcessingSettings,
	}

	response, err := aiClient.CreateChatCompletion(aiMessages, true)
	if err != nil {
		return models.Receipt{}, err
	}

	err = json.Unmarshal([]byte(response), &receipt)
	if err != nil {
		return models.Receipt{}, err
	}

	return receipt, nil
}

func (service ReceiptProcessingService) buildPrompt(receiptProcessingSettings models.ReceiptProcessingSettings, ocrText string) (string, error) {
	promptRepository := repositories.NewPromptRepository(service.TX)

	stringPromptId := simpleutils.UintToString(receiptProcessingSettings.PromptId)

	prompt, err := promptRepository.GetPromptById(stringPromptId)
	if err != nil {
		return "", err
	}

	templateVariableMap, err := service.buildTemplateVariableMap(ocrText)
	if err != nil {
		return "", err
	}

	regex := utils.GetTriggerRegex()
	realPrompt := regex.ReplaceAllStringFunc(prompt.Prompt, func(variable string) string {
		templateVariable := structs.PromptTemplateVariable(variable)
		return templateVariableMap[templateVariable]
	})

	return realPrompt, nil
}

func (service ReceiptProcessingService) buildTemplateVariableMap(ocrText string) (map[structs.PromptTemplateVariable]string, error) {
	result := make(map[structs.PromptTemplateVariable]string)

	categoriesString, err := service.getCategoriesString()
	if err != nil {
		return result, err
	}

	tagsString, err := service.getTagsString()
	if err != nil {
		return result, err
	}

	currentYearString := simpleutils.UintToString(uint(time.Now().Year()))

	result[structs.CATEGORIES] = categoriesString
	result[structs.TAGS] = tagsString
	result[structs.OCR_TEXT] = ocrText
	result[structs.CURRENT_YEAR] = currentYearString

	return result, nil
}

func (service ReceiptProcessingService) getCategoriesString() (string, error) {
	categoryRepository := repositories.NewCategoryRepository(nil)
	categories, err := categoryRepository.GetAllCategories("id, name, description")
	if err != nil {
		return "", err
	}

	categoriesBytes, err := json.Marshal(categories)
	if err != nil {
		return "", err
	}

	return string(categoriesBytes), nil
}

func (service ReceiptProcessingService) getTagsString() (string, error) {
	tagsRepository := repositories.NewTagsRepository(nil)
	tags, err := tagsRepository.GetAllTags("id, name")
	if err != nil {
		return "", err
	}

	tagsBytes, err := json.Marshal(tags)
	if err != nil {
		return "", err
	}

	return string(tagsBytes), nil
}
