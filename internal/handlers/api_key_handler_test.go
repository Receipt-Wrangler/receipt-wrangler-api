package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strings"
	"testing"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/go-chi/chi/v5"
)

func tearDownApiKeyHandlerTests() {
	repositories.TruncateTestDb()
}

// Helper function to create test pepper for API key creation
func createTestPepper() {
	pepperService := services.NewPepperService(nil)
	err := pepperService.InitPepper()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize pepper: %v", err))
	}
}

// Helper function to create JWT context for tests
func createJWTContext(r *http.Request, userId uint, userRole models.UserRole) *http.Request {
	newContext := context.WithValue(
		r.Context(),
		jwtmiddleware.ContextKey{},
		&validator.ValidatedClaims{
			CustomClaims: &structs.Claims{
				UserId:   userId,
				UserRole: userRole,
			},
		},
	)
	return r.WithContext(newContext)
}

// Helper function to create test API keys for pagination tests
func createTestApiKeysForHandlerTests() {
	db := repositories.GetDB()
	user1 := uint(1)
	user2 := uint(2)

	apiKeys := []models.ApiKey{
		{
			ID:          "handler-key-1",
			UserID:      &user1,
			CreatedBy:   &user1,
			Name:        "Alpha Handler Key",
			Description: "First handler test key",
			Prefix:      "key",
			Hmac:        "handler-hmac-1",
			Version:     1,
			Scope:       "r",
		},
		{
			ID:          "handler-key-2",
			UserID:      &user1,
			CreatedBy:   &user1,
			Name:        "Beta Handler Key",
			Description: "Second handler test key",
			Prefix:      "key",
			Hmac:        "handler-hmac-2",
			Version:     1,
			Scope:       "rw",
		},
		{
			ID:          "handler-key-3",
			UserID:      &user2,
			CreatedBy:   &user2,
			Name:        "Gamma Handler Key",
			Description: "Third handler test key",
			Prefix:      "key",
			Hmac:        "handler-hmac-3",
			Version:     1,
			Scope:       "w",
		},
	}

	for _, apiKey := range apiKeys {
		db.Create(&apiKey)
	}
}

// CreateApiKey Handler Tests

func TestCreateApiKey_Success(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer tearDownApiKeyHandlerTests()
	createTestPepper()

	command := commands.UpsertApiKeyCommand{
		Name:        "Test API Key",
		Description: "Test description",
		Scope:       "r",
	}

	bytes, _ := json.Marshal(command)
	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/api-keys", reader)
	r = createJWTContext(r, 1, models.USER)

	CreateApiKey(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}

	// Verify response contains the generated API key
	var response structs.ApiKeyResult
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if response.Key == "" {
		utils.PrintTestError(t, "empty key", "non-empty key")
	}

	// Verify key format (should be key.1.{id}.{secret})
	parts := strings.Split(response.Key, ".")
	if len(parts) != 4 {
		utils.PrintTestError(t, len(parts), 4)
	}

	if parts[0] != "key" {
		utils.PrintTestError(t, parts[0], "key")
	}

	if parts[1] != "1" {
		utils.PrintTestError(t, parts[1], "1")
	}
}

