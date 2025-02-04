package wranglerasynq

import (
	"errors"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
)

func SetActivityCanBeRestarted(activities *[]structs.Activity) error {
	inspector, err := GetAsynqInspector()
	if err != nil {
		return err
	}
	defer inspector.Close()

	rerunableArchivedTasks, err := inspector.ListArchivedTasks(string(models.QuickScanQueue))
	if err != nil {
		// We do not return this error because it will happen on a fresh redis instance, with no quick scans ever ran
		return nil
	}

	archivedEmailProcessingTasks, err := inspector.ListArchivedTasks(string(models.EmailReceiptProcessingQueue))
	if err != nil {
		// We do not return this error because it will happen on a fresh redis instance, with no email processing ever ran
		return nil
	}
	rerunableArchivedTasks = append(rerunableArchivedTasks, archivedEmailProcessingTasks...)

	systemTaskRepository := repositories.NewSystemTaskRepository(nil)

	for i := range *activities {
		activity := &(*activities)[i]

		if activity.Type == models.QUICK_SCAN || activity.Type == models.EMAIL_UPLOAD {
			systemTask, err := systemTaskRepository.GetSystemTaskById(activity.Id)
			if err != nil {
				return err
			}

			for i := 0; i < len(rerunableArchivedTasks); i++ {
				task := rerunableArchivedTasks[i]
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
