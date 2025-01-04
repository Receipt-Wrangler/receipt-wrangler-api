package config

import (
	"fmt"
	"github.com/hibiken/asynq"
	"os"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
)

func GetAsynqRedisClientConnectionOptions() (asynq.RedisClientOpt, error) {
	connectionString, err := BuildRedisConnectionString()
	if err != nil {
		return asynq.RedisClientOpt{}, err
	}

	return asynq.RedisClientOpt{Addr: connectionString}, nil
}

func GetRedisConfig() (structs.RedisConfig, error) {
	port, err := simpleutils.StringToInt(os.Getenv("REDIS_PORT"))
	if err != nil {
		return structs.RedisConfig{}, fmt.Errorf("invalid REDIS_PORT environment variable: " + err.Error())
	}

	return structs.RedisConfig{
		Host: os.Getenv("REDIS_HOST"),
		Port: port,
	}, nil
}

func BuildRedisConnectionString() (string, error) {
	redisConfig, err := GetRedisConfig()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port), nil
}
