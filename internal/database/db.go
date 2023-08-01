package db

import (
	"fmt"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"

	"gorm.io/gorm"
)

var db *gorm.DB

func BuildConnectionString() string {
	envVariables := config.GetEnvVariables()
	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", envVariables["MYSQL_USER"], envVariables["MYSQL_PASSWORD"], envVariables["MYSQL_HOST"], envVariables["MYSQL_DATABASE"])
	return connectionString
}

func Connect() error {
	connectedDb, err := gorm.Open(mysql.Open(BuildConnectionString()), &gorm.Config{})

	if err != nil {
		return err
	}

	db = connectedDb
	return nil
}

func MakeMigrations() {
	db.AutoMigrate(
		&models.RefreshToken{},
		&models.User{},
		&models.Receipt{},
		&models.Item{},
		&models.FileData{},
		&models.Tag{},
		&models.Category{},
		&models.Group{},
		&models.GroupMember{},
		&models.Comment{},
		&models.Notification{},
	)
}

func GetDB() *gorm.DB {
	return db
}

func InitTestDb() {
	sqlite, err := gorm.Open(sqlite.Open("file:test.db?_pragma=foreign_keys(1)"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db = sqlite
}
