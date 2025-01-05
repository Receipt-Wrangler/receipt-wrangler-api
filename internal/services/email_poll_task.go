package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
)

type EmailPollPayload struct {
	Token            *structs.Claims
	PaidByUserId     uint
	GroupId          uint
	Status           models.ReceiptStatus
	TempPath         string
	OriginalFileName string
}

func HandleEmailPollTask(context context.Context, task *asynq.Task) error {
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
		payload.PaidByUserId,
		payload.GroupId,
		payload.Status,
		payload.TempPath,
		payload.OriginalFileName,
		taskId,
	)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		return err
	}

	return nil
}
