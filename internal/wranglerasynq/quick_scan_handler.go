package wranglerasynq

import (
	"context"
	"encoding/json"
	"github.com/hibiken/asynq"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
)

type QuickScanTaskPayload struct {
	Token            *structs.Claims
	PaidByUserId     uint
	GroupId          uint
	Status           models.ReceiptStatus
	TempPath         string
	OriginalFileName string
}

func HandleQuickScanTask(context context.Context, task *asynq.Task) error {
	taskId, err := GetTaskIdFromContext(context)
	if err != nil {
		return HandleError(err)
	}

	var payload QuickScanTaskPayload

	err = json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return HandleError(err)
	}

	receiptService := services.NewReceiptService(nil)
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
		return HandleError(err)
	}

	return nil
}
