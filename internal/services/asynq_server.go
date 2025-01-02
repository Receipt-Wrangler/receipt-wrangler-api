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
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeTest, tasks.HandleTestTask)

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