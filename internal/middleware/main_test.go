package middleware

import (
	"fmt"
	"os"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/utils"
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
	utils.SetUpTestEnv()
	InitMiddlewareLogger()
	db.InitTestDb()
	db.MakeMigrations()

	defer func() {
	}()

	return m.Run(), nil
}