func TestCreateApiKey_MinimalFields(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer tearDownApiKeyHandlerTests()
	createTestPepper()

	command := commands.UpsertApiKeyCommand{
		Name:  "Minimal Key",
		Scope: "w",
	}

	bytes, _ := json.Marshal(command)
	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/api-keys", reader)
	r = createJWTContext(r, 1, models.USER)

	CreateApiKey(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}

func TestCreateApiKey_AllValidScopes(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer tearDownApiKeyHandlerTests()
	createTestPepper()

	scopes := []string{"r", "w", "rw"}

	for _, scope := range scopes {
		command := commands.UpsertApiKeyCommand{
			Name:        fmt.Sprintf("Key with scope %s", scope),
			Description: fmt.Sprintf("Testing scope %s", scope),
			Scope:       scope,
		}

		bytes, _ := json.Marshal(command)
		reader := strings.NewReader(string(bytes))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/api-keys", reader)
		r = createJWTContext(r, 1, models.USER)

		CreateApiKey(w, r)

		if w.Result().StatusCode != http.StatusOK {
			utils.PrintTestError(t, w.Result().StatusCode, fmt.Sprintf("http.StatusOK for scope %s", scope))
		}
	}
}

func TestCreateApiKey_ValidationErrors(t *testing.T) {
	defer tearDownApiKeyHandlerTests()

	tests := map[string]struct {
		input  commands.UpsertApiKeyCommand
		expect int
	}{
		"missing name": {
			input: commands.UpsertApiKeyCommand{
				Description: "Missing name",
				Scope:       "r",
			},
			expect: http.StatusBadRequest,
		},
		"missing scope": {
			input: commands.UpsertApiKeyCommand{
				Name:        "Missing scope",
				Description: "No scope provided",
			},
			expect: http.StatusBadRequest,
		},
		"invalid scope": {
			input: commands.UpsertApiKeyCommand{
				Name:        "Invalid scope",
				Description: "Bad scope value",
				Scope:       "invalid",
			},
			expect: http.StatusBadRequest,
		},
		"empty name": {
			input: commands.UpsertApiKeyCommand{
				Name:        "",
				Description: "Empty name",
				Scope:       "r",
			},
			expect: http.StatusBadRequest,
		},
		"empty scope": {
			input: commands.UpsertApiKeyCommand{
				Name:        "Empty scope",
				Description: "Empty scope value",
				Scope:       "",
			},
			expect: http.StatusBadRequest,
		},
	}

	for name, test := range tests {
		bytes, _ := json.Marshal(test.input)
		reader := strings.NewReader(string(bytes))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/api-keys", reader)
		r = createJWTContext(r, 1, models.USER)

		CreateApiKey(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, w.Result().StatusCode, fmt.Sprintf("%s expected %d", name, test.expect))
		}
	}
}

func TestCreateApiKey_MalformedJSON(t *testing.T) {
	defer tearDownApiKeyHandlerTests()

	reader := strings.NewReader("{invalid json")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/api-keys", reader)
	r = createJWTContext(r, 1, models.USER)

	CreateApiKey(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestCreateApiKey_EmptyBody(t *testing.T) {
	defer tearDownApiKeyHandlerTests()

	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/api-keys", reader)
	r = createJWTContext(r, 1, models.USER)

	CreateApiKey(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestCreateApiKey_AsAdmin(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer tearDownApiKeyHandlerTests()
	createTestPepper()

	command := commands.UpsertApiKeyCommand{
		Name:        "Admin API Key",
		Description: "Created by admin",
		Scope:       "rw",
	}

	bytes, _ := json.Marshal(command)
	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/api-keys", reader)
	r = createJWTContext(r, 1, models.ADMIN)

	CreateApiKey(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}

// GetPagedApiKeys Handler Tests

func TestGetPagedApiKeys_Success(t *testing.T) {
	defer tearDownApiKeyHandlerTests()
	createTestApiKeysForHandlerTests()

	command := commands.PagedApiKeyRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:          1,
			PageSize:      10,
			OrderBy:       "name",
			SortDirection: commands.ASCENDING,
		},
		ApiKeyFilter: commands.ApiKeyFilter{
			AssociatedApiKeys: commands.ASSOCIATED_API_KEYS_MINE,
		},
	}

	bytes, _ := json.Marshal(command)
	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/api-keys/paged", reader)
	r = createJWTContext(r, 1, models.USER)

	GetPagedApiKeys(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}

	var response structs.PagedData
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// User 1 should have 2 API keys
	if response.TotalCount != 2 {
		utils.PrintTestError(t, response.TotalCount, int64(2))
	}

	if len(response.Data) != 2 {
		utils.PrintTestError(t, len(response.Data), 2)
	}
}

func TestGetPagedApiKeys_AdminViewAll(t *testing.T) {
	defer tearDownApiKeyHandlerTests()
	createTestApiKeysForHandlerTests()

	command := commands.PagedApiKeyRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:          1,
			PageSize:      10,
			OrderBy:       "name",
			SortDirection: commands.ASCENDING,
		},
		ApiKeyFilter: commands.ApiKeyFilter{
			AssociatedApiKeys: commands.ASSOCIATED_API_KEYS_ALL,
		},
	}

	bytes, _ := json.Marshal(command)
	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/api-keys/paged", reader)
	r = createJWTContext(r, 1, models.ADMIN)

	GetPagedApiKeys(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}

	var response structs.PagedData
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Should see all 3 API keys
	if response.TotalCount != 3 {
		utils.PrintTestError(t, response.TotalCount, int64(3))
	}
}

func TestGetPagedApiKeys_UserCannotViewAll(t *testing.T) {
	defer tearDownApiKeyHandlerTests()
	createTestApiKeysForHandlerTests()

	command := commands.PagedApiKeyRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:          1,
			PageSize:      10,
			OrderBy:       "name",
			SortDirection: commands.ASCENDING,
		},
		ApiKeyFilter: commands.ApiKeyFilter{
			AssociatedApiKeys: commands.ASSOCIATED_API_KEYS_ALL,
		},
	}

	bytes, _ := json.Marshal(command)
	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/api-keys/paged", reader)
	r = createJWTContext(r, 1, models.USER)

	GetPagedApiKeys(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusBadRequest)
	}
}

