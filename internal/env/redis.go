package config

import (
	"fmt"
	"github.com/hibiken/asynq"
	"os"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
)

func GetAsynqRedisClientConnectionOptions() (asynq.RedisClientOpt, error) {
	connectionString, err := buildRedisAddressString()
	if err != nil {
		return asynq.RedisClientOpt{}, err
	}

	redisConfig, err := GetRedisConfig()
	if err != nil {
		return asynq.RedisClientOpt{}, err
	}

	return asynq.RedisClientOpt{
			Addr:     connectionString,
			Username: redisConfig.Username,
			Password: redisConfig.Password,
		},
		nil
}

func GetRedisConfig() (structs.RedisConfig, error) {
	port, err := simpleutils.StringToInt(os.Getenv(string(constants.RedisPort)))
	if err != nil {
		return structs.RedisConfig{}, fmt.Errorf("invalid REDIS_PORT environment variable: " + err.Error())
	}

	username := os.Getenv(string(constants.RedisUser))
	password := os.Getenv(string(constants.RedisPassword))

	return structs.RedisConfig{
		Host:     os.Getenv(string(constants.RedisHost)),
		Port:     port,
		Username: username,
		Password: password,
	}, nil
}

func buildRedisAddressString() (string, error) {
	redisConfig, err := GetRedisConfig()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port), nil
}
