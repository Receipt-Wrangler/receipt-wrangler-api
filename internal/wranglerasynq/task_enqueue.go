package wranglerasynq

import (
	"github.com/hibiken/asynq"
	"receipt-wrangler/api/internal/repositories"
)

type QueueName string

const (
	QuickScanQueue                QueueName = "quick_scan"
	EmailPollingQueue             QueueName = "email_polling"
	EmailReceiptProcessingQueue   QueueName = "email_receipt_processing"
	EmailReceiptImageCleanupQueue QueueName = "email_receipt_image_cleanup"
)

func EnqueueTask(task *asynq.Task, queue QueueName) (*asynq.TaskInfo, error) {
	client := repositories.GetAsynqClient()
	return client.Enqueue(task, asynq.MaxRetry(3), asynq.Queue(string(queue)))
}

func RegisterTask(cronspec string, task *asynq.Task, queue QueueName, maxRetry int) (string, error) {
	return scheduler.Register(cronspec, task, asynq.MaxRetry(maxRetry), asynq.Queue(string(queue)))
}
