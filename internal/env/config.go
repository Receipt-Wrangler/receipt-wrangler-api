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

func SetConfig() {
	setBasePath()
	env := getEnv()
	jsonFile, err := os.Open(filepath.Join(basePath, "config."+env+".json"))

	if err != nil {
		panic(err.Error())
	}

	bytes, err := ioutil.ReadAll(jsonFile)

	var configFile structs.Config
	marshalErr := json.Unmarshal(bytes, &configFile)

	if marshalErr != nil {
		panic(err.Error())
	}

	jsonFile.Close()
	config = configFile
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
