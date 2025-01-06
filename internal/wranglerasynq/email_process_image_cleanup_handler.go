package wranglerasynq

import (
	"context"
	"encoding/json"
	"github.com/hibiken/asynq"
	"os"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
)

type attachmentMapKey struct {
	OriginalFilePath string
	ImageForOcrPath  string
}

func HandleEmailProcessImageCleanUpTask(context context.Context, task *asynq.Task) error {
	tasks, err := getTaskInfo()
	if err != nil {
		return err
	}

	attachmentMap, err := buildAttachmentMap(tasks)
	if err != nil {
		return err
	}

	err = cleanupImages(attachmentMap)
	if err != nil {
		return err
	}

	return nil
}

func getTaskInfo() ([]*asynq.TaskInfo, error) {
	opts, err := config.GetAsynqRedisClientConnectionOptions()
	if err != nil {
		return nil, err
	}
	inspector := asynq.NewInspector(opts)
	return inspector.ListScheduledTasks(string(EmailReceiptProcessingQueue))
}

func buildAttachmentMap(taskInfos []*asynq.TaskInfo) (map[attachmentMapKey][]*asynq.TaskInfo, error) {
	attachmentMap := make(map[attachmentMapKey][]*asynq.TaskInfo)
	for _, taskInfo := range taskInfos {
		var payload EmailProcessTaskPayload
		payloadBytes := taskInfo.Payload
		err := json.Unmarshal(payloadBytes, &payload)
		if err != nil {
			return nil, err
		}

		key := attachmentMapKey{
			OriginalFilePath: payload.TempFilePath,
			ImageForOcrPath:  payload.ImageForOcrPath,
		}
		attachmentMap[key] = append(attachmentMap[key], taskInfo)
	}

	return attachmentMap, nil
}

func cleanupImages(attachmentMap map[attachmentMapKey][]*asynq.TaskInfo) error {
	for imageForOcrPath, tasksForImage := range attachmentMap {
		allTasksCompleted := true
		for _, task := range tasksForImage {
			if task.State != asynq.TaskStateCompleted {
				allTasksCompleted = false
				break
			}
		}

		if allTasksCompleted {
			err := os.Remove(imageForOcrPath.ImageForOcrPath)
			if err != nil {
				logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
			}

			err = os.Remove(imageForOcrPath.OriginalFilePath)
			if err != nil {
				logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
			}
		}
	}

	return nil
}
