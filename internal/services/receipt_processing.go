package services

import (
	"encoding/json"
	"gopkg.in/gographics/imagick.v3/imagick"
	"gorm.io/gorm"
	"os"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strings"
	"time"
)

type ReceiptProcessingService struct {
	BaseService
	ReceiptProcessingSettings         models.ReceiptProcessingSettings
	FallbackReceiptProcessingSettings models.ReceiptProcessingSettings
	Group                             models.Group
}

func NewSystemReceiptProcessingService(tx *gorm.DB, groupId string) (ReceiptProcessingService, error) {
	systemSettingsRepository := repositories.NewSystemSettingsRepository(tx)
	systemReceiptProcessingSettings, err := systemSettingsRepository.GetSystemReceiptProcessingSettings()
	group := models.Group{}
	if err != nil {
		return ReceiptProcessingService{}, err
	}

	if len(groupId) > 0 {
		groupRepository := repositories.NewGroupRepository(tx)
		groupToUse, err := groupRepository.GetGroupById(groupId, false, true, false)
		if err != nil {
			return ReceiptProcessingService{}, err
		}

		group = groupToUse

		if groupToUse.GroupSettings.PromptId != nil && *groupToUse.GroupSettings.PromptId > 0 {
			systemReceiptProcessingSettings.ReceiptProcessingSettings.PromptId = *groupToUse.GroupSettings.PromptId
		}

		if groupToUse.GroupSettings.FallbackPromptId != nil &&
			*groupToUse.GroupSettings.FallbackPromptId > 0 &&
			systemReceiptProcessingSettings.FallbackReceiptProcessingSettings.ID != 0 {
			systemReceiptProcessingSettings.FallbackReceiptProcessingSettings.PromptId = *groupToUse.GroupSettings.FallbackPromptId
		}
	}

	return ReceiptProcessingService{
		BaseService:                       BaseService{TX: tx},
		ReceiptProcessingSettings:         systemReceiptProcessingSettings.ReceiptProcessingSettings,
		FallbackReceiptProcessingSettings: systemReceiptProcessingSettings.FallbackReceiptProcessingSettings,
		Group:                             group,
	}, nil

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

func (service ReceiptProcessingService) ReadReceiptImage(
	imagePath string,
) (commands.UpsertReceiptCommand, commands.ReceiptProcessingMetadata, error) {
	var receipt commands.UpsertReceiptCommand
	metadata := commands.ReceiptProcessingMetadata{}

	result, err := service.processImage(
		imagePath,
		service.ReceiptProcessingSettings,
	)
	metadata.OcrSystemTaskCommand = result.OcrSystemTaskCommand
	metadata.PromptSystemTaskCommand = result.PromptSystemTaskCommand
	metadata.ChatCompletionSystemTaskCommand = result.ChatCompletionSystemTaskCommand
	metadata.ReceiptProcessingSettingsIdRan = service.ReceiptProcessingSettings.ID
	if err != nil {
		metadata.DidReceiptProcessingSettingsSucceed = false
		metadata.RawResponse = err.Error()

		if service.FallbackReceiptProcessingSettings.ID > 0 {
			fallbackResult, fallbackErr := service.processImage(
				imagePath,
				service.FallbackReceiptProcessingSettings,
			)
			metadata.FallbackReceiptProcessingSettingsIdRan = service.FallbackReceiptProcessingSettings.ID
			metadata.FallbackOcrSystemTaskCommand = fallbackResult.OcrSystemTaskCommand
			metadata.FallbackPromptSystemTaskCommand = fallbackResult.PromptSystemTaskCommand
			metadata.FallbackChatCompletionSystemTaskCommand = fallbackResult.ChatCompletionSystemTaskCommand
			receipt = fallbackResult.Receipt
			err = fallbackErr

			if err != nil {
				metadata.DidFallbackReceiptProcessingSettingsSucceed = false
				metadata.FallbackRawResponse = err.Error()
			} else {
				metadata.DidFallbackReceiptProcessingSettingsSucceed = true
				metadata.FallbackRawResponse = fallbackResult.RawResponse
			}
		}
	} else {
		metadata.DidReceiptProcessingSettingsSucceed = true
		metadata.RawResponse = result.RawResponse
		receipt = result.Receipt
	}

	return receipt, metadata, err
}

func (service ReceiptProcessingService) processImage(
	imagePath string,
	receiptProcessingSettings models.ReceiptProcessingSettings,
) (commands.ReceiptProcessingResult, error) {
	aiMessages := []structs.AiClientMessage{}
	receipt := commands.UpsertReceiptCommand{}
	result := commands.ReceiptProcessingResult{}
	ocrText := ""
	base64Image := ""

	if receiptProcessingSettings.IsVisionModel {
		if receiptProcessingSettings.AiType == models.OLLAMA {
			ollamaImage, err := service.getOllamaBase64Image(imagePath)
			if err != nil {
				return result, err
			}

			base64Image = ollamaImage
		}

		if receiptProcessingSettings.AiType == models.OPEN_AI_NEW || receiptProcessingSettings.AiType == models.OPEN_AI_CUSTOM {
			openAiImage, err := service.getOpenAiBase64Image(imagePath)
			if err != nil {
				return result, err
			}

			base64Image = openAiImage
		}

		if receiptProcessingSettings.AiType == models.GEMINI_NEW {
			geminiImage, err := service.getGeminiImage(imagePath)
			if err != nil {
				return result, err
			}

			base64Image = geminiImage
		}
	} else {
		ocrService := NewOcrService(service.TX, receiptProcessingSettings)
		resultText, ocrSystemTaskCommand, err := ocrService.ReadImage(imagePath)
		result.OcrSystemTaskCommand = ocrSystemTaskCommand
		if err != nil {
			return result, err
		}

		ocrText = resultText
	}

	prompt, promptSystemTask, err := service.buildPrompt(receiptProcessingSettings, ocrText)
	result.PromptSystemTaskCommand = promptSystemTask
	if err != nil {
		return result, err
	}

	message := structs.AiClientMessage{
		Role:    "user",
		Content: prompt,
	}
	if len(base64Image) > 0 {
		message.Images = []string{base64Image}
	}

	aiMessages = append(aiMessages, message)

	aiClient := AiService{
		ReceiptProcessingSettings: receiptProcessingSettings,
	}

	response, chatCompletionSystemTaskCommand, err := aiClient.CreateChatCompletion(structs.AiChatCompletionOptions{
		Messages:   aiMessages,
		DecryptKey: true,
	})
	result.ChatCompletionSystemTaskCommand = chatCompletionSystemTaskCommand
	result.RawResponse = response
	if err != nil {
		return result, err
	}

	cleanedResponse := service.cleanResponse(response)

	err = json.Unmarshal([]byte(cleanedResponse), &receipt)
	if err != nil {
		return result, err
	}

	result.Receipt = receipt
	return result, nil
}

func (service ReceiptProcessingService) cleanResponse(response string) string {
	response = strings.ReplaceAll(response, "```json", "")
	response = strings.ReplaceAll(response, "```", "")
	return response
}

// TODO: move to new ai client
func (service ReceiptProcessingService) getOllamaBase64Image(imagePath string) (string, error) {
	mw := imagick.NewMagickWand()
	err := mw.ReadImage(imagePath)
	if err != nil {
		return "", err
	}

	fileBytes, err := mw.GetImageBlob()
	if err != nil {
		return "", err
	}

	return utils.Base64Encode(fileBytes), nil
}

func (service ReceiptProcessingService) getOpenAiBase64Image(imagePath string) (string, error) {
	fileRepository := repositories.NewFileRepository(service.TX)
	fileBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return "", err
	}

	uri, err := fileRepository.BuildEncodedImageString(fileBytes)
	if err != nil {
		return "", err
	}

	return uri, nil
}

