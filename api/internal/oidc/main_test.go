package oidc

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

	// The OIDC provider's client secret is encrypted at rest, so this package needs
	// a key the way internal/ai and internal/middleware do. Set once for the whole
	// package rather than per test, because the fixtures that create providers run
	// outside any single test's t.Setenv scope.
	if err := os.Setenv("ENCRYPTION_KEY", "oidc-test-encryption-key"); err != nil {
		return 1, err
	}

	repositories.SetUpTestEnv()
	repositories.InitTestDb()
	if err := repositories.MakeMigrations(); err != nil {
		return 1, err
	}
	return m.Run(), nil
}

func teardown() {
	repositories.TestTeardown()
}
