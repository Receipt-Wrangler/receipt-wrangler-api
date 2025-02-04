package wranglerasynq

import (
	"github.com/hibiken/asynq"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
)

var server *asynq.Server

func StartEmbeddedAsynqServer() error {
	opts, err := config.GetAsynqRedisClientConnectionOptions()
	if err != nil {
		return err
	}
	queuePriorityMap := map[string]int{}
	defaultQueueConfigurationMap := models.GetDefaultQueueConfigurationMap()

	systemSettingsRepository := repositories.NewSystemSettingsRepository(nil)
	systemSettings, err := systemSettingsRepository.GetSystemSettings()
	if err != nil {
		return err
	}

	for _, queueName := range models.GetQueueNames() {
		var queueConfigurationToUse models.TaskQueueConfiguration
		for _, queueConfiguration := range systemSettings.TaskQueueConfigurations {
			if queueConfiguration.Name == queueName {
				queueConfigurationToUse = queueConfiguration
				break
			}
		}

		if queueConfigurationToUse.ID == 0 {
			queueConfigurationToUse = defaultQueueConfigurationMap[queueName]
		}

		queuePriorityMap[string(queueName)] = queueConfigurationToUse.Priority
	}

	asynqConfig := asynq.Config{
		Concurrency: systemSettings.TaskConcurrency,
		Queues:      queuePriorityMap,
	}

	server = asynq.NewServer(
		opts,
		asynqConfig,
	)

	mux := BuildMux()

	go func() {
		logging.LogStd(logging.LOG_LEVEL_INFO, "Starting Asynq Server")
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
	mux.HandleFunc(RefreshTokenCleanUp, HandleRefreshTokenCleanupTask)

	return mux
}

func ShutDownEmbeddedAsynqServer() {
	server.Shutdown()
}

func RestartEmbeddedAsynqServer() error {
	ShutDownEmbeddedAsynqServer()
	return StartEmbeddedAsynqServer()
}