func TestGetPagedApiKeys_Pagination(t *testing.T) {
	defer tearDownApiKeyHandlerTests()
	createTestApiKeysForHandlerTests()

	tests := map[string]struct {
		page     int
		pageSize int
		expected int
	}{
		"first page size 1": {
			page:     1,
			pageSize: 1,
			expected: 1,
		},
		"second page size 1": {
			page:     2,
			pageSize: 1,
			expected: 1,
		},
		"page size 2": {
			page:     1,
			pageSize: 2,
			expected: 2,
		},
		"no limit": {
			page:     1,
			pageSize: -1,
			expected: 2,
		},
	}

	for name, test := range tests {
		command := commands.PagedApiKeyRequestCommand{
			PagedRequestCommand: commands.PagedRequestCommand{
				Page:          test.page,
				PageSize:      test.pageSize,
				OrderBy:       "name",
				SortDirection: commands.ASCENDING,
			},
			ApiKeyFilter: commands.ApiKeyFilter{
				AssociatedApiKeys: commands.ASSOCIATED_API_KEYS_MINE,
			},
		}

		bytes, _ := json.Marshal(command)
		reader := strings.NewReader(string(bytes))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/api-keys/paged", reader)
		r = createJWTContext(r, 1, models.USER)

		GetPagedApiKeys(w, r)

		if w.Result().StatusCode != http.StatusOK {
			utils.PrintTestError(t, w.Result().StatusCode, fmt.Sprintf("http.StatusOK for test %s", name))
		}

		var response structs.PagedData
		err := json.NewDecoder(w.Body).Decode(&response)
		if err != nil {
			utils.PrintTestError(t, err, "no error")
		}

		if len(response.Data) != test.expected {
			utils.PrintTestError(t, len(response.Data), fmt.Sprintf("%d for test %s", test.expected, name))
		}
	}
}

func TestGetPagedApiKeys_Sorting(t *testing.T) {
	defer tearDownApiKeyHandlerTests()
	createTestApiKeysForHandlerTests()

	validColumns := []string{"name", "description", "created_at", "updated_at"}
	sortDirections := []commands.SortDirection{commands.ASCENDING, commands.DESCENDING}

	for _, column := range validColumns {
		for _, direction := range sortDirections {
			command := commands.PagedApiKeyRequestCommand{
				PagedRequestCommand: commands.PagedRequestCommand{
					Page:          1,
					PageSize:      10,
					OrderBy:       column,
					SortDirection: direction,
				},
				ApiKeyFilter: commands.ApiKeyFilter{
					AssociatedApiKeys: commands.ASSOCIATED_API_KEYS_MINE,
				},
			}

			bytes, _ := json.Marshal(command)
			reader := strings.NewReader(string(bytes))
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/api/api-keys/paged", reader)
			r = createJWTContext(r, 1, models.USER)

			GetPagedApiKeys(w, r)

			if w.Result().StatusCode != http.StatusOK {
				utils.PrintTestError(t, w.Result().StatusCode, fmt.Sprintf("http.StatusOK for column %s direction %s", column, direction))
			}
		}
	}
}

