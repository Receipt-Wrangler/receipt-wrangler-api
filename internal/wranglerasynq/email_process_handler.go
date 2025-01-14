package wranglerasynq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"time"
)

type EmailProcessTaskPayload struct {
	GroupSettingsId uint
	ImageForOcrPath string
	TempFilePath    string
	Metadata        structs.EmailMetadata
	Attachment      structs.Attachment
}

func HandleEmailProcessTask(context context.Context, task *asynq.Task) error {
	db := repositories.GetDB()
	systemTaskService := services.NewSystemTaskService(nil)
	groupSettingsRepository := repositories.NewGroupSettingsRepository(nil)
	var payload EmailProcessTaskPayload

	taskId, err := GetTaskIdFromContext(context)
	if err != nil {
		return HandleError(err)
	}

	err = json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return HandleError(err)
	}

	fileBytes, err := utils.ReadFile(payload.TempFilePath)
	if err != nil {
		return HandleError(err)
	}

	groupIdString := utils.UintToString(payload.GroupSettingsId)
	groupSettingsToUse, err := groupSettingsRepository.GetGroupSettingsById(groupIdString)
	if err != nil {
		return HandleError(err)
	}

	emailProcessStart := time.Now()

	if groupSettingsToUse.ID == 0 {
		return HandleError(fmt.Errorf("could not find group settings with id %d", payload.GroupSettingsId))
	}

	start := time.Now()
	baseCommand, processingMetadata, err := services.ReadReceiptImageFromFileOnly(payload.ImageForOcrPath, groupIdString)
	end := time.Now()

	if err != nil {
		return HandleError(err)
	}

	command := baseCommand
	command.GroupId = groupSettingsToUse.GroupId

	if len(command.Status) == 0 {
		command.Status = groupSettingsToUse.EmailDefaultReceiptStatus
	}

	if command.PaidByUserID == 0 {
		command.PaidByUserID = *groupSettingsToUse.EmailDefaultReceiptPaidById
	}

	command.CreatedByString = "Email Integration"

	err = db.Transaction(func(tx *gorm.DB) error {
		receiptRepository := repositories.NewReceiptRepository(tx)
		receiptImageRepository := repositories.NewReceiptImageRepository(tx)
		systemTaskRepository := repositories.NewSystemTaskRepository(tx)
		systemTaskService.SetTransaction(tx)
		emailProcessEnd := time.Now()

		metadataBytes, err := json.Marshal(payload.Metadata)
		if err != nil {
			return HandleError(err)
		}

		emailReadSystemTask, err := systemTaskRepository.CreateSystemTask(
			commands.UpsertSystemTaskCommand{
				Type:                 models.EMAIL_READ,
				Status:               models.SYSTEM_TASK_SUCCEEDED,
				AssociatedEntityType: models.SYSTEM_EMAIL,
				AssociatedEntityId:   groupSettingsToUse.SystemEmail.ID,
				StartedAt:            emailProcessStart,
				EndedAt:              &emailProcessEnd,
				RanByUserId:          nil,
				ResultDescription:    string(metadataBytes),
				AsynqTaskId:          taskId,
			},
		)
		if err != nil {
			return HandleError(err)
		}

		processingSystemTasks, err := systemTaskService.CreateSystemTasksFromMetadata(
			processingMetadata,
			start,
			end,
			models.EMAIL_UPLOAD,
			nil,
			func(command commands.UpsertSystemTaskCommand) *uint {
				return &emailReadSystemTask.ID
			},
		)

		createReceiptStart := time.Now()
		createdReceipt, err := receiptRepository.CreateReceipt(command, 0)
		_, taskErr := systemTaskService.CreateReceiptUploadedSystemTask(
			err,
			createdReceipt,
			processingSystemTasks,
			time.Now(),
		)
		if taskErr != nil {
			return HandleError(taskErr)
		}
		if err != nil {
			tx.Commit()
			return HandleError(taskErr)
		}
		createReceiptEnd := time.Now()

		fileData := models.FileData{
			ReceiptId: createdReceipt.ID,
			Name:      payload.Attachment.Filename,
			FileType:  payload.Attachment.FileType,
			Size:      payload.Attachment.Size,
		}

		_, err = receiptImageRepository.CreateReceiptImage(fileData, fileBytes)
		if err != nil {
			return HandleError(err)
		}

		err = systemTaskService.AssociateSystemTasksToReceipt(
			createdReceipt.ID,
			emailReadSystemTask.ID,
			createReceiptStart,
			createReceiptEnd,
		)
		if err != nil {
			tx.Commit()
			return HandleError(err)
		}

		return nil
	})

	return err
}
