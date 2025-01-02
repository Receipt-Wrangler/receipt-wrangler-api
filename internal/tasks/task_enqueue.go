package tasks

import (
	"github.com/hibiken/asynq"
	"receipt-wrangler/api/internal/repositories"
)

func EnqueueTask(task *asynq.Task) (*asynq.TaskInfo, error) {
	client := repositories.GetAsynqClient()
	return client.Enqueue(task, asynq.MaxRetry(3))
}
