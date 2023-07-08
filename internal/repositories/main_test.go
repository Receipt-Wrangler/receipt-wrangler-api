package repositories

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
	db.InitTestDb()
	db.MakeMigrations()

	defer teardown()

	return m.Run(), nil
}

func teardown() {
	utils.TestTeardown()

}
