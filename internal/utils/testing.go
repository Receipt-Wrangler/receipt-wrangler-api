package utils

import (
	"fmt"
	"os"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"testing"

	"gorm.io/gorm"
)

func PrintTestError(t *testing.T, actual any, expected any) {
	t.Errorf("Expected %s, but got %s", expected, actual)
}

func SetUpTestEnv() {
	os.Args = append(os.Args, "-env=test")
	config.SetConfigs()
	logging.InitLog()
}

func TruncateTable(db *gorm.DB, tableName string) error {
	query := fmt.Sprintf("DELETE FROM %s", tableName)
	return db.Exec(query).Error
}
