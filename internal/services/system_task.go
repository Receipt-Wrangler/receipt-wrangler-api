package services

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
	"time"
)

type SystemTaskService struct {
	BaseService
}

func NewSystemTaskService(tx *gorm.DB) SystemTaskService {
	service := SystemTaskService{BaseService: BaseService{
		DB: repositories.GetDB(),
		TX: tx,
	}}
	return service
}

func (service SystemTaskService) BuildSuccessReceiptProcessResultDescription(metadata commands.ReceiptProcessingMetadata) string {
	receiptProcessingSettingsRepository := repositories.NewReceiptProcessingSettings(nil)
	idToUse := uint(0)

	if metadata.ReceiptProcessingSettingsIdRan > 0 && metadata.DidReceiptProcessingSettingsSucceed {
		idToUse = metadata.ReceiptProcessingSettingsIdRan
	} else if metadata.FallbackReceiptProcessingSettingsIdRan > 0 && metadata.DidFallbackReceiptProcessingSettingsSucceed {
		idToUse = metadata.FallbackReceiptProcessingSettingsIdRan
	}

	if idToUse == 0 {
		return ""
	}

	idToUseString := simpleutils.UintToString(idToUse)

	receiptProcessingSettings, err := receiptProcessingSettingsRepository.GetReceiptProcessingSettingsById(idToUseString)
	if err != nil {
		return ""
	}

	return fmt.Sprintf(
		"Receipt processing settings: %s were used, and the raw response generated was: %s",
		receiptProcessingSettings.Name, metadata.RawResponse,
	)
}

func (service SystemTaskService) CreateSystemTasksFromMetadata(metadata commands.ReceiptProcessingMetadata, startDate time.Time, endDate time.Time, taskType models.SystemTaskType, userId *uint, parentAssociatedSystemTaskId func(command commands.UpsertSystemTaskCommand) *uint) (structs.ReceiptProcessingSystemTasks, error) {
	systemTaskRepository := repositories.NewSystemTaskRepository(service.TX)
	result := structs.ReceiptProcessingSystemTasks{}

	if metadata.ReceiptProcessingSettingsIdRan > 0 {
		systemTask := commands.UpsertSystemTaskCommand{
			Type:                 taskType,
			Status:               service.BoolToSystemTaskStatus(metadata.DidReceiptProcessingSettingsSucceed),
			AssociatedEntityType: models.RECEIPT_PROCESSING_SETTINGS,
			AssociatedEntityId:   metadata.ReceiptProcessingSettingsIdRan,
			StartedAt:            startDate,
			EndedAt:              &endDate,
			ResultDescription:    metadata.RawResponse,
			RanByUserId:          userId,
		}

		if parentAssociatedSystemTaskId != nil {
			systemTask.AssociatedSystemTaskId = parentAssociatedSystemTaskId(systemTask)
		}

		createdSystemTask, err := systemTaskRepository.CreateSystemTask(systemTask)
		if err != nil {
			return structs.ReceiptProcessingSystemTasks{}, err
		}
		result.SystemTask = createdSystemTask

		_, err = service.CreateChildSystemTasks(createdSystemTask, metadata)
		if err != nil {
			return structs.ReceiptProcessingSystemTasks{}, err
		}
	}

	if metadata.FallbackReceiptProcessingSettingsIdRan > 0 {
		fallbackSystemTask := commands.UpsertSystemTaskCommand{
			Type:                 taskType,
			Status:               service.BoolToSystemTaskStatus(metadata.DidFallbackReceiptProcessingSettingsSucceed),
			AssociatedEntityType: models.RECEIPT_PROCESSING_SETTINGS,
			AssociatedEntityId:   metadata.FallbackReceiptProcessingSettingsIdRan,
			StartedAt:            startDate,
			EndedAt:              &endDate,
			ResultDescription:    metadata.FallbackRawResponse,
			RanByUserId:          userId,
		}

		if parentAssociatedSystemTaskId != nil {
			fallbackSystemTask.AssociatedSystemTaskId = parentAssociatedSystemTaskId(fallbackSystemTask)
		}

		createdFallbackSystemTask, err := systemTaskRepository.CreateSystemTask(fallbackSystemTask)
		if err != nil {
			return structs.ReceiptProcessingSystemTasks{}, err
		}

		_, err = service.CreateChildSystemTasks(createdFallbackSystemTask, metadata)
		if err != nil {
			return structs.ReceiptProcessingSystemTasks{}, err
		}

		result.FallbackSystemTask = createdFallbackSystemTask
	}

	return result, nil
}

