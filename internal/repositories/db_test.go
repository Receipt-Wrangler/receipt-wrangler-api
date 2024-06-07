package repositories

import (
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"testing"
)

func teardownDbTests() {
}

func TestShouldCreateTheCorrectMariaDbString(t *testing.T) {
	defer teardownDbTests()

	dbConfig := structs.DatabaseConfig{
		User:     "root",
		Password: "password",
		Host:     "db",
		Port:     3306,
		Name:     "receipt_wrangler",
	}

	connectionString := BuildMariaDbConnectionString(dbConfig)
	expected := "root:password@tcp(db:3306)/receipt_wrangler?charset=utf8mb4&parseTime=True&loc=Local"

	if connectionString != expected {
		utils.PrintTestError(t, connectionString, expected)
	}
}

func TestShouldCreateTheCorrectPostgresqlString(t *testing.T) {
	defer teardownDbTests()

	dbConfig := structs.DatabaseConfig{
		User:     "root",
		Password: "password",
		Host:     "db",
		Port:     3306,
		Name:     "receipt_wrangler",
	}

	connectionString := BuildPostgresqlConnectionString(dbConfig)
	expected := "host=db user=root password=password dbname=receipt_wrangler port=3306 sslmode=disable"

	if connectionString != expected {
		utils.PrintTestError(t, connectionString, expected)
	}
}

func TestShouldCreateTheCorrectSqliteString(t *testing.T) {
	defer teardownDbTests()

	dbConfig := structs.DatabaseConfig{
		Filename: "test.db",
	}

	connectionString, _ := BuildSqliteConnectionString(dbConfig)
	expected := "file:./sqlite/test.db?_pragma=foreign_keys(1)"

	if connectionString != expected {
		utils.PrintTestError(t, connectionString, expected)
	}
}
