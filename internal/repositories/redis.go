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
	connectionString, err := config.BuildRedisConnectionString()
	if err != nil {
		return err
	}

	client = asynq.NewClient(asynq.RedisClientOpt{
		Addr: connectionString,
	})
	err = client.Ping()
	if err != nil {
		return err
	}

	return nil
}

func ShutdownAsynqClient() error {
	return client.Close()
}
