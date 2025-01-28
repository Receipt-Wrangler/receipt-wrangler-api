package wranglerasynq

import (
	"errors"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
)

func SetActivityCanBeRestarted(activities *[]structs.Activity) error {
	inspector, err := GetAsynqInspector()
	if err != nil {
		return err
	}

	archivedTasks, err := inspector.ListArchivedTasks(string(models.QuickScanQueue))
	if err != nil {
		// We do not return this error because it will happen on a fresh redis instance, with no quick scans ever ran
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		return nil
	}
	systemTaskRepository := repositories.NewSystemTaskRepository(nil)

	for i := range *activities {
		activity := &(*activities)[i]

		if activity.Type == models.QUICK_SCAN {
			systemTask, err := systemTaskRepository.GetSystemTaskById(activity.Id)
			if err != nil {
				return err
			}

			for i := 0; i < len(archivedTasks); i++ {
				task := archivedTasks[i]
				if task.ID == systemTask.AsynqTaskId {
					activity.CanBeRestarted = true
					break
				}
			}
		}
	}

	return nil
}

func SystemTaskToQueueName(taskType models.SystemTaskType) (string, error) {
	if string(taskType) == string(models.QUICK_SCAN) {
		return string(models.QuickScanQueue), nil
	}

	if string(taskType) == string(models.EMAIL_UPLOAD) {
		return string(models.EmailReceiptProcessingQueue), nil
	}

	return "", errors.New("unsupported task type")
}
