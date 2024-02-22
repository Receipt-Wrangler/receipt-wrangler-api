package repositories

import (
	"fmt"
	"os"
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
	SetUpTestEnv()
	InitTestDb()
	MakeMigrations()
	return m.Run(), nil
}

func teardown() {
	TestTeardown()
}
