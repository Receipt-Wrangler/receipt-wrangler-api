package wranglerasynq

import (
	"context"
	"github.com/hibiken/asynq"
)

func HandleEmailPollTask(context context.Context, task *asynq.Task) error {
	return nil
	// email.CallClient(true)
}
