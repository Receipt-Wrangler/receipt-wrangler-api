package config

import (
	"flag"
	"os"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
	"strings"
)

var config structs.Config
var basePath string
var env string

func GetConfig() structs.Config {
	return config
}

func GetSecretKey() string {
	if len(os.Getenv("SECRET_KEY")) == 0 {
		logging.LogStd(logging.LOG_LEVEL_FATAL, constants.EMPTY_SECRET_KEY_ERROR)
	}

	return os.Getenv("SECRET_KEY")
}

func GetDatabaseConfig() (structs.DatabaseConfig, error) {
	port := os.Getenv("DB_PORT")
	portInt, err := simpleutils.StringToInt(port)
	if err != nil {

	}

	return structs.DatabaseConfig{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		Host:     os.Getenv("DB_HOST"),
		Port:     portInt,
		Engine:   os.Getenv("DB_ENGINE"),
		Filename: os.Getenv("DB_FILENAME"),
	}, nil
}

func GetBasePath() string {
	envBase := os.Getenv("BASE_PATH")
	if len(envBase) == 0 {
		return basePath
	}

	return envBase
}

func GetEncryptionKey() string {
	if len(os.Getenv("ENCRYPTION_KEY")) == 0 {
		logging.LogStd(logging.LOG_LEVEL_FATAL, constants.EMPTY_ENCRYPTION_KEY_ERROR)
	}

	return os.Getenv("ENCRYPTION_KEY")
}

func CheckRequiredEnvironmentVariables() {
	GetEncryptionKey()
	GetSecretKey()
}

func GetDeployEnv() string {
	return env
}

func SetConfigs() error {
	setEnv()
	setBasePath()

	return nil
}

func setEnv() {
	envFlag := flag.String("env", "dev", "set runtime environment")
	flag.Parse()

	env = *envFlag
	os.Setenv("ENV", env)
}

func setBasePath() {
	cwd, _ := os.Getwd()
	result := ""
	paths := strings.Split(cwd, "/")

	for i := 0; i < len(paths); i++ {
		result += "/" + paths[i]
		if paths[i] == "receipt-wrangler-api" {
			basePath = result
			return
		}
	}
}
