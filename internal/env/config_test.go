package env

import (
	"os"
	"receipt-wrangler/api/internal/utils"
	"testing"
)

func tearDownConfigTests() {
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_ENGINE")
	os.Unsetenv("DB_FILENAME")
}

func TestShouldGetEmptyDatabaseConfig(t *testing.T) {
	tearDownConfigTests()

	dbConfig, err := GetDatabaseConfig()
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if dbConfig.User != "" {
		utils.PrintTestError(t, "Expected empty string", dbConfig.User)
	}

	if dbConfig.Password != "" {
		utils.PrintTestError(t, "Expected empty string", dbConfig)
	}

	if dbConfig.Name != "" {
		utils.PrintTestError(t, "Expected empty string", dbConfig)
	}

	if dbConfig.Host != "" {
		utils.PrintTestError(t, "Expected empty string", dbConfig)
	}

	if dbConfig.Port != 0 {
		utils.PrintTestError(t, "Expected 0", dbConfig)
	}

	if dbConfig.Engine != "" {
		utils.PrintTestError(t, "Expected empty string", dbConfig)
	}

	if dbConfig.Filename != "" {
		utils.PrintTestError(t, "Expected empty string", dbConfig)
	}
}

func TestShouldGetFullConfigForSqlite(t *testing.T) {
	defer tearDownConfigTests()

	os.Setenv("DB_USER", "test_user")
	os.Setenv("DB_PASSWORD", "test_password")
	os.Setenv("DB_NAME", "test_name")
	os.Setenv("DB_HOST", "test_host")
	os.Setenv("DB_PORT", "1234")
	os.Setenv("DB_ENGINE", "sqlite")
	os.Setenv("DB_FILENAME", "test_filename")

	dbConfig, err := GetDatabaseConfig()
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if dbConfig.User != "test_user" {
		utils.PrintTestError(t, "Expected test_user", dbConfig.User)
	}

	if dbConfig.Password != "test_password" {
		utils.PrintTestError(t, "Expected test_password", dbConfig.Password)
	}

	if dbConfig.Name != "test_name" {
		utils.PrintTestError(t, "Expected test_name", dbConfig.Name)
	}

	if dbConfig.Host != "test_host" {
		utils.PrintTestError(t, "Expected test_host", dbConfig.Host)
	}

	if dbConfig.Port != 0 {
		utils.PrintTestError(t, "Expected 0", dbConfig.Port)
	}

	if dbConfig.Engine != "sqlite" {
		utils.PrintTestError(t, "Expected sqlite", dbConfig.Engine)
	}

	if dbConfig.Filename != "test_filename" {
		utils.PrintTestError(t, "Expected test_filename", dbConfig.Filename)
	}
}

func TestShouldGetFullConfigForMariaDb(t *testing.T) {
	defer tearDownConfigTests()

	os.Setenv("DB_USER", "test_user")
	os.Setenv("DB_PASSWORD", "test_password")
	os.Setenv("DB_NAME", "test_name")
	os.Setenv("DB_HOST", "test_host")
	os.Setenv("DB_PORT", "1234")
	os.Setenv("DB_ENGINE", "mariadb")
	os.Setenv("DB_FILENAME", "test_filename")

	dbConfig, err := GetDatabaseConfig()
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if dbConfig.User != "test_user" {
		utils.PrintTestError(t, "Expected test_user", dbConfig.User)
	}

	if dbConfig.Password != "test_password" {
		utils.PrintTestError(t, "Expected test_password", dbConfig.Password)
	}

	if dbConfig.Name != "test_name" {
		utils.PrintTestError(t, "Expected test_name", dbConfig.Name)
	}

	if dbConfig.Host != "test_host" {
		utils.PrintTestError(t, "Expected test_host", dbConfig.Host)
	}

	if dbConfig.Port != 1234 {
		utils.PrintTestError(t, "Expected 1234", dbConfig.Port)
	}

	if dbConfig.Engine != "mariadb" {
		utils.PrintTestError(t, "Expected mariadb", dbConfig.Engine)
	}

	if dbConfig.Filename != "test_filename" {
		utils.PrintTestError(t, "Expected test_filename", dbConfig.Filename)
	}
}

func TestShouldReturnErrorDueToBadPort(t *testing.T) {
	defer tearDownConfigTests()
	os.Setenv("DB_PORT", "not a number")
	os.Setenv("DB_ENGINE", "postgresql")

	_, err := GetDatabaseConfig()
	if err == nil {
		utils.PrintTestError(t, err, "Parse error")
		return
	}
}

func TestShouldParsePortCorrectlyForPostGresql(t *testing.T) {
	defer tearDownConfigTests()
	os.Setenv("DB_PORT", "1234")
	os.Setenv("DB_ENGINE", "postgresql")

	dbConfig, err := GetDatabaseConfig()
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if dbConfig.Port != 1234 {
		utils.PrintTestError(t, "Expected 1234", dbConfig.Port)
	}
}

func TestShouldParsePortCorrectlyForMariaDb(t *testing.T) {
	defer tearDownConfigTests()
	os.Setenv("DB_PORT", "1234")
	os.Setenv("DB_ENGINE", "mariadb")

	dbConfig, err := GetDatabaseConfig()
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if dbConfig.Port != 1234 {
		utils.PrintTestError(t, "Expected 1234", dbConfig.Port)
	}
}

func TestShouldParsePortCorrectlyForMysql(t *testing.T) {
	defer tearDownConfigTests()
	os.Setenv("DB_PORT", "1234")
	os.Setenv("DB_ENGINE", "mysql")

	dbConfig, err := GetDatabaseConfig()
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if dbConfig.Port != 1234 {
		utils.PrintTestError(t, "Expected 1234", dbConfig.Port)
	}
}
