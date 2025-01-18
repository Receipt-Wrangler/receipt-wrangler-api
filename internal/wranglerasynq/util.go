package wranglerasynq

import (
	"fmt"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
)

func SetActivityCanBeRestarted(activities *[]structs.Activity) error {
	inspector, err := GetAsynqInspector()
	if err != nil {
		return err
	}
	archivedTasks, err := inspector.ListScheduledTasks(string(models.QuickScanQueue))
	if err != nil {
		return err
	}

	systemTaskRepository := repositories.NewSystemTaskRepository(nil)

	fmt.Println(archivedTasks)

	for _, activity := range *activities {
		if activity.Type == models.QUICK_SCAN && activity.AssociatedSystemTaskId != nil {
			associatedSystemTask, err := systemTaskRepository.GetSystemTaskById(*activity.AssociatedSystemTaskId)
			if err != nil {
				return err
			}
			for i := 0; i < len(archivedTasks); i++ {
				task := archivedTasks[i]
				if task.ID == associatedSystemTask.AsynqTaskId {
					activity.CanBeRestarted = true
					fmt.Println("hit")
					break
				}
			}
		}
	}

	return nil
}
