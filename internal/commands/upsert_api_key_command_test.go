package commands

import (
	"receipt-wrangler/api/internal/utils"
	"testing"
)

func TestUpsertApiKeyCommand_Validate_ValidInputs(t *testing.T) {
	tests := map[string]struct {
		command UpsertApiKeyCommand
		valid   bool
	}{
		"valid read scope": {
			command: UpsertApiKeyCommand{
				Name:        "Test Key",
				Description: "Test description",
				Scope:       "r",
			},
			valid: true,
		},
		"valid write scope": {
			command: UpsertApiKeyCommand{
				Name:        "Test Key",
				Description: "Test description",
				Scope:       "w",
			},
			valid: true,
		},
		"valid read-write scope": {
			command: UpsertApiKeyCommand{
				Name:        "Test Key",
				Description: "Test description",
				Scope:       "rw",
			},
			valid: true,
		},
		"valid without description": {
			command: UpsertApiKeyCommand{
				Name:  "Test Key",
				Scope: "r",
			},
			valid: true,
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			vErr := test.command.Validate()

			if test.valid && len(vErr.Errors) > 0 {
				utils.PrintTestError(t, len(vErr.Errors), 0)
			}

			if !test.valid && len(vErr.Errors) == 0 {
				utils.PrintTestError(t, len(vErr.Errors), "greater than 0")
			}
		})
	}
}

func TestUpsertApiKeyCommand_Validate_InvalidInputs(t *testing.T) {
	tests := map[string]struct {
		command       UpsertApiKeyCommand
		expectedError string
	}{
		"missing name": {
			command: UpsertApiKeyCommand{
				Description: "Test description",
				Scope:       "r",
			},
			expectedError: "name",
		},
		"missing scope": {
			command: UpsertApiKeyCommand{
				Name:        "Test Key",
				Description: "Test description",
			},
			expectedError: "scope",
		},
		"invalid scope - empty string": {
			command: UpsertApiKeyCommand{
				Name:        "Test Key",
				Description: "Test description",
				Scope:       "",
			},
			expectedError: "scope",
		},
		"invalid scope - read": {
			command: UpsertApiKeyCommand{
				Name:        "Test Key",
				Description: "Test description",
				Scope:       "read",
			},
			expectedError: "scope",
		},
		"invalid scope - write": {
			command: UpsertApiKeyCommand{
				Name:        "Test Key",
				Description: "Test description",
				Scope:       "write",
			},
			expectedError: "scope",
		},
		"invalid scope - admin": {
			command: UpsertApiKeyCommand{
				Name:        "Test Key",
				Description: "Test description",
				Scope:       "admin",
			},
			expectedError: "scope",
		},
		"invalid scope - random string": {
			command: UpsertApiKeyCommand{
				Name:        "Test Key",
				Description: "Test description",
				Scope:       "invalid",
			},
			expectedError: "scope",
		},
		"invalid scope - mixed case": {
			command: UpsertApiKeyCommand{
				Name:        "Test Key",
				Description: "Test description",
				Scope:       "R",
			},
			expectedError: "scope",
		},
		"invalid scope - with spaces": {
			command: UpsertApiKeyCommand{
				Name:        "Test Key",
				Description: "Test description",
				Scope:       " r ",
			},
			expectedError: "scope",
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			vErr := test.command.Validate()

			if len(vErr.Errors) == 0 {
				utils.PrintTestError(t, len(vErr.Errors), "greater than 0")
			}

			if _, exists := vErr.Errors[test.expectedError]; !exists {
				utils.PrintTestError(t, "error should exist for field", test.expectedError)
			}
		})
	}
}

func TestUpsertApiKeyCommand_Validate_MultipleErrors(t *testing.T) {
	command := UpsertApiKeyCommand{
		Description: "Test description",
		Scope:       "invalid",
	}

	vErr := command.Validate()

	if len(vErr.Errors) != 2 {
		utils.PrintTestError(t, len(vErr.Errors), 2)
	}

	if _, exists := vErr.Errors["name"]; !exists {
		utils.PrintTestError(t, "error should exist for field", "name")
	}

	if _, exists := vErr.Errors["scope"]; !exists {
		utils.PrintTestError(t, "error should exist for field", "scope")
	}

	if vErr.Errors["name"] != "Name is required" {
		utils.PrintTestError(t, vErr.Errors["name"], "Name is required")
	}

	if vErr.Errors["scope"] != "Scope must be one of: r, w, rw" {
		utils.PrintTestError(t, vErr.Errors["scope"], "Scope must be one of: r, w, rw")
	}
}

func TestUpsertApiKeyCommand_Validate_EmptyCommand(t *testing.T) {
	command := UpsertApiKeyCommand{}

	vErr := command.Validate()

	if len(vErr.Errors) != 2 {
		utils.PrintTestError(t, len(vErr.Errors), 2)
	}

	if _, exists := vErr.Errors["name"]; !exists {
		utils.PrintTestError(t, "error should exist for field", "name")
	}

	if _, exists := vErr.Errors["scope"]; !exists {
		utils.PrintTestError(t, "error should exist for field", "scope")
	}
}
