package wranglerasynq

import (
	"github.com/hibiken/asynq"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
)

func EnqueueTask(task *asynq.Task, queue models.QueueName) (*asynq.TaskInfo, error) {
	client := repositories.GetAsynqClient()
	return client.Enqueue(task, asynq.MaxRetry(3), asynq.Queue(string(queue)))
}

func RegisterTask(cronspec string, task *asynq.Task, queue models.QueueName, maxRetry int) (string, error) {
	return scheduler.Register(cronspec, task, asynq.MaxRetry(maxRetry), asynq.Queue(string(queue)))
}
