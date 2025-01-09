package models

import "receipt-wrangler/api/internal/wranglerasynq"

type AsynqQueueConfiguration struct {
	QueueName wranglerasynq.QueueName
	Priority  int
}
