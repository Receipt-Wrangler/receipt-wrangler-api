package services

import (
	"github.com/hibiken/asynq"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/tasks"
)

var server *asynq.Server

func StartEmbeddedAsynqServer() error {
	connectionString, err := config.BuildRedisConnectionString()
	if err != nil {
		return err
	}

	server = asynq.NewServer(
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

	go func() {
		err = server.Run(mux)
		if err != nil {
			logging.LogStd(logging.LOG_LEVEL_FATAL, err.Error())
		}
	}()

	return nil
}

func ShutDownEmbeddedAsynqServer() {
	server.Shutdown()
}