func TestGetPagedApiKeys_ValidationErrors(t *testing.T) {
	defer tearDownApiKeyHandlerTests()

	tests := map[string]struct {
		input  commands.PagedApiKeyRequestCommand
		expect int
	}{
		"missing associated api keys": {
			input: commands.PagedApiKeyRequestCommand{
				PagedRequestCommand: commands.PagedRequestCommand{
					Page:          1,
					PageSize:      10,
					OrderBy:       "name",
					SortDirection: commands.ASCENDING,
				},
				ApiKeyFilter: commands.ApiKeyFilter{},
			},
			expect: http.StatusBadRequest,
		},
		"invalid page": {
			input: commands.PagedApiKeyRequestCommand{
				PagedRequestCommand: commands.PagedRequestCommand{
					Page:          0,
					PageSize:      10,
					OrderBy:       "name",
					SortDirection: commands.ASCENDING,
				},
				ApiKeyFilter: commands.ApiKeyFilter{
					AssociatedApiKeys: commands.ASSOCIATED_API_KEYS_MINE,
				},
			},
			expect: http.StatusBadRequest,
		},
		"invalid page size": {
			input: commands.PagedApiKeyRequestCommand{
				PagedRequestCommand: commands.PagedRequestCommand{
					Page:          1,
					PageSize:      0,
					OrderBy:       "name",
					SortDirection: commands.ASCENDING,
				},
				ApiKeyFilter: commands.ApiKeyFilter{
					AssociatedApiKeys: commands.ASSOCIATED_API_KEYS_MINE,
				},
			},
			expect: http.StatusBadRequest,
		},
		"invalid sort direction": {
			input: commands.PagedApiKeyRequestCommand{
				PagedRequestCommand: commands.PagedRequestCommand{
					Page:          1,
					PageSize:      10,
					OrderBy:       "name",
					SortDirection: "invalid",
				},
				ApiKeyFilter: commands.ApiKeyFilter{
					AssociatedApiKeys: commands.ASSOCIATED_API_KEYS_MINE,
				},
			},
			expect: http.StatusBadRequest,
		},
	}

	for name, test := range tests {
		bytes, _ := json.Marshal(test.input)
		reader := strings.NewReader(string(bytes))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/api-keys/paged", reader)
		r = createJWTContext(r, 1, models.USER)

		GetPagedApiKeys(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, w.Result().StatusCode, fmt.Sprintf("%s expected %d", name, test.expect))
		}
	}
}

