package db

import (
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func Connect() {
	config := config.GetConfig()
	connectedDb, err := gorm.Open(mysql.Open(config.ConnectionString), &gorm.Config{})

	if err != nil {
		panic(err.Error())
	}

	db = connectedDb
}

func MakeMigrations() {
	db.AutoMigrate(&models.User{}, &models.Receipt{}, &models.Item{}, &models.FileData{})
}

func GetDB() *gorm.DB {
	return db
}
