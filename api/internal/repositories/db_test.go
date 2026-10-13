package repositories

import (
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"testing"
)

func teardownDbTests() {
}

// foreignKeysPragma reads SQLite's foreign_keys pragma through the package
// database as it currently stands.
func foreignKeysPragma(t *testing.T) int {
	t.Helper()

	var enabled int
	if err := GetDB().Raw("PRAGMA foreign_keys").Scan(&enabled).Error; err != nil {
		t.Fatalf("failed to read the foreign_keys pragma: %v", err)
	}

	return enabled
}

// The hook must really switch enforcement off, or a cleanup test built on it
// passes on the cascade it exists to rule out, and its restore must put the
// enforcing pool back for every test after it.
func TestUseTestDbWithoutForeignKeysForTests(t *testing.T) {
	if foreignKeysPragma(t) != 1 {
		t.Fatalf("expected the test database to enforce foreign keys")
	}

	restore, err := UseTestDbWithoutForeignKeysForTests()
	if err != nil {
		t.Fatalf("UseTestDbWithoutForeignKeysForTests failed: %v", err)
	}
	// Deferred as well as called below: a t.Fatalf in between, including one
	// inside foreignKeysPragma, must not leave later tests on this handle.
	defer restore()

	if foreignKeysPragma(t) != 0 {
		t.Fatalf("expected foreign keys to be off while the hook is in effect")
	}

	restore()

	if foreignKeysPragma(t) != 1 {
		t.Fatalf("expected foreign keys to be enforced again after restore")
	}
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
