package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strings"
	"testing"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

func tearDownSystemSettingsTest() {
	repositories.TruncateTestDb()
}

func TestShouldNotAllowUserToGetSystemSettings(t *testing.T) {
	defer tearDownSystemSettingsTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	expectedStatusCode := http.StatusForbidden

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	GetSystemSettings(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func TestShouldGetSystemSettingsWhenThereAreNoSettings(t *testing.T) {
	defer tearDownSystemSettingsTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	expectedStatusCode := http.StatusOK

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	GetSystemSettings(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func TestShouldGetSystemSettingsWhenThereAreExistingSettings(t *testing.T) {
	defer tearDownSystemSettingsTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	expectedStatusCode := http.StatusOK

	db := repositories.GetDB()
	db.Create(&models.SystemSettings{})

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	GetSystemSettings(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func TestShouldValidateUpsertSystemSettingsCommand(t *testing.T) {
	defer tearDownSystemEmailTest()
	db := repositories.GetDB()
	db.Create(&models.SystemSettings{})
	db.Create(&models.Prompt{})
	db.Create(&models.ReceiptProcessingSettings{
		Name:     "test",
		PromptId: 1,
	})
	db.Create(&models.ReceiptProcessingSettings{
		Name:     "test2",
		PromptId: 1,
	})

	defaultAsynqConfigCommands := make([]commands.UpsertTaskQueueConfigurationCommand, 0)
	for _, config := range models.GetAllDefaultQueueConfigurations() {
		defaultAsynqConfigCommands = append(defaultAsynqConfigCommands, commands.UpsertTaskQueueConfigurationCommand{
			Name:     config.Name,
			Priority: 1,
		})
	}

	id := uint(1)
	id2 := uint(2)

	tests := map[string]struct {
		input  commands.UpsertSystemSettingsCommand
		expect int
	}{
		"empty body": {
			expect: http.StatusBadRequest,
		},
		"empty command": {
			input:  commands.UpsertSystemSettingsCommand{},
			expect: http.StatusBadRequest,
		},
		"invalid email polling interval": {
			input: commands.UpsertSystemSettingsCommand{
				EmailPollingInterval: -1,
			},
			expect: http.StatusBadRequest,
		},
		"invalid receipt processing settings ID": {
			input: commands.UpsertSystemSettingsCommand{
				ReceiptProcessingSettingsId: new(uint),
			},
			expect: http.StatusBadRequest,
		},
		"invalid fallback receipt processing settings ID": {
			input: commands.UpsertSystemSettingsCommand{
				FallbackReceiptProcessingSettingsId: new(uint),
			},
			expect: http.StatusBadRequest,
		},
		"fallback receipt processing settings ID without receipt processing settings ID": {
			input: commands.UpsertSystemSettingsCommand{
				FallbackReceiptProcessingSettingsId: &id,
			},
			expect: http.StatusBadRequest,
		},
		"fallback receipt processing settings ID same as receipt processing settings ID": {
			input: commands.UpsertSystemSettingsCommand{
				ReceiptProcessingSettingsId:         &id,
				FallbackReceiptProcessingSettingsId: &id,
			},
			expect: http.StatusBadRequest,
		},
		"bad num workers": {
			input: commands.UpsertSystemSettingsCommand{
				EmailPollingInterval:                1,
				EnableLocalSignUp:                   true,
				ReceiptProcessingSettingsId:         &id,
				FallbackReceiptProcessingSettingsId: &id2,
				NumWorkers:                          0,
				CurrencyThousandthsSeparator:        models.COMMA,
				CurrencyDecimalSeparator:            models.DOT,
				CurrencySymbolPosition:              models.START,
			},
			expect: http.StatusBadRequest,
		},
		"missing currency thousandths separator": {
			input: commands.UpsertSystemSettingsCommand{
				EmailPollingInterval:                1,
				EnableLocalSignUp:                   true,
				ReceiptProcessingSettingsId:         &id,
				FallbackReceiptProcessingSettingsId: &id2,
				NumWorkers:                          1,
				CurrencyDecimalSeparator:            models.DOT,
				CurrencySymbolPosition:              models.START,
				TaskConcurrency:                     10,
				TaskQueueConfigurations:             defaultAsynqConfigCommands,
			},
			expect: http.StatusBadRequest,
		},
		"missing currency decimal separator": {
			input: commands.UpsertSystemSettingsCommand{
				EmailPollingInterval:                1,
				EnableLocalSignUp:                   true,
				ReceiptProcessingSettingsId:         &id,
				FallbackReceiptProcessingSettingsId: &id2,
				NumWorkers:                          1,
				CurrencyThousandthsSeparator:        models.COMMA,
				CurrencySymbolPosition:              models.START,
				TaskConcurrency:                     10,
				TaskQueueConfigurations:             defaultAsynqConfigCommands,
			},
			expect: http.StatusBadRequest,
		},
		"missing missing currency symbol position": {
			input: commands.UpsertSystemSettingsCommand{
				EmailPollingInterval:                1,
				EnableLocalSignUp:                   true,
				ReceiptProcessingSettingsId:         &id,
				FallbackReceiptProcessingSettingsId: &id2,
				NumWorkers:                          1,
				CurrencyThousandthsSeparator:        models.COMMA,
				CurrencyDecimalSeparator:            models.DOT,
				TaskConcurrency:                     10,
				TaskQueueConfigurations:             defaultAsynqConfigCommands,
			},
			expect: http.StatusBadRequest,
		},
		"valid command": {
			input: commands.UpsertSystemSettingsCommand{
				EmailPollingInterval:                1,
				CurrencyDisplay:                     "something else",
				EnableLocalSignUp:                   true,
				ReceiptProcessingSettingsId:         &id,
				FallbackReceiptProcessingSettingsId: &id2,
				NumWorkers:                          1,
				CurrencyThousandthsSeparator:        models.COMMA,
				CurrencyDecimalSeparator:            models.DOT,
				CurrencySymbolPosition:              models.START,
				TaskConcurrency:                     10,
				TaskQueueConfigurations:             defaultAsynqConfigCommands,
			},
			expect: http.StatusOK,
		},
	}

	for _, test := range tests {
		bytes, _ := json.Marshal(test.input)
		reader := strings.NewReader(string(bytes))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api", reader)

		newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
		r = r.WithContext(newContext)

		UpdateSystemSettings(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, w.Result().StatusCode, test.expect)
		}
	}
}
