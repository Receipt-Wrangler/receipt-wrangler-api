package connect

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

type Config struct {
	ConnectionString string
}

func Connect() {
	connectionString := getConnectionString()
	connectedDb, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{})

	if err != nil {
		panic(err.Error())
	}

	db = connectedDb
}

func getConnectionString() string {
	jsonFile, err := os.Open("config.json")

	if err != nil {
		panic(err.Error())
	}

	bytes, err := ioutil.ReadAll(jsonFile)

	var config Config
	marshalErr := json.Unmarshal(bytes, &config)

	if marshalErr != nil {
		panic(err.Error())
	}

	jsonFile.Close()
	return config.ConnectionString
}

func GetDB() *gorm.DB {
	return db
}
