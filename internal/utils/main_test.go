package utils

import (
	"fmt"
	"os"
	db "receipt-wrangler/api/internal/database"
	"testing"
)

func TestMain(m *testing.M) {
	code, err := run(m)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

func run(m *testing.M) (code int, err error) {
	SetUpTestEnv()
	db.InitTestDb()
	db.MakeMigrations()

	defer teardown()

	return m.Run(), nil
}

func teardown() {
	TestTeardown()
}
