package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
)

type EmailDeliveryPayload struct {
	UserID     int
	TemplateID string
}

func HandleTestTask(ctx context.Context, t *asynq.Task) error {
	var message string
	fmt.Println("In the task")
	err := json.Unmarshal(t.Payload(), &message)
	if err != nil {
		return err
	}

	fmt.Println(message)
	return nil
}
