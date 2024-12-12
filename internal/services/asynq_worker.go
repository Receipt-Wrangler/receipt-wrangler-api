package services

import (
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/tasks"

	"github.com/hibiken/asynq"
)

func StartAsynqWorker() (*asynq.Server, error) {
	connectionString, err := config.BuildRedisConnectionString()
	if err != nil {
		return nil, err
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: connectionString},
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 10,
			// Optionally specify multiple queues with different priority.
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeEmailDelivery, tasks.HandleEmailDeliveryTask)

	err = srv.Run(mux)
	if err != nil {
		return nil, err
	}

	return srv, nil
}
