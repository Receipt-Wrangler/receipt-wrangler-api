package repositories

import (
	"fmt"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"

	"gorm.io/gorm"
)

var db *gorm.DB

func BuildMariaDbConnectionString(dbConfig structs.DatabaseConfig) string {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Name)
	return connectionString
}

func BuildPostgresqlConnectionString(dbConfig structs.DatabaseConfig) string {
	connectionString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", dbConfig.Host, dbConfig.User, dbConfig.Password, dbConfig.Name, fmt.Sprint(dbConfig.Port))
	return connectionString
}

func BuildSqliteConnectionString(dbConfig structs.DatabaseConfig) (string, error) {
	err := utils.DirectoryExists("./sqlite", true)
	if err != nil {
		return "", err
	}

	connectionString := fmt.Sprintf("file:./sqlite/%s?_pragma=foreign_keys(1)", dbConfig.Filename)
	return connectionString, nil
}

func Connect() error {
	dbConfig := config.GetConfig().Database
	dbEngine := dbConfig.Engine

	// TODO: remove
	fmt.Println(config.GetConfig())
	fmt.Println(dbConfig)

	var err error
	var connectedDb *gorm.DB

	if dbEngine == "mariadb" || dbEngine == "mysql" {
		connectedDb, err = gorm.Open(mysql.Open(BuildMariaDbConnectionString(dbConfig)), &gorm.Config{})
		if err != nil {
			return err
		}
	}

	if dbEngine == "postgresql" {
		connectedDb, err = gorm.Open(postgres.Open(BuildPostgresqlConnectionString(dbConfig)), &gorm.Config{})
		if err != nil {
			return err
		}
	}

	if dbEngine == "sqlite" {
		connectionString, err := BuildSqliteConnectionString(dbConfig)
		if err != nil {
			return err
		}
		connectedDb, err = gorm.Open(sqlite.Open(connectionString), &gorm.Config{})
		if err != nil {
			return err

		}
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
		&models.Dashboard{},
		&models.Widget{},
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
