package config

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
	defer tearDownConfigTests()

	_, err := GetDatabaseConfig()
	if err == nil {
		utils.PrintTestError(t, err, "Parse error")
		return
	}
}

func TestShouldGetFullConfig(t *testing.T) {
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

	if dbConfig.Port != 1234 {
		utils.PrintTestError(t, "Expected 1234", dbConfig.Port)
	}

	if dbConfig.Engine != "sqlite" {
		utils.PrintTestError(t, "Expected sqlite", dbConfig.Engine)
	}

	if dbConfig.Filename != "test_filename" {
		utils.PrintTestError(t, "Expected test_filename", dbConfig.Filename)
	}
}

func TestShouldReturnErrorDueToBadPort(t *testing.T) {
	defer tearDownConfigTests()
	os.Setenv("DB_PORT", "not a number")

	_, err := GetDatabaseConfig()
	if err == nil {
		utils.PrintTestError(t, err, "Parse error")
		return
	}

}
