package config

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"receipt-wrangler/api/internal/structs"
	"strings"
)

var config structs.Config
var featureConfig structs.FeatureConfig
var basePath string
var env string
var envVariables = make(map[string]string)

func GetConfig() structs.Config {
	return config
}

func GetFeatureConfig() structs.FeatureConfig {
	return featureConfig
}

func GetBasePath() string {
	return basePath
}

func GetEnvVariables() map[string]string {
	return envVariables
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

	err = setFeatureConfig()
	if err != nil {
		return err
	}

	return nil
}

func ReadEnvVariables() error {
	envKeys := []string{"DB_ROOT_PASSWORD", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_HOST", "DB_PORT", "DB_ENGINE", "DB_FILENAME"}
	for _, key := range envKeys {
		value := os.Getenv(key)
		envVariables[key] = value
	}
	return nil
}

func setFeatureConfig() error {
	path := filepath.Join(basePath, "config", "feature-config."+env+".json")
	jsonFile, err := os.Open(path)

	if err != nil {
		featureConfig = structs.FeatureConfig{
			EnableLocalSignUp: false,
		}
		return nil
	}

	bytes, err := ioutil.ReadAll(jsonFile)

	var configFile structs.FeatureConfig
	marshalErr := json.Unmarshal(bytes, &configFile)

	if marshalErr != nil {
		return err
	}

	jsonFile.Close()
	featureConfig = configFile
	return nil
}

func setSettingsConfig() error {
	path := filepath.Join(basePath, "config", "config."+env+".json")
	jsonFile, err := os.Open(path)

	if err != nil {
		return err
	}

	bytes, err := ioutil.ReadAll(jsonFile)

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