func TestGetPagedApiKeys_InvalidOrderByColumn(t *testing.T) {
	defer tearDownApiKeyHandlerTests()
	createTestApiKeysForHandlerTests()

	command := commands.PagedApiKeyRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:          1,
			PageSize:      10,
			OrderBy:       "invalid_column",
			SortDirection: commands.ASCENDING,
		},
		ApiKeyFilter: commands.ApiKeyFilter{
			AssociatedApiKeys: commands.ASSOCIATED_API_KEYS_MINE,
		},
	}

	bytes, _ := json.Marshal(command)
	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/api-keys/paged", reader)
	r = createJWTContext(r, 1, models.USER)

	GetPagedApiKeys(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestGetPagedApiKeys_MalformedJSON(t *testing.T) {
	defer tearDownApiKeyHandlerTests()

	reader := strings.NewReader("{invalid json")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/api-keys/paged", reader)
	r = createJWTContext(r, 1, models.USER)

	GetPagedApiKeys(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestGetPagedApiKeys_EmptyBody(t *testing.T) {
	defer tearDownApiKeyHandlerTests()

	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/api-keys/paged", reader)
	r = createJWTContext(r, 1, models.USER)

	GetPagedApiKeys(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestGetPagedApiKeys_EmptyDatabase(t *testing.T) {
	defer tearDownApiKeyHandlerTests()

	command := commands.PagedApiKeyRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:          1,
			PageSize:      10,
			OrderBy:       "name",
			SortDirection: commands.ASCENDING,
		},
		ApiKeyFilter: commands.ApiKeyFilter{
			AssociatedApiKeys: commands.ASSOCIATED_API_KEYS_MINE,
		},
	}

	bytes, _ := json.Marshal(command)
	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/api-keys/paged", reader)
	r = createJWTContext(r, 1, models.USER)

	GetPagedApiKeys(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}

	var response structs.PagedData
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if response.TotalCount != 0 {
		utils.PrintTestError(t, response.TotalCount, int64(0))
	}

	if len(response.Data) != 0 {
		utils.PrintTestError(t, len(response.Data), 0)
	}
}

func TestGetPagedApiKeys_DifferentUsers(t *testing.T) {
	defer tearDownApiKeyHandlerTests()
	createTestApiKeysForHandlerTests()

	// Test user 1 (should see 2 keys)
	command := commands.PagedApiKeyRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:          1,
			PageSize:      10,
			OrderBy:       "name",
			SortDirection: commands.ASCENDING,
		},
		ApiKeyFilter: commands.ApiKeyFilter{
			AssociatedApiKeys: commands.ASSOCIATED_API_KEYS_MINE,
		},
	}

	bytes, _ := json.Marshal(command)
	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/api-keys/paged", reader)
	r = createJWTContext(r, 1, models.USER)

	GetPagedApiKeys(w, r)

	var response1 structs.PagedData
	json.NewDecoder(w.Body).Decode(&response1)

	// Test user 2 (should see 1 key)
	reader = strings.NewReader(string(bytes))
	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/api/api-keys/paged", reader)
	r = createJWTContext(r, 2, models.USER)

	GetPagedApiKeys(w, r)

	var response2 structs.PagedData
	json.NewDecoder(w.Body).Decode(&response2)

	if response1.TotalCount != 2 {
		utils.PrintTestError(t, response1.TotalCount, int64(2))
	}

	if response2.TotalCount != 1 {
		utils.PrintTestError(t, response2.TotalCount, int64(1))
	}
}

// UpdateApiKey Handler Tests

func TestUpdateApiKey_Success(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer tearDownApiKeyHandlerTests()
	createTestPepper()
	createTestApiKeysForHandlerTests()

	command := commands.UpsertApiKeyCommand{
		Name:        "Updated Test API Key",
		Description: "Updated test description",
		Scope:       "rw",
	}

	bytes, _ := json.Marshal(command)
	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/api-keys/handler-key-1", reader)
	r = createJWTContext(r, 1, models.USER)

	// Add URL parameter for key ID
	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext())
	r = r.WithContext(ctx)
	chi.URLParam(r, "id")
	rctx := chi.RouteContext(r.Context())
	rctx.URLParams.Add("id", "handler-key-1")

	UpdateApiKey(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}

func TestUpdateApiKey_MinimalFields(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer tearDownApiKeyHandlerTests()
	createTestPepper()
	createTestApiKeysForHandlerTests()

	command := commands.UpsertApiKeyCommand{
		Name:  "Minimal Update",
		Scope: "r",
	}

	bytes, _ := json.Marshal(command)
	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/api-keys/handler-key-1", reader)
	r = createJWTContext(r, 1, models.USER)

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext())
	r = r.WithContext(ctx)
	chi.URLParam(r, "id")
	rctx := chi.RouteContext(r.Context())
	rctx.URLParams.Add("id", "handler-key-1")

	UpdateApiKey(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}

func TestUpdateApiKey_AllValidScopes(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer tearDownApiKeyHandlerTests()
	createTestPepper()
	createTestApiKeysForHandlerTests()

	scopes := []string{"r", "w", "rw"}

	for i, scope := range scopes {
		keyId := fmt.Sprintf("handler-key-%d", (i%2)+1) // Alternate between key-1 and key-2

		command := commands.UpsertApiKeyCommand{
			Name:        fmt.Sprintf("Key with scope %s", scope),
			Description: fmt.Sprintf("Testing scope %s", scope),
			Scope:       scope,
		}

		bytes, _ := json.Marshal(command)
		reader := strings.NewReader(string(bytes))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("PUT", fmt.Sprintf("/api/api-keys/%s", keyId), reader)
		r = createJWTContext(r, 1, models.USER)

		ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext())
		r = r.WithContext(ctx)
		chi.URLParam(r, "id")
		rctx := chi.RouteContext(r.Context())
		rctx.URLParams.Add("id", keyId)

		UpdateApiKey(w, r)

		if w.Result().StatusCode != http.StatusOK {
			utils.PrintTestError(t, w.Result().StatusCode, fmt.Sprintf("http.StatusOK for scope %s", scope))
		}
	}
}

