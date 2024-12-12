package repositories

import (
	"fmt"
	"github.com/hibiken/asynq"
	config "receipt-wrangler/api/internal/env"
)

var client *asynq.Client

func GetAsynqRedisClient() *asynq.Client {
	return client
}

func ConnectToRedis() error {
	redisConfig, err := config.GetRedisConfig()
	if err != nil {
		return err
	}

	connectionString := fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port)

	client = asynq.NewClient(asynq.RedisClientOpt{
		Addr: connectionString,
	})
	err = client.Ping()
	if err != nil {
		return err
	}

	return nil
}
