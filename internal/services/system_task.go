package services

import (
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

func (service SystemTaskService) CreateSystemTasksFromMetadata(metadata commands.ReceiptProcessingMetadata, startDate time.Time, endDate time.Time, taskType models.SystemTaskType, userId uint) (structs.ReceiptProcessingSystemTasks, error) {
	systemTaskRepository := repositories.NewSystemTaskRepository(nil)
	var chatCompletionSystemTask models.SystemTask
	result := structs.ReceiptProcessingSystemTasks{}

	if len(metadata.ChatCompletionSystemTaskCommand.Type) > 0 {
		createdSystemTask, err := systemTaskRepository.CreateSystemTask(metadata.ChatCompletionSystemTaskCommand)
		if err != nil {
			return structs.ReceiptProcessingSystemTasks{}, err
		}

		chatCompletionSystemTask = createdSystemTask
	}

	if metadata.ReceiptProcessingSettingsIdRan > 0 {
		systemTask := commands.UpsertSystemTaskCommand{
			Type:                 taskType,
			Status:               service.BoolToSystemTaskStatus(metadata.DidReceiptProcessingSettingsSucceed),
			AssociatedEntityType: models.RECEIPT_PROCESSING_SETTINGS,
			AssociatedEntityId:   metadata.ReceiptProcessingSettingsIdRan,
			StartedAt:            startDate,
			EndedAt:              &endDate,
			ResultDescription:    metadata.RawResponse,
			RanByUserId:          &userId,
		}

		if systemTask.Status == models.SYSTEM_TASK_SUCCEEDED && chatCompletionSystemTask.ID > 0 {
			systemTask.AssociatedSystemTaskId = &chatCompletionSystemTask.ID
		}

		createdSystemTask, err := systemTaskRepository.CreateSystemTask(systemTask)
		if err != nil {
			return structs.ReceiptProcessingSystemTasks{}, err
		}
		result.SystemTask = createdSystemTask
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
			RanByUserId:          &userId,
		}

		if fallbackSystemTask.Status == models.SYSTEM_TASK_SUCCEEDED && chatCompletionSystemTask.ID > 0 {
			fallbackSystemTask.AssociatedSystemTaskId = &chatCompletionSystemTask.ID
		}

		createdFallbackSystemTask, err := systemTaskRepository.CreateSystemTask(fallbackSystemTask)
		if err != nil {
			return structs.ReceiptProcessingSystemTasks{}, err
		}

		result.FallbackSystemTask = createdFallbackSystemTask
	}

	return result, nil
}

func (service SystemTaskService) BoolToSystemTaskStatus(value bool) models.SystemTaskStatus {
	if value {
		return models.SYSTEM_TASK_SUCCEEDED
	}
	return models.SYSTEM_TASK_FAILED
}
