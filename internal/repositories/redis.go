package repositories

import (
	"github.com/hibiken/asynq"
	config "receipt-wrangler/api/internal/env"
)

var client *asynq.Client

func GetAsynqClient() *asynq.Client {
	return client
}

func ConnectToRedis() error {
	opts, err := config.GetAsynqRedisClientConnectionOptions()
	if err != nil {
		return err
	}

	client = asynq.NewClient(opts)
	err = client.Ping()
	if err != nil {
		return err
	}

	return nil
}

func ShutdownAsynqClient() error {
	return client.Close()
}
