package repositories

import (
	"github.com/hibiken/asynq"
)

var client *asynq.Client

func GetAsynqRedisClient() *asynq.Client {
	return client
}

// TODO: get redis host port from env var
func ConnectToRedis() error {
	client = asynq.NewClient(asynq.RedisClientOpt{
		Addr: "172.17.0.3:6379",
	})
	err := client.Ping()
	if err != nil {
		return err
	}
	return nil
}
