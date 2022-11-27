package utils

import (
	"fmt"
	"os"
	db "receipt-wrangler/api/internal/database"
	config "receipt-wrangler/api/internal/env"
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
	os.Args = append(os.Args, "-env=test")
	config.SetConfig()
	containerId := db.InitTestDb()
	db.Connect()
	db.MakeMigrations()

	defer func() {
		db.TeardownTestDb(containerId)
	}()

	return m.Run(), nil
}

func TestInitTokenValidatorReturnsValidator(t *testing.T) {
	v, _ := InitTokenValidator()

	if v == nil {
		printTestError(t, v, "instance of validator")
	}
}

func TestGenerateJWTGeneratesTokens(t *testing.T) {
	GenerateJWT(1)
}
