package services

import (
	"github.com/hibiken/asynq"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
)

var scheduler *asynq.Scheduler

// TODO: finish implementing startup/shutdown of scheduler, then add email polling task id to system settings
func StartEmbeddedAsynqScheduler() error {
	opts, err := config.GetAsynqRedisClientConnectionOptions()
	if err != nil {
		return err
	}

	schedulerOpts := asynq.SchedulerOpts{}

	scheduler = asynq.NewScheduler(
		opts,
		&schedulerOpts,
	)

	go func() {
		err = scheduler.Start()
		if err != nil {
			logging.LogStd(logging.LOG_LEVEL_FATAL, err.Error())
		}
	}()

	return nil
}

func GetAsynqScheduler() *asynq.Scheduler {
	return scheduler
}

func ShutDownEmbeddedAsynqScheduler() {
	scheduler.Shutdown()
}
