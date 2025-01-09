package wranglerasynq

import (
	"github.com/hibiken/asynq"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
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
				string(QuickScanQueue):                4,
				string(EmailReceiptProcessingQueue):   3,
				string(EmailPollingQueue):             2,
				string(EmailReceiptImageCleanupQueue): 1,
			},
		},
	)

	mux := BuildMux()

	go func() {
		err = server.Run(mux)
		if err != nil {
			logging.LogStd(logging.LOG_LEVEL_FATAL, err.Error())
		}
	}()

	return nil
}

func BuildMux() *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.HandleFunc(QuickScan, HandleQuickScanTask)
	mux.HandleFunc(EmailPoll, HandleEmailPollTask)
	mux.HandleFunc(EmailProcess, HandleEmailProcessTask)
	mux.HandleFunc(EmailProcessImageCleanUp, HandleEmailProcessImageCleanUpTask)

	return mux
}

func ShutDownEmbeddedAsynqServer() {
	server.Shutdown()
}
