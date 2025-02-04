package wranglerasynq

import (
	"github.com/hibiken/asynq"
	"receipt-wrangler/api/internal/models"
)

func StartSystemCleanUpTasks() error {
	inspector, err := GetAsynqInspector()
	if err != nil {
		return err
	}
	defer inspector.Close()

	cleanUpQueue := models.SystemCleanUpQueue

	inspector.DeleteAllScheduledTasks(string(cleanUpQueue))
	refreshTokenTask := asynq.NewTask(RefreshTokenCleanUp, nil)
	_, err = RegisterTask("@every 24h", refreshTokenTask, cleanUpQueue, 0)

	return err
}
