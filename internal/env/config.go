package config

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
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
		if errors.Is(err, os.ErrNotExist) {
			configStub := structs.Config{}
			bytes, err := json.Marshal(configStub)
			if err != nil {
				return err
			}

			os.WriteFile(path, bytes, 0644)
		}
		return err
	}

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	var configFile structs.Config
	marshalErr := json.Unmarshal(bytes, &configFile)

	if marshalErr != nil {
		return err
	}

	jsonFile.Close()
	config = configFile
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
