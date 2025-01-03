package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"mime/multipart"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
)

type QuickScanTaskPayload struct {
	Token        *structs.Claims
	File         multipart.File
	FileHeader   *multipart.FileHeader
	PaidByUserId uint
	GroupId      uint
	Status       models.ReceiptStatus
}

func HandleQuickScanTask(context context.Context, task *asynq.Task) error {
	taskId, ok := asynq.GetTaskID(context)
	if ok == false {
		errMessage := "taskId not found"
		logging.LogStd(logging.LOG_LEVEL_ERROR, errMessage)
		return fmt.Errorf(errMessage)
	}

	var payload QuickScanTaskPayload

	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		return err
	}

	receiptService := NewReceiptService(nil)
	_, err = receiptService.QuickScan(
		payload.Token,
		payload.File,
		payload.FileHeader,
		payload.PaidByUserId,
		payload.GroupId,
		payload.Status,
		taskId,
	)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		return err
	}

	return nil
}
