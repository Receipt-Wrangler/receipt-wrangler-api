package wranglerasynq

import (
	"github.com/hibiken/asynq"
	"receipt-wrangler/api/internal/repositories"
)

func EnqueueTask(task *asynq.Task) (*asynq.TaskInfo, error) {
	client := repositories.GetAsynqClient()
	return client.Enqueue(task, asynq.MaxRetry(3))
}

func RegisterTask(cronspec string, task *asynq.Task) (string, error) {
	return scheduler.Register(cronspec, task, asynq.MaxRetry(3))
}

func UnregisterTask(taskId string) error {
	return scheduler.Unregister(taskId)
}
