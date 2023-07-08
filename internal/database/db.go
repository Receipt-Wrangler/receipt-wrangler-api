package db

import (
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"

	"gorm.io/gorm"
)

var db *gorm.DB

func Connect() error {
	config := config.GetConfig()
	connectedDb, err := gorm.Open(mysql.Open(config.ConnectionString), &gorm.Config{})

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
