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
	host := fmt.Sprintf("%s:%d", dbConfig.Host, dbConfig.Port)
	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbConfig.User, dbConfig.Password, host, dbConfig.Name)
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
	dbConfig, err := config.GetDatabaseConfig()
	if err != nil {
		return err
	}

	dbEngine := dbConfig.Engine

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

	if connectedDb == nil {
		return fmt.Errorf("database engine of: %s! check your config to make sure it is correct", dbEngine)
	}

	db = connectedDb
	return nil
}

func MakeMigrations() error {
	err := db.AutoMigrate(
		&models.RefreshToken{},
		&models.User{},
		&models.CustomField{},
		&models.CustomFieldValue{},
		&models.CustomFieldOption{},
		&models.Receipt{},
		&models.Item{},
		&models.FileData{},
		&models.Tag{},
		&models.Category{},
		&models.Group{},
		&models.GroupMember{},
		&models.Comment{},
		&models.Notification{},
		&models.UserShortcut{},
		&models.UserPrefernces{},
		&models.SubjectLineRegex{},
		&models.GroupSettingsWhiteListEmail{},
		&models.GroupSettings{},
		&models.Dashboard{},
		&models.Widget{},
		&models.TaskQueueConfiguration{},
		&models.SystemSettings{},
		&models.SystemEmail{},
		&models.SystemTask{},
		&models.ReceiptProcessingSettings{},
		&models.Prompt{},
		&models.GroupReceiptSettings{},
		&models.Pepper{},
		&models.ApiKey{},
	)

	return err
}

func GetDB() *gorm.DB {
	return db
}

func InitDB() error {
	var systemSettingsCount int64
	if err := db.Model(&models.SystemSettings{}).Count(&systemSettingsCount).Error; err != nil {
		return err
	}

	if systemSettingsCount == 0 {
		err := db.Create(&models.SystemSettings{})
		if err.Error != nil {
			return err.Error
		}
	}

	if config.GetDeployEnv() != "test" {
		userRepository := NewUserRepository(nil)
		err := userRepository.CreateUserIfNoneExist()
		if err != nil {
			return err
		}
	}

	return nil
}

func InitTestDb() {
	sqlite, err := gorm.Open(sqlite.Open("file:test.db?_pragma=foreign_keys(1)"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db = sqlite
}
