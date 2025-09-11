package models

import (
	"receipt-wrangler/api/internal/utils"
	"testing"
)

func TestApiKeyScope_IsValid(t *testing.T) {
	validTests := []ApiKeyScope{
		API_KEY_SCOPE_READ,
		API_KEY_SCOPE_WRITE,
		API_KEY_SCOPE_READ_WRITE,
	}

	for _, scope := range validTests {
		t.Run(string(scope), func(t *testing.T) {
			if !scope.IsValid() {
				utils.PrintTestError(t, scope.IsValid(), true)
			}
		})
	}
}

func TestApiKeyScope_IsValid_InvalidScopes(t *testing.T) {
	invalidTests := []ApiKeyScope{
		"",
		"invalid",
		"read",
		"write",
		"admin",
		"R",
		"W",
		"RW",
		" r ",
		"r,w",
	}

	for _, scope := range invalidTests {
		t.Run(string(scope), func(t *testing.T) {
			if scope.IsValid() {
				utils.PrintTestError(t, scope.IsValid(), false)
			}
		})
	}
}

func TestApiKeyScope_Scan(t *testing.T) {
	var scope ApiKeyScope
	err := scope.Scan("r")

	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	if scope != API_KEY_SCOPE_READ {
		utils.PrintTestError(t, scope, API_KEY_SCOPE_READ)
	}
}
