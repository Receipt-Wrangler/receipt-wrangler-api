package wranglerasynq

import (
	"github.com/hibiken/asynq"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
)

var server *asynq.Server

func StartEmbeddedAsynqServer() error {
	opts, err := config.GetAsynqRedisClientConnectionOptions()
	if err != nil {
		return err
	}

	server = asynq.NewServer(
		opts,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				string(models.QuickScanQueue):                4,
				string(models.EmailReceiptProcessingQueue):   3,
				string(models.EmailPollingQueue):             2,
				string(models.EmailReceiptImageCleanupQueue): 1,
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
