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

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}})
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

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}})
	r = r.WithContext(newContext)

	grantAllAppPerms(t, 1)

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

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}})
	r = r.WithContext(newContext)

	grantAllAppPerms(t, 1)

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

	grantAllAppPerms(t, 1)

	for _, test := range tests {
		bytes, _ := json.Marshal(test.input)
		reader := strings.NewReader(string(bytes))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api", reader)

		newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}})
		r = r.WithContext(newContext)

		UpdateSystemSettings(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, w.Result().StatusCode, test.expect)
		}
	}
}

// Regression: the update writes every column via Select("*"), so a PUT body that
// omits a configured lifetime used to persist 0 and silently reset an admin's
// session length to the default. Driven through the handler with a RAW JSON body
// rather than a typed command, because the whole point is which keys are absent
// from the wire -- a struct literal cannot express that.
func TestUpdateSystemSettingsPreservesOmittedRefreshTokenLifetimes(t *testing.T) {
	defer tearDownSystemSettingsTest()

	db := repositories.GetDB()
	db.Create(&models.SystemSettings{})
	grantAllAppPerms(t, 1)

	// An admin has configured both lifetimes away from the defaults.
	err := db.Model(&models.SystemSettings{}).
		Where("id = ?", 1).
		Updates(map[string]interface{}{
			"refresh_token_valid_for_hours":     720,
			"mcp_refresh_token_valid_for_hours": 6,
			"temp_file_retention_hours":         1080,
			"secure_cookies":                    true,
		}).Error
	if err != nil {
		t.Fatalf("failed to seed configured lifetimes: %v", err)
	}

	queueConfigs := make([]map[string]interface{}, 0)
	for _, config := range models.GetAllDefaultQueueConfigurations() {
		queueConfigs = append(queueConfigs, map[string]interface{}{
			"name":     config.Name,
			"priority": 1,
		})
	}

	// A valid body that simply does not mention either lifetime key.
	body := map[string]interface{}{
		"currencyDisplay":              "$",
		"currencySymbolPosition":       models.START,
		"currencyThousandthsSeparator": models.COMMA,
		"currencyDecimalSeparator":     models.DOT,
		"currencyHideDecimalPlaces":    false,
		"taskConcurrency":              1,
		"emailPollingInterval":         60,
		"taskQueueConfigurations":      queueConfigs,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal body: %v", err)
	}

	r := httptest.NewRequest("PUT", "/api", strings.NewReader(string(bodyBytes)))
	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}})
	r = r.WithContext(newContext)
	w := httptest.NewRecorder()

	UpdateSystemSettings(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}

	updated, err := repositories.NewSystemSettingsRepository(nil).GetSystemSettings()
	if err != nil {
		t.Fatalf("failed to read back system settings: %v", err)
	}

	if updated.RefreshTokenValidForHours != 720 {
		utils.PrintTestError(t, updated.RefreshTokenValidForHours, 720)
	}

	if updated.McpRefreshTokenValidForHours != 6 {
		utils.PrintTestError(t, updated.McpRefreshTokenValidForHours, 6)
	}

	// Same hazard, same fix: the retention window is a pointer on the command and
	// is dropped from the UPDATE when the key is absent.
	if updated.TempFileRetentionHours != 1080 {
		utils.PrintTestError(t, updated.TempFileRetentionHours, 1080)
	}

	// A security toggle has the most to lose here: a client that does not know
	// the key must not silently turn Secure cookies back off.
	if !updated.SecureCookies {
		utils.PrintTestError(t, updated.SecureCookies, true)
	}
}