func (service ReceiptProcessingService) getGeminiImage(imagePath string) (string, error) {
	fileBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return "", err
	}

	return utils.Base64Encode(fileBytes), nil
}

func (service ReceiptProcessingService) buildPrompt(
	receiptProcessingSettings models.ReceiptProcessingSettings,
	ocrText string,
) (string, commands.UpsertSystemTaskCommand, error) {
	systemTaskCommand := commands.UpsertSystemTaskCommand{
		Type:                 models.PROMPT_GENERATED,
		Status:               models.SYSTEM_TASK_SUCCEEDED,
		AssociatedEntityType: models.PROMPT,
		AssociatedEntityId:   receiptProcessingSettings.PromptId,
		StartedAt:            time.Now(),
	}

	promptRepository := repositories.NewPromptRepository(service.TX)

	stringPromptId := utils.UintToString(receiptProcessingSettings.PromptId)

	prompt, err := promptRepository.GetPromptById(stringPromptId)
	if err != nil {
		systemTaskCommand.Status = models.SYSTEM_TASK_FAILED
		systemTaskCommand.ResultDescription = err.Error()
		endedAt := time.Now()
		systemTaskCommand.EndedAt = &endedAt

		return "", systemTaskCommand, err
	}

	templateVariableMap, err := service.buildTemplateVariableMap(ocrText)
	if err != nil {
		systemTaskCommand.Status = models.SYSTEM_TASK_FAILED
		systemTaskCommand.ResultDescription = err.Error()
		endedAt := time.Now()
		systemTaskCommand.EndedAt = &endedAt

		return "", systemTaskCommand, err
	}

	regex := utils.GetTriggerRegex()
	realPrompt := regex.ReplaceAllStringFunc(prompt.Prompt, func(variable string) string {
		templateVariable := structs.PromptTemplateVariable(variable)
		return templateVariableMap[templateVariable]
	})

	endedAt := time.Now()
	systemTaskCommand.EndedAt = &endedAt
	systemTaskCommand.ResultDescription = realPrompt

	return realPrompt, systemTaskCommand, nil
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

	currentYearString := utils.UintToString(uint(time.Now().Year()))

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
