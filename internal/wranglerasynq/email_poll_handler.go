package wranglerasynq

import (
	"context"
	"github.com/hibiken/asynq"
)

func HandleEmailPollTask(context context.Context, task *asynq.Task) error {
	groupIds := make([]string, 0)
	return CallClient(true, groupIds)
}
