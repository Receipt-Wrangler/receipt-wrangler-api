package services

import (
	"fmt"
	"receipt-wrangler/api/internal/ai"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"time"
)

func NewAiService(receiptProcessingSettingsId string) (*AiService, error) {
	repository := repositories.NewReceiptProcessingSettings(nil)
	client := &AiService{}

	receiptProcessingSettings, err := repository.GetReceiptProcessingSettingsById(receiptProcessingSettingsId)
	if err != nil {
		return nil, err
	}
	client.ReceiptProcessingSettings = receiptProcessingSettings

	return client, nil
}

type AiService struct {
	ReceiptProcessingSettings models.ReceiptProcessingSettings
}

func (service *AiService) CreateChatCompletion(options structs.AiChatCompletionOptions) (string, commands.UpsertSystemTaskCommand, error) {
	systemTask := commands.UpsertSystemTaskCommand{
		Type:                 models.CHAT_COMPLETION,
		Status:               models.SYSTEM_TASK_SUCCEEDED,
		AssociatedEntityType: models.RECEIPT_PROCESSING_SETTINGS,
		AssociatedEntityId:   service.ReceiptProcessingSettings.ID,
		StartedAt:            time.Now(),
	}
	response, rawResponse := "", ""
	err := error(nil)
	var client ai.Client

	switch service.ReceiptProcessingSettings.AiType {
	case models.OPEN_AI_CUSTOM_NEW:
		client = ai.NewOpenAiClient(options, service.ReceiptProcessingSettings)

	case models.OPEN_AI_NEW:
		client = ai.NewOpenAiClient(options, service.ReceiptProcessingSettings)

	case models.OLLAMA:
		client = ai.NewOllamaClient(options, service.ReceiptProcessingSettings)

	case models.GEMINI_NEW:
		client = ai.NewGeminiClient(options, service.ReceiptProcessingSettings)

	default:
		return "", systemTask, fmt.Errorf("invalid ai type: %s", service.ReceiptProcessingSettings.AiType)
	}

	result, err := client.GetChatCompletion()
	if err != nil {
		endedAt := time.Now()
		systemTask.Status = models.SYSTEM_TASK_FAILED
		systemTask.ResultDescription = fmt.Sprintf("Error: %s; RawResponse: %s", err.Error(), rawResponse)
		systemTask.EndedAt = &endedAt

		return "", systemTask, err
	}
	response = result.Response
	rawResponse = result.RawResponse

	endedAt := time.Now()
	systemTask.ResultDescription = rawResponse
	systemTask.EndedAt = &endedAt

	return response, systemTask, nil
}

func (service *AiService) CheckConnectivity(ranByUserId uint, decryptKey bool) (models.SystemTask, error) {
	messages := []structs.AiClientMessage{
		{
			Role:    "user",
			Content: "Respond with 'hello' if you are there!",
		},
	}

	systemTaskCommand := commands.UpsertSystemTaskCommand{
		Type:        models.RECEIPT_PROCESSING_SETTINGS_CONNECTIVITY_CHECK,
		RanByUserId: &ranByUserId,
	}

	startedAt := time.Now()
	response, _, err := service.CreateChatCompletion(structs.AiChatCompletionOptions{
		Messages:   messages,
		DecryptKey: decryptKey,
	})
	if err != nil {
		systemTaskCommand.Status = models.SYSTEM_TASK_FAILED
		systemTaskCommand.ResultDescription = err.Error()
	} else {
		systemTaskCommand.Status = models.SYSTEM_TASK_SUCCEEDED
		systemTaskCommand.ResultDescription = fmt.Sprintf(
			"The configured model responded with: %s in response to: %s",
			response, messages[0].Content)
	}
	endedAt := time.Now()

	systemTaskCommand.StartedAt = startedAt
	systemTaskCommand.EndedAt = &endedAt

	if service.ReceiptProcessingSettings.ID > 0 {
		systemTaskCommand.AssociatedEntityId = service.ReceiptProcessingSettings.ID
		systemTaskCommand.AssociatedEntityType = models.RECEIPT_PROCESSING_SETTINGS

		systemTaskRepository := repositories.NewSystemTaskRepository(nil)
		return systemTaskRepository.CreateSystemTask(systemTaskCommand)
	}

	return models.SystemTask{
		Status: systemTaskCommand.Status,
	}, nil
}
