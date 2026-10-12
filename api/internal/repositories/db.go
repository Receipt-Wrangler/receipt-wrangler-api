package repositories

import (
	"fmt"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"time"

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

const (
	// dbConnectMaxAttempts and dbConnectRetryDelay control how long Connect waits
	// for the database to become reachable before giving up. This lets the API
	// survive the database not being ready yet at startup (e.g. when containers are
	// brought up by Docker restart policies or an orchestrator that does not
	// guarantee start order), instead of exiting fatally on the first connection
	// error and leaving the container in a broken state.
	dbConnectMaxAttempts = 10
	dbConnectRetryDelay  = 3 * time.Second
)

func Connect() error {
	dbConfig, err := config.GetDatabaseConfig()
	if err != nil {
		return err
	}

	dbEngine := dbConfig.Engine

	var dialector gorm.Dialector

	// Network-backed engines can be transiently unreachable at startup, so they
	// get connection retries. Sqlite is file-based — failures there are
	// configuration problems, not races, and should surface immediately.
	maxAttempts := dbConnectMaxAttempts

	switch dbEngine {
	case "mariadb", "mysql":
		dialector = mysql.Open(BuildMariaDbConnectionString(dbConfig))
	case "postgresql":
		dialector = postgres.Open(BuildPostgresqlConnectionString(dbConfig))
	case "sqlite":
		connectionString, sqliteErr := BuildSqliteConnectionString(dbConfig)
		if sqliteErr != nil {
			return sqliteErr
		}
		dialector = sqlite.Open(connectionString)
		maxAttempts = 1
	default:
		return fmt.Errorf("database engine of: %s! check your config to make sure it is correct", dbEngine)
	}

	connectedDb, err := openWithRetry(dialector, maxAttempts)
	if err != nil {
		return err
	}

	db = connectedDb
	return nil
}

// openWithRetry opens the database connection, retrying transient failures so the
// API does not exit fatally when the database is briefly unreachable at startup.
func openWithRetry(dialector gorm.Dialector, maxAttempts int) (*gorm.DB, error) {
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		connectedDb, err := gorm.Open(dialector, &gorm.Config{})
		if err == nil {
			return connectedDb, nil
		}

		lastErr = err
		if attempt < maxAttempts {
			logging.LogStd(
				logging.LOG_LEVEL_INFO,
				fmt.Sprintf(
					"database not ready (attempt %d/%d): %s; retrying in %s",
					attempt, maxAttempts, err.Error(), dbConnectRetryDelay,
				),
			)
			time.Sleep(dbConnectRetryDelay)
		}
	}

	return nil, fmt.Errorf("failed to connect to the database after %d attempts: %w", maxAttempts, lastErr)
}

func MakeMigrations() error {
	err := db.AutoMigrate(
		&models.RefreshToken{},
		&models.AppRole{},
		&models.AppRolePermission{},
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
		&models.GroupRoleDefinition{},
		&models.GroupRolePermission{},
		&models.GroupRoleCategoryGrant{},
		&models.GroupRoleTagGrant{},
		&models.GroupRolePaidByUserGrant{},
		&models.GroupMember{},
		&models.GroupMemberCategoryGrant{},
		&models.GroupMemberTagGrant{},
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
		&models.ReportTemplate{},
		&models.ReportTemplateGroup{},
		&models.GroupRoleReportTemplateGrant{},
		&models.GroupReceiptSettings{},
		&models.GroupReceiptSettingsCustomField{},
		&models.GroupReceiptSettingsSummaryCustomField{},
		&models.GroupReceiptSettingsSummaryStatus{},
		&models.Pepper{},
		&models.ApiKey{},
		&models.DataMigration{},
		&models.OAuthClient{},
		&models.OAuthAuthorizationCode{},
		&models.OidcProvider{},
		&models.OidcIdentity{},
		&models.OidcAuthSession{},
		&models.OidcExchangeCode{},
		&models.OidcLinkLaunch{},
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

	if err := SeedSystemRoles(); err != nil {
		return err
	}

	// Must run after the roles are seeded and before the bootstrap admin /
	// data migration so the default app and group roles exist for assignment.
	if err := EnsureDefaultRoles(); err != nil {
		return err
	}

	if config.GetDeployEnv() != "test" {
		userRepository := NewUserRepository(nil)
		err := userRepository.CreateUserIfNoneExist()
		if err != nil {
			return err
		}
	}

	// Runs after the bootstrap admin is created so a fresh install's first user
	// is assigned its legacy-equivalent role by the same one-time migration.
	if err := RunDataMigrations(); err != nil {
		return err
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
