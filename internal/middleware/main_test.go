package middleware

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
	defer teardown()
	repositories.SetUpTestEnv()
	repositories.InitTestDb()
	repositories.MakeMigrations()
	return m.Run(), nil
}

func teardown() {
	repositories.TestTeardown()
}