func (service SystemTaskService) CreateChildSystemTasks(parentSystemTask models.SystemTask, metadata commands.ReceiptProcessingMetadata) ([]models.SystemTask, error) {
	var systemTasks []models.SystemTask
	systemTaskRepository := repositories.NewSystemTaskRepository(service.TX)
	isFallback := parentSystemTask.AssociatedEntityId == metadata.FallbackReceiptProcessingSettingsIdRan

	if !isFallback && len(metadata.OcrSystemTaskCommand.Type) > 0 {
		metadata.OcrSystemTaskCommand.AssociatedSystemTaskId = &parentSystemTask.ID
		ocrSystemTask, err := systemTaskRepository.CreateSystemTask(metadata.OcrSystemTaskCommand)
		if err != nil {
			return []models.SystemTask{}, err
		}

		systemTasks = append(systemTasks, ocrSystemTask)
	}

	// TODO: create tasks for prompt,
	if !isFallback && len(metadata.PromptSystemTaskCommand.Type) > 0 {
		metadata.PromptSystemTaskCommand.AssociatedSystemTaskId = &parentSystemTask.ID
		promptSystemTask, err := systemTaskRepository.CreateSystemTask(metadata.PromptSystemTaskCommand)
		if err != nil {
			return []models.SystemTask{}, err
		}

		systemTasks = append(systemTasks, promptSystemTask)
	}

	if !isFallback && len(metadata.ChatCompletionSystemTaskCommand.Type) > 0 {
		metadata.ChatCompletionSystemTaskCommand.AssociatedSystemTaskId = &parentSystemTask.ID
		chatCompletionSystemTask, err := systemTaskRepository.CreateSystemTask(metadata.ChatCompletionSystemTaskCommand)
		if err != nil {
			return []models.SystemTask{}, err
		}

		systemTasks = append(systemTasks, chatCompletionSystemTask)
	}

	if isFallback && len(metadata.FallbackOcrSystemTaskCommand.Type) > 0 {
		metadata.FallbackOcrSystemTaskCommand.AssociatedSystemTaskId = &parentSystemTask.ID
		fallbackOcrSystemTask, err := systemTaskRepository.CreateSystemTask(metadata.FallbackOcrSystemTaskCommand)
		if err != nil {
			return []models.SystemTask{}, err
		}

		systemTasks = append(systemTasks, fallbackOcrSystemTask)
	}

	if isFallback && len(metadata.FallbackPromptSystemTaskCommand.Type) > 0 {
		metadata.FallbackPromptSystemTaskCommand.AssociatedSystemTaskId = &parentSystemTask.ID
		fallbackPromptSystemTask, err := systemTaskRepository.CreateSystemTask(metadata.FallbackPromptSystemTaskCommand)
		if err != nil {
			return []models.SystemTask{}, err
		}

		systemTasks = append(systemTasks, fallbackPromptSystemTask)
	}

	if isFallback && len(metadata.FallbackChatCompletionSystemTaskCommand.Type) > 0 {
		metadata.FallbackChatCompletionSystemTaskCommand.AssociatedSystemTaskId = &parentSystemTask.ID
		fallbackChatCompletionSystemTask, err := systemTaskRepository.CreateSystemTask(metadata.FallbackChatCompletionSystemTaskCommand)
		if err != nil {
			return []models.SystemTask{}, err
		}

		systemTasks = append(systemTasks, fallbackChatCompletionSystemTask)
	}

	return systemTasks, nil
}

func (service SystemTaskService) BoolToSystemTaskStatus(value bool) models.SystemTaskStatus {
	if value {
		return models.SYSTEM_TASK_SUCCEEDED
	}
	return models.SYSTEM_TASK_FAILED
}

func (service SystemTaskService) CreateReceiptUploadedSystemTask(
	createReceiptError error,
	createdReceipt models.Receipt,
	quickScanSystemTasks structs.ReceiptProcessingSystemTasks,
	uploadStart time.Time,
) error {
	systemTaskRepository := repositories.NewSystemTaskRepository(service.GetDB())
	receiptProcessingSettingsId := quickScanSystemTasks.SystemTask.AssociatedEntityId
	systemTaskId := quickScanSystemTasks.SystemTask.ID
	status := models.SYSTEM_TASK_SUCCEEDED
	uploadFinished := time.Now()
	resultDescription := ""

	if quickScanSystemTasks.FallbackSystemTask.Status == models.SYSTEM_TASK_SUCCEEDED {
		receiptProcessingSettingsId = quickScanSystemTasks.FallbackSystemTask.AssociatedEntityId
		systemTaskId = quickScanSystemTasks.FallbackSystemTask.ID
	}

	receiptBytes, err := json.Marshal(createdReceipt)
	if err != nil {
		return err
	}

	resultDescription = string(receiptBytes)

	if createReceiptError != nil {
		status = models.SYSTEM_TASK_FAILED
		resultDescription = createReceiptError.Error()
	}

	_, err = systemTaskRepository.CreateSystemTask(commands.UpsertSystemTaskCommand{
		Type:                   models.RECEIPT_UPLOADED,
		Status:                 status,
		AssociatedEntityType:   models.RECEIPT_PROCESSING_SETTINGS,
		AssociatedEntityId:     receiptProcessingSettingsId,
		StartedAt:              uploadStart,
		EndedAt:                &uploadFinished,
		ResultDescription:      resultDescription,
		AssociatedSystemTaskId: &systemTaskId,
	})
	if err != nil {
		return err
	}

	return nil
}