func TestUpdateApiKey_ValidationErrors(t *testing.T) {
	defer tearDownApiKeyHandlerTests()
	createTestApiKeysForHandlerTests()

	tests := map[string]struct {
		input  commands.UpsertApiKeyCommand
		expect int
	}{
		"missing name": {
			input: commands.UpsertApiKeyCommand{
				Description: "Missing name",
				Scope:       "r",
			},
			expect: http.StatusBadRequest,
		},
		"missing scope": {
			input: commands.UpsertApiKeyCommand{
				Name:        "Missing scope",
				Description: "No scope provided",
			},
			expect: http.StatusBadRequest,
		},
		"invalid scope": {
			input: commands.UpsertApiKeyCommand{
				Name:        "Invalid scope",
				Description: "Bad scope value",
				Scope:       "invalid",
			},
			expect: http.StatusBadRequest,
		},
		"empty name": {
			input: commands.UpsertApiKeyCommand{
				Name:        "",
				Description: "Empty name",
				Scope:       "r",
			},
			expect: http.StatusBadRequest,
		},
		"empty scope": {
			input: commands.UpsertApiKeyCommand{
				Name:        "Empty scope",
				Description: "Empty scope value",
				Scope:       "",
			},
			expect: http.StatusBadRequest,
		},
	}

	for name, test := range tests {
		bytes, _ := json.Marshal(test.input)
		reader := strings.NewReader(string(bytes))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("PUT", "/api/api-keys/handler-key-1", reader)
		r = createJWTContext(r, 1, models.USER)

		ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext())
		r = r.WithContext(ctx)
		chi.URLParam(r, "id")
		rctx := chi.RouteContext(r.Context())
		rctx.URLParams.Add("id", "handler-key-1")

		UpdateApiKey(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, w.Result().StatusCode, fmt.Sprintf("%s expected %d", name, test.expect))
		}
	}
}

