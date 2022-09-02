package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"receipt-wrangler/api/internal/structs"
)

var config structs.Config

func GetConfig() structs.Config {
	return config
}

func SetConfig() {
	jsonFile, err := os.Open("config.json")

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
