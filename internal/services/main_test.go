package services

import (
	"fmt"
	"os"
	"receipt-wrangler/api/internal/repositories"
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
	repositories.SetUpTestEnv()
	repositories.InitTestDb()
	repositories.MakeMigrations()

	defer teardown()

	return m.Run(), nil
}

func teardown() {
	repositories.TestTeardown()
}
