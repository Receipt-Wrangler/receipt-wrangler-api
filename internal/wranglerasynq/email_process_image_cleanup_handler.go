package wranglerasynq

import (
	"context"
	"encoding/json"
	"github.com/hibiken/asynq"
	"os"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
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
	inspector, err := GetAsynqInspector()
	if err != nil {
		return nil, err
	}
	defer inspector.Close()

	queueName := string(models.EmailReceiptProcessingQueue)
	var allTasks []*asynq.TaskInfo

	// The functions have additional optional parameters, so we need to wrap them
	statusFuncs := []func(string, ...asynq.ListOption) ([]*asynq.TaskInfo, error){
		inspector.ListScheduledTasks,
		inspector.ListPendingTasks,
		inspector.ListActiveTasks,
		inspector.ListRetryTasks,
		inspector.ListArchivedTasks,
		inspector.ListCompletedTasks,
	}

	for _, getStatus := range statusFuncs {
		tasks, err := getStatus(queueName)
		if err != nil {
			return nil, err
		}
		allTasks = append(allTasks, tasks...)
	}

	return allTasks, nil
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
			if task.State != asynq.TaskStateCompleted && task.State != asynq.TaskStateArchived {
				allTasksCompleted = false
				break
			}
		}

		if allTasksCompleted {
			if utils.FileExists(imageForOcrPath.ImageForOcrPath) {
				err := os.Remove(imageForOcrPath.ImageForOcrPath)
				if err != nil {
					logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
				}
			}

			if utils.FileExists(imageForOcrPath.OriginalFilePath) {
				err := os.Remove(imageForOcrPath.OriginalFilePath)
				if err != nil {
					logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
				}
			}
		}
	}

	return nil
}
