package wranglerasynq

import (
	"context"
	"encoding/json"
	"github.com/hibiken/asynq"
)

type EmailPollTaskPayload struct {
	PollAllGroups bool
	GroupIds      []string
}

func HandleEmailPollTask(context context.Context, task *asynq.Task) error {
	payload := task.Payload()

	if payload == nil {
		groupIds := make([]string, 0)
		return CallClient(true, groupIds)
	}

	var emailPollTaskPayload EmailPollTaskPayload

	err := json.Unmarshal(payload, &emailPollTaskPayload)
	if err != nil {
		return err
	}

	return CallClient(emailPollTaskPayload.PollAllGroups, emailPollTaskPayload.GroupIds)
}
