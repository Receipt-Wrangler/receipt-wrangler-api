package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/structs"
	"strings"
)

var config structs.Config
var basePath string
var env string

func GetConfig() structs.Config {
	return config
}

func GetFeatureConfig() structs.FeatureConfig {
	return config.Features
}

func GetBasePath() string {
	envBase := os.Getenv("BASE_PATH")
	if len(envBase) == 0 {
		return basePath
	}

	return envBase
}

func GetDeployEnv() string {
	return env
}

func SetConfigs() error {
	setEnv()
	setBasePath()

	err := setSettingsConfig()
	if err != nil {
		return err
	}

	return nil
}

func setSettingsConfig() error {
	path := filepath.Join(basePath, "config", "config."+env+".json")
	jsonFile, err := os.Open(path)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) && env != "test" {
			configStub := structs.Config{}
			bytes, err := json.MarshalIndent(configStub, "", "  ")
			if err != nil {
				return err
			}

			err = os.WriteFile(path, bytes, 0644)
			if err != nil {
				return err
			}
			logging.GetLogger().Fatalf(fmt.Sprintf("Config file not found at %s. A stub file has been created. Please fill in the necessary fields and restart the container.", path))
		}
		return err
	}
	defer func() {
		closeErr := jsonFile.Close()
		if closeErr != nil {
			err = closeErr
		}
	}()

	bytes, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	marshalErr := json.Unmarshal(bytes, &config)
	if marshalErr != nil {
		return err
	}

	if config.AiSettings.NumWorkers == 0 {
		config.AiSettings.NumWorkers = 1
	}

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
