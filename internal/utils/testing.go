package utils

import (
	"os"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"testing"

	db "receipt-wrangler/api/internal/database"
)

func PrintTestError(t *testing.T, actual any, expected any) {
	t.Errorf("Expected %s, but got %s", expected, actual)
}

func SetUpTestEnv() {
	os.Args = append(os.Args, "-env=test")
	config.SetConfigs()
	logging.InitLog()
	db.InitTestDb()
	db.MakeMigrations()
}