// The flip side: an explicitly sent value must still be written, so the merge
// above cannot be mistaken for "these fields are read-only".
func TestUpdateSystemSettingsPersistsExplicitRefreshTokenLifetimes(t *testing.T) {
	defer tearDownSystemSettingsTest()

	db := repositories.GetDB()
	db.Create(&models.SystemSettings{})
	grantAllAppPerms(t, 1)

	queueConfigs := make([]commands.UpsertTaskQueueConfigurationCommand, 0)
	for _, config := range models.GetAllDefaultQueueConfigurations() {
		queueConfigs = append(queueConfigs, commands.UpsertTaskQueueConfigurationCommand{
			Name:     config.Name,
			Priority: 1,
		})
	}

	appHours := 168
	mcpHours := 12
	secureCookies := true
	command := commands.UpsertSystemSettingsCommand{
		CurrencyDisplay:              "$",
		CurrencySymbolPosition:       models.START,
		CurrencyThousandthsSeparator: models.COMMA,
		CurrencyDecimalSeparator:     models.DOT,
		TaskConcurrency:              1,
		EmailPollingInterval:         60,
		TaskQueueConfigurations:      queueConfigs,
		RefreshTokenValidForHours:    &appHours,
		McpRefreshTokenValidForHours: &mcpHours,
		SecureCookies:                &secureCookies,
	}
	bodyBytes, err := json.Marshal(command)
	if err != nil {
		t.Fatalf("failed to marshal command: %v", err)
	}

	r := httptest.NewRequest("PUT", "/api", strings.NewReader(string(bodyBytes)))
	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}})
	r = r.WithContext(newContext)
	w := httptest.NewRecorder()

	UpdateSystemSettings(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}

	updated, err := repositories.NewSystemSettingsRepository(nil).GetSystemSettings()
	if err != nil {
		t.Fatalf("failed to read back system settings: %v", err)
	}

	if updated.RefreshTokenValidForHours != appHours {
		utils.PrintTestError(t, updated.RefreshTokenValidForHours, appHours)
	}

	if updated.McpRefreshTokenValidForHours != mcpHours {
		utils.PrintTestError(t, updated.McpRefreshTokenValidForHours, mcpHours)
	}

	if !updated.SecureCookies {
		utils.PrintTestError(t, updated.SecureCookies, true)
	}
}

// The end-to-end shape of the upgrade bug: an install whose task_queue_configuration
// table predates models.SystemCleanUpQueue has four rows, the settings form builds
// its inputs from whatever the GET returned, and so it submits four --
// UpsertSystemSettingsCommand.Validate requires one per queue name and rejected the
// entire body with a 400. tempFileRetentionHours is the setting that made this
// reachable in practice: changing it is exactly the save that failed.
func TestUpdateSystemSettingsSucceedsOnAnInstallMissingAQueueConfiguration(t *testing.T) {
	defer tearDownSystemSettingsTest()

	db := repositories.GetDB()
	db.Create(&models.SystemSettings{})
	grantAllAppPerms(t, 1)

	for _, queueName := range models.GetQueueNames() {
		if queueName == models.SystemCleanUpQueue {
			continue
		}

		err := db.Create(&models.TaskQueueConfiguration{
			Name:             queueName,
			Priority:         3,
			SystemSettingsId: 1,
		}).Error
		if err != nil {
			t.Fatalf("failed to seed queue configuration %s: %v", queueName, err)
		}
	}

	// Build the queue list the way the settings form does: from the payload the
	// server just handed the client, never from the client's own enum.
	served, err := repositories.NewSystemSettingsRepository(nil).GetSystemSettings()
	if err != nil {
		t.Fatalf("failed to read system settings: %v", err)
	}

	queueConfigs := make([]map[string]interface{}, 0)
	for _, config := range served.TaskQueueConfigurations {
		queueConfigs = append(queueConfigs, map[string]interface{}{
			"name":     config.Name,
			"priority": config.Priority,
		})
	}

	body := map[string]interface{}{
		"currencyDisplay":              "$",
		"currencySymbolPosition":       models.START,
		"currencyThousandthsSeparator": models.COMMA,
		"currencyDecimalSeparator":     models.DOT,
		"currencyHideDecimalPlaces":    false,
		"taskConcurrency":              1,
		"emailPollingInterval":         60,
		"tempFileRetentionHours":       4320,
		"taskQueueConfigurations":      queueConfigs,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal body: %v", err)
	}

	r := httptest.NewRequest("PUT", "/api", strings.NewReader(string(bodyBytes)))
	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}})
	r = r.WithContext(newContext)
	w := httptest.NewRecorder()

	UpdateSystemSettings(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}

	// A 200 that dropped the value would be the worse outcome of the two, so the
	// setting the save exists to change is asserted alongside the status.
	updated, err := repositories.NewSystemSettingsRepository(nil).GetSystemSettings()
	if err != nil {
		t.Fatalf("failed to read back system settings: %v", err)
	}

	if updated.TempFileRetentionHours != 4320 {
		utils.PrintTestError(t, updated.TempFileRetentionHours, 4320)
	}

	var persistedCount int64
	if err := db.Model(&models.TaskQueueConfiguration{}).Count(&persistedCount).Error; err != nil {
		t.Fatalf("failed to count queue configurations: %v", err)
	}

	// And the row heals, so the next read no longer has to invent it.
	if persistedCount != int64(len(models.GetQueueNames())) {
		utils.PrintTestError(t, persistedCount, len(models.GetQueueNames()))
	}
}
