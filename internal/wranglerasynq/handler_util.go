package wranglerasynq

import (
	"context"
	"fmt"
	"github.com/hibiken/asynq"
	"receipt-wrangler/api/internal/logging"
)

func GetTaskIdFromContext(ctx context.Context) (string, error) {
	taskId, ok := asynq.GetTaskID(ctx)
	if ok == false {
		errMessage := "taskId not found"
		return "", fmt.Errorf(errMessage)
	}

	return taskId, nil
}

func HandleError(err error) error {
	logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
	return err
}
