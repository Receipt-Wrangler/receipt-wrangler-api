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
var basePath string

func GetConfig() structs.Config {
	return config
}

func GetBasePath() string {
	return basePath
}

func SetConfig() error {
	setBasePath()
	env := getEnv()
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

func getEnv() string {
	envFlag := flag.String("env", "dev", "set runtime environment")
	flag.Parse()

	return *envFlag
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