func TestUpdateApiKey_MalformedJSON(t *testing.T) {
	defer tearDownApiKeyHandlerTests()
	createTestApiKeysForHandlerTests()

	reader := strings.NewReader("{invalid json")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/api-keys/handler-key-1", reader)
	r = createJWTContext(r, 1, models.USER)

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext())
	r = r.WithContext(ctx)
	chi.URLParam(r, "id")
	rctx := chi.RouteContext(r.Context())
	rctx.URLParams.Add("id", "handler-key-1")

	UpdateApiKey(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestUpdateApiKey_EmptyBody(t *testing.T) {
	defer tearDownApiKeyHandlerTests()
	createTestApiKeysForHandlerTests()

	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/api-keys/handler-key-1", reader)
	r = createJWTContext(r, 1, models.USER)

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext())
	r = r.WithContext(ctx)
	chi.URLParam(r, "id")
	rctx := chi.RouteContext(r.Context())
	rctx.URLParams.Add("id", "handler-key-1")

	UpdateApiKey(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestUpdateApiKey_NonExistentKey(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer tearDownApiKeyHandlerTests()
	createTestPepper()

	command := commands.UpsertApiKeyCommand{
		Name:        "Update Non-existent",
		Description: "Trying to update a key that doesn't exist",
		Scope:       "r",
	}

	bytes, _ := json.Marshal(command)
	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/api-keys/non-existent-key", reader)
	r = createJWTContext(r, 1, models.USER)

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext())
	r = r.WithContext(ctx)
	chi.URLParam(r, "id")
	rctx := chi.RouteContext(r.Context())
	rctx.URLParams.Add("id", "non-existent-key")

	UpdateApiKey(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestUpdateApiKey_WrongUser(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer tearDownApiKeyHandlerTests()
	createTestPepper()
	createTestApiKeysForHandlerTests()

	command := commands.UpsertApiKeyCommand{
		Name:        "Unauthorized Update",
		Description: "User 1 trying to update User 2's key",
		Scope:       "r",
	}

	bytes, _ := json.Marshal(command)
	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/api-keys/handler-key-3", reader) // handler-key-3 belongs to user 2
	r = createJWTContext(r, 1, models.USER)                                // But user 1 is making the request

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext())
	r = r.WithContext(ctx)
	chi.URLParam(r, "id")
	rctx := chi.RouteContext(r.Context())
	rctx.URLParams.Add("id", "handler-key-3")

	UpdateApiKey(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestUpdateApiKey_AsAdmin(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer tearDownApiKeyHandlerTests()
	createTestPepper()
	createTestApiKeysForHandlerTests()

	command := commands.UpsertApiKeyCommand{
		Name:        "Admin Updated Key",
		Description: "Updated by admin user",
		Scope:       "rw",
	}

	bytes, _ := json.Marshal(command)
	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/api-keys/handler-key-1", reader)
	r = createJWTContext(r, 1, models.ADMIN)

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext())
	r = r.WithContext(ctx)
	chi.URLParam(r, "id")
	rctx := chi.RouteContext(r.Context())
	rctx.URLParams.Add("id", "handler-key-1")

	UpdateApiKey(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}

func TestUpdateApiKey_UserUpdatingOwnKey(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer tearDownApiKeyHandlerTests()
	createTestPepper()
	createTestApiKeysForHandlerTests()

	// User 2 updating their own key (handler-key-3)
	command := commands.UpsertApiKeyCommand{
		Name:        "My Updated Key",
		Description: "User updating their own key",
		Scope:       "rw",
	}

	bytes, _ := json.Marshal(command)
	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/api-keys/handler-key-3", reader)
	r = createJWTContext(r, 2, models.USER) // User 2 updating their own key

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext())
	r = r.WithContext(ctx)
	chi.URLParam(r, "id")
	rctx := chi.RouteContext(r.Context())
	rctx.URLParams.Add("id", "handler-key-3")

	UpdateApiKey(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}

func TestUpdateApiKey_EmptyKeyId(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer tearDownApiKeyHandlerTests()
	createTestPepper()

	command := commands.UpsertApiKeyCommand{
		Name:        "Empty Key ID Test",
		Description: "Testing with empty key ID",
		Scope:       "r",
	}

	bytes, _ := json.Marshal(command)
	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/api-keys/", reader)
	r = createJWTContext(r, 1, models.USER)

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext())
	r = r.WithContext(ctx)
	chi.URLParam(r, "id")
	rctx := chi.RouteContext(r.Context())
	rctx.URLParams.Add("id", "")

	UpdateApiKey(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestUpdateApiKey_URLEncodedId(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer tearDownApiKeyHandlerTests()
	createTestPepper()
	createTestApiKeysForHandlerTests()

	// Create an API key with special characters that would be URL encoded
	keyId := "cFU4ckxyWWI3UEVYbDFvVUczTWJDLU5sWlFkUTltR1ZNcE9DOUpNZ1pSZz0="

	// Create the API key first
	apiKey := models.ApiKey{
		ID:          keyId,
		UserID:      &[]uint{1}[0],
		Name:        "Test API Key with Special Chars",
		Description: "Test description",
		Scope:       "rw",
		Prefix:      "rw-v1",
		Hmac:        "test-hmac",
		Version:     1,
		CreatedBy:   &[]uint{1}[0],
	}

	apiKeyRepository := repositories.NewApiKeyRepository(nil)
	_, err := apiKeyRepository.CreateApiKey(apiKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error creating test key")
		return
	}

	command := commands.UpsertApiKeyCommand{
		Name:        "Updated Test API Key",
		Description: "Updated test description",
		Scope:       "r",
	}

	bytes, _ := json.Marshal(command)
	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()

	// URL encode the key ID as the frontend would do
	encodedKeyId := "cFU4ckxyWWI3UEVYbDFvVUczTWJDLU5sWlFkUTltR1ZNcE9DOUpNZ1pSZz0%3D"
	r := httptest.NewRequest("PUT", "/api/api-keys/"+encodedKeyId, reader)
	r = createJWTContext(r, 1, models.USER)

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext())
	r = r.WithContext(ctx)
	chi.URLParam(r, "id")
	rctx := chi.RouteContext(r.Context())
	rctx.URLParams.Add("id", encodedKeyId)

	UpdateApiKey(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}

func TestDeleteApiKey_Success(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer tearDownApiKeyHandlerTests()
	createTestPepper()
	createTestApiKeysForHandlerTests()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/api/api-keys/handler-key-1", nil)
	r = createJWTContext(r, 1, models.USER) // User 1 deleting their own key

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext())
	r = r.WithContext(ctx)
	chi.URLParam(r, "id")
	rctx := chi.RouteContext(r.Context())
	rctx.URLParams.Add("id", "handler-key-1")

	DeleteApiKey(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}

func TestDeleteApiKey_AdminCanDeleteAnyKey(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer tearDownApiKeyHandlerTests()
	createTestPepper()
	createTestApiKeysForHandlerTests()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/api/api-keys/handler-key-3", nil)
	r = createJWTContext(r, 1, models.ADMIN) // Admin deleting another user's key

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext())
	r = r.WithContext(ctx)
	chi.URLParam(r, "id")
	rctx := chi.RouteContext(r.Context())
	rctx.URLParams.Add("id", "handler-key-3")

	DeleteApiKey(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}

func TestDeleteApiKey_UserCannotDeleteOtherUsersKey(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer tearDownApiKeyHandlerTests()
	createTestPepper()
	createTestApiKeysForHandlerTests()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/api/api-keys/handler-key-3", nil)
	r = createJWTContext(r, 1, models.USER) // User 1 trying to delete User 2's key

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext())
	r = r.WithContext(ctx)
	chi.URLParam(r, "id")
	rctx := chi.RouteContext(r.Context())
	rctx.URLParams.Add("id", "handler-key-3")

	DeleteApiKey(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestDeleteApiKey_NonExistentKey(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer tearDownApiKeyHandlerTests()
	createTestPepper()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/api/api-keys/non-existent-key", nil)
	r = createJWTContext(r, 1, models.USER)

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext())
	r = r.WithContext(ctx)
	chi.URLParam(r, "id")
	rctx := chi.RouteContext(r.Context())
	rctx.URLParams.Add("id", "non-existent-key")

	DeleteApiKey(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestDeleteApiKey_URLEncodedId(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer tearDownApiKeyHandlerTests()
	createTestPepper()

	// Create an API key with special characters that would be URL encoded
	keyId := "cFU4ckxyWWI3UEVYbDFvVUczTWJDLU5sWlFkUTltR1ZNcE9DOUpNZ1pSZz0="

	apiKey := models.ApiKey{
		ID:          keyId,
		UserID:      &[]uint{1}[0],
		Name:        "Test API Key for Delete",
		Description: "Test description",
		Scope:       "rw",
		Prefix:      "rw-v1",
		Hmac:        "test-hmac",
		Version:     1,
		CreatedBy:   &[]uint{1}[0],
	}

	apiKeyRepository := repositories.NewApiKeyRepository(nil)
	_, err := apiKeyRepository.CreateApiKey(apiKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error creating test key")
		return
	}

	w := httptest.NewRecorder()

	// URL encode the key ID as the frontend would do
	encodedKeyId := "cFU4ckxyWWI3UEVYbDFvVUczTWJDLU5sWlFkUTltR1ZNcE9DOUpNZ1pSZz0%3D"
	r := httptest.NewRequest("DELETE", "/api/api-keys/"+encodedKeyId, nil)
	r = createJWTContext(r, 1, models.USER)

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext())
	r = r.WithContext(ctx)
	chi.URLParam(r, "id")
	rctx := chi.RouteContext(r.Context())
	rctx.URLParams.Add("id", encodedKeyId)

	DeleteApiKey(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}
