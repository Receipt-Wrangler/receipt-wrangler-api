package repositories

import (
	"fmt"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"

	"gorm.io/gorm"
)

var db *gorm.DB

func BuildMariaDbConnectionString() string {
	envVariables := config.GetEnvVariables()
	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", envVariables["DB_USER"], envVariables["DB_PASSWORD"], envVariables["DB_HOST"], envVariables["DB_NAME"])
	return connectionString
}

func BuildPostgresqlConnectionString() string {
	envVariables := config.GetEnvVariables()
	connectionString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", envVariables["DB_HOST"], envVariables["DB_USER"], envVariables["DB_PASSWORD"], envVariables["DB_NAME"], envVariables["DB_PORT"])
	return connectionString
}

func BuildSqliteConnectionString() (string, error) {
	envVariables := config.GetEnvVariables()
	err := utils.DirectoryExists("./sqlite", true)
	if err != nil {
		return "", err
	}

	connectionString := fmt.Sprintf("file:./sqlite/%s?_pragma=foreign_keys(1)", envVariables["DB_FILENAME"])
	return connectionString, nil
}

func Connect() error {
	envVariables := config.GetEnvVariables()
	dbEngine := envVariables["DB_ENGINE"]
	var err error
	var connectedDb *gorm.DB

	if dbEngine == "mariadb" || dbEngine == "mysql" {
		connectedDb, err = gorm.Open(mysql.Open(BuildMariaDbConnectionString()), &gorm.Config{})
	}

	if dbEngine == "postgresql" {
		connectedDb, err = gorm.Open(postgres.Open(BuildPostgresqlConnectionString()), &gorm.Config{})
	}

	if dbEngine == "sqlite" {
		connectionString, err := BuildSqliteConnectionString()
		if err != nil {
			return err
		}
		connectedDb, err = gorm.Open(sqlite.Open(connectionString), &gorm.Config{})
	}

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
		&models.UserPrefernces{},
		&models.SubjectLineRegex{},
		&models.GroupSettingsWhiteListEmail{},
		&models.GroupSettings{},
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
