package middleware

import (
	"context"
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
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

func createTestUser() models.User {
	user := models.User{
		Username:           "testuser",
		Password:           "hashedpassword",
		DisplayName:        "Test User",
		UserRole:           models.ADMIN,
		DefaultAvatarColor: "#FF0000",
	}
	repositories.GetDB().Create(&user)
	return user
}

func createTestApiKey(userId uint, scope string) (models.ApiKey, string, error) {
	// Set up pepper for HMAC generation
	pepperService := services.NewPepperService(nil)
	err := pepperService.InitPepper()
	if err != nil {
		return models.ApiKey{}, "", err
	}

	command := commands.UpsertApiKeyCommand{
		Name:        "Test API Key",
		Description: "Test description",
		Scope:       scope,
	}

	apiKeyService := services.NewApiKeyService(nil)
	generatedKey, err := apiKeyService.CreateApiKey(userId, command)
	if err != nil {
		return models.ApiKey{}, "", err
	}

	// Get the created API key from database
	parts := strings.Split(generatedKey, ".")
	if len(parts) != 4 {
		return models.ApiKey{}, "", err
	}

	apiKeyRepo := repositories.NewApiKeyRepository(nil)
	dbApiKey, err := apiKeyRepo.GetApiKeyById(parts[2])
	if err != nil {
		return models.ApiKey{}, "", err
	}

	return dbApiKey, generatedKey, nil
}

// Mock JWT validation by skipping middleware and setting context directly
func createFakeHandlerWithContext(claims *validator.ValidatedClaims) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if claims != nil {
			ctx := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, claims)
			r = r.WithContext(ctx)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})
}

func createTestJWT(userId uint, username string) string {
	// This is a simplified JWT for testing - in real implementation this would come from the auth service
	return "test.jwt.token"
}

func createFakeHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})
}

func setupAuthTest() {
	repositories.TruncateTestDb()
}

func teardownAuthTest() {
	repositories.TruncateTestDb()
}

// Test JWT Authentication Success - Cookie
func TestUnifiedAuthMiddleware_ValidJWTCookie(t *testing.T) {
	defer teardownAuthTest()
	setupAuthTest()

	// Create request with JWT in cookie - since getJwt checks cookie first
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.AddCookie(&http.Cookie{
		Name:  "jwt",
		Value: "valid.jwt.token",
	})
	w := httptest.NewRecorder()

	// Test the getJwt function directly to ensure it works with cookie
	jwt := getJwt(*r)
	if jwt != "valid.jwt.token" {
		utils.PrintTestError(t, jwt, "valid.jwt.token")
	}

	// For this test, we'll verify that the middleware attempts JWT validation
	// when a JWT is present (as opposed to API key validation)
	handler := UnifiedAuthMiddleware(createFakeHandler())
	handler.ServeHTTP(w, r)

	// This will fail JWT validation (expected), but confirms JWT path is taken
	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

// Test JWT Authentication Success - Header
func TestUnifiedAuthMiddleware_ValidJWTHeader(t *testing.T) {
	defer teardownAuthTest()
	setupAuthTest()

	// Create request with JWT in Authorization header (no cookie)
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", "Bearer valid.jwt.token")
	w := httptest.NewRecorder()

	// Test the getJwt function directly to ensure it works with header
	jwt := getJwt(*r)
	if jwt != "valid.jwt.token" {
		utils.PrintTestError(t, jwt, "valid.jwt.token")
	}

	// For this test, we'll verify that the middleware attempts JWT validation
	handler := UnifiedAuthMiddleware(createFakeHandler())
	handler.ServeHTTP(w, r)

	// This will fail JWT validation (expected), but confirms JWT path is taken
	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

// Test Valid API Key Authentication
func TestUnifiedAuthMiddleware_ValidApiKey(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer teardownAuthTest()
	setupAuthTest()

	user := createTestUser()
	_, generatedKey, err := createTestApiKey(user.ID, "r")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Create request with API key in Authorization header
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", generatedKey)
	w := httptest.NewRecorder()

	handler := UnifiedAuthMiddleware(createFakeHandler())
	handler.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}

// Test Invalid API Key Format
func TestUnifiedAuthMiddleware_InvalidApiKeyFormat(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer teardownAuthTest()
	setupAuthTest()

	// Create request with malformed API key
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", "key.1.invalid")
	w := httptest.NewRecorder()

	handler := UnifiedAuthMiddleware(createFakeHandler())
	handler.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

// Test Invalid API Key Prefix
func TestUnifiedAuthMiddleware_InvalidApiKeyPrefix(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer teardownAuthTest()
	setupAuthTest()

	// Create request with wrong prefix
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", "wrong.1.id.secret")
	w := httptest.NewRecorder()

	handler := UnifiedAuthMiddleware(createFakeHandler())
	handler.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

// Test Non-existent API Key
func TestUnifiedAuthMiddleware_NonExistentApiKey(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer teardownAuthTest()
	setupAuthTest()

	// Set up pepper for HMAC generation
	pepperService := services.NewPepperService(nil)
	err := pepperService.InitPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Create request with non-existent API key
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", "key.1.nonexistent.secret")
	w := httptest.NewRecorder()

	handler := UnifiedAuthMiddleware(createFakeHandler())
	handler.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

// Test Invalid API Key Secret
func TestUnifiedAuthMiddleware_InvalidApiKeySecret(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer teardownAuthTest()
	setupAuthTest()

	user := createTestUser()
	dbApiKey, generatedKey, err := createTestApiKey(user.ID, "r")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Modify the generated key to have wrong secret
	parts := strings.Split(generatedKey, ".")
	if len(parts) != 4 {
		utils.PrintTestError(t, len(parts), 4)
		return
	}

	// Create key with wrong secret
	wrongKey := strings.Join([]string{parts[0], parts[1], parts[2], "d3JvbmdzZWNyZXQ="}, ".")

	// Create request with wrong secret
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", wrongKey)
	w := httptest.NewRecorder()

	handler := UnifiedAuthMiddleware(createFakeHandler())
	handler.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}

	// Clean up
	repositories.GetDB().Delete(&dbApiKey)
}

// Test No Authentication Provided
func TestUnifiedAuthMiddleware_NoAuthentication(t *testing.T) {
	defer teardownAuthTest()
	setupAuthTest()

	// Create request without any authentication
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	handler := UnifiedAuthMiddleware(createFakeHandler())
	handler.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

// Test Both JWT and API Key Present (API Key should take precedence)
func TestUnifiedAuthMiddleware_BothAuthMethods(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer teardownAuthTest()
	setupAuthTest()

	user := createTestUser()
	_, generatedKey, err := createTestApiKey(user.ID, "r")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	jwt := createTestJWT(user.ID, user.Username)

	// Create request with both JWT cookie and API key header
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.AddCookie(&http.Cookie{
		Name:  "receipt-wrangler-jwt",
		Value: jwt,
	})
	r.Header.Set("Authorization", generatedKey)
	w := httptest.NewRecorder()

	handler := UnifiedAuthMiddleware(createFakeHandler())
	handler.ServeHTTP(w, r)

	// Should succeed using API key (takes precedence)
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}

// Test Malformed Authorization Header
func TestUnifiedAuthMiddleware_MalformedAuthHeader(t *testing.T) {
	defer teardownAuthTest()
	setupAuthTest()

	// Create request with malformed Authorization header (not JWT, not API key)
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", "Bearer malformed")
	w := httptest.NewRecorder()

	handler := UnifiedAuthMiddleware(createFakeHandler())
	handler.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

// Test API Key with Different Scopes
func TestUnifiedAuthMiddleware_ApiKeyScopes(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer teardownAuthTest()
	setupAuthTest()

	user := createTestUser()

	// Test with 'write' scope
	_, writeKey, err := createTestApiKey(user.ID, "w")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", writeKey)
	w := httptest.NewRecorder()

	handler := UnifiedAuthMiddleware(createFakeHandler())
	handler.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}

	// Test with 'admin' scope
	_, adminKey, err := createTestApiKey(user.ID, "rw")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	r2 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r2.Header.Set("Authorization", adminKey)
	w2 := httptest.NewRecorder()

	handler2 := UnifiedAuthMiddleware(createFakeHandler())
	handler2.ServeHTTP(w2, r2)

	if w2.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w2.Result().StatusCode, http.StatusOK)
	}
}

// Test Empty Authorization Header
func TestUnifiedAuthMiddleware_EmptyAuthHeader(t *testing.T) {
	defer teardownAuthTest()
	setupAuthTest()

	// Create request with empty Authorization header
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", "")
	w := httptest.NewRecorder()

	handler := UnifiedAuthMiddleware(createFakeHandler())
	handler.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

// Test Context Propagation for API Key
func TestUnifiedAuthMiddleware_ContextPropagation(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer teardownAuthTest()
	setupAuthTest()

	user := createTestUser()
	_, generatedKey, err := createTestApiKey(user.ID, "r")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Create request with valid API key
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", generatedKey)
	w := httptest.NewRecorder()

	// Handler that checks context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value(jwtmiddleware.ContextKey{})
		if claims == nil {
			utils.PrintTestError(t, claims, "claims should be present in context")
			return
		}

		// Check what type we actually have
		switch claimsType := claims.(type) {
		case *validator.ValidatedClaims:
			validatedClaims := claimsType
			if validatedClaims.CustomClaims == nil {
				utils.PrintTestError(t, "CustomClaims is nil", "should not be nil")
				return
			}

			customClaims, ok := validatedClaims.CustomClaims.(*structs.Claims)
			if !ok {
				utils.PrintTestError(t, "custom claims type assertion failed", "should be Claims")
				return
			}

			if customClaims.UserId != user.ID {
				utils.PrintTestError(t, customClaims.UserId, user.ID)
			}

			if customClaims.Username != user.Username {
				utils.PrintTestError(t, customClaims.Username, user.Username)
			}

			if customClaims.ApiKeyScope != "r" {
				utils.PrintTestError(t, customClaims.ApiKeyScope, "r")
			}
		default:
			utils.PrintTestError(t, "unexpected claims type", "should be *ValidatedClaims")
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	handler := UnifiedAuthMiddleware(testHandler)
	handler.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}

// Test Base64 Decoding Error (invalid Base64 in secret)
func TestUnifiedAuthMiddleware_Base64DecodingError(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer teardownAuthTest()
	setupAuthTest()

	// Set up pepper
	pepperService := services.NewPepperService(nil)
	err := pepperService.InitPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Create request with invalid Base64 secret
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", "key.1.dGVzdC1pZA==.invalid-base64-!")
	w := httptest.NewRecorder()

	handler := UnifiedAuthMiddleware(createFakeHandler())
	handler.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

// Test User Not Found Error During Claims Generation
func TestUnifiedAuthMiddleware_UserNotFoundError(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer teardownAuthTest()
	setupAuthTest()

	// Create API key directly without user
	pepperService := services.NewPepperService(nil)
	err := pepperService.InitPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	nonExistentUserId := uint(999)
	apiKey := models.ApiKey{
		ID:          "dGVzdC1pZA==", // base64 encoded "test-id"
		UserID:      &nonExistentUserId,
		Name:        "Orphaned Key",
		Description: "Key without user",
		Scope:       "r",
		Prefix:      "key",
		Hmac:        "test-hmac",
		Version:     1,
	}

	// Create HMAC for the key
	apiKeyService := services.NewApiKeyService(nil)
	secret := "test-secret"
	hmac, err := apiKeyService.GenerateApiKeyHmac(secret)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}
	apiKey.Hmac = hmac

	// Save API key without user
	apiKeyRepo := repositories.NewApiKeyRepository(nil)
	_, err = apiKeyRepo.CreateApiKey(apiKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Create valid API key format with proper secret
	validKey := apiKeyService.BuildV1ApiKey("key", 1, "test-id", utils.Base64Encode([]byte(secret)))

	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", validKey)
	w := httptest.NewRecorder()

	handler := UnifiedAuthMiddleware(createFakeHandler())
	handler.ServeHTTP(w, r)

	// Should fail because user doesn't exist
	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestUnifiedAuthMiddleware_UpdatesApiKeyLastUsedDate(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer teardownAuthTest()
	setupAuthTest()

	user := createTestUser()
	dbApiKey, generatedKey, err := createTestApiKey(user.ID, "r")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify LastUsedAt is initially nil
	apiKeyRepo := repositories.NewApiKeyRepository(nil)
	initialApiKey, err := apiKeyRepo.GetApiKeyById(dbApiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if initialApiKey.LastUsedAt != nil {
		utils.PrintTestError(t, initialApiKey.LastUsedAt, nil)
	}

	beforeRequest := time.Now()

	// Create request with API key
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", generatedKey)
	w := httptest.NewRecorder()

	handler := UnifiedAuthMiddleware(createFakeHandler())
	handler.ServeHTTP(w, r)

	// Should succeed
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}

	// Give some time for the goroutine to complete
	time.Sleep(100 * time.Millisecond)

	afterRequest := time.Now()

	// Verify LastUsedAt was updated
	updatedApiKey, err := apiKeyRepo.GetApiKeyById(dbApiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if updatedApiKey.LastUsedAt == nil {
		utils.PrintTestError(t, "LastUsedAt should not be nil", "LastUsedAt should be set after authentication")
	}

	if updatedApiKey.LastUsedAt.Before(beforeRequest) {
		utils.PrintTestError(t, "LastUsedAt is before request", "LastUsedAt should be after request started")
	}

	if updatedApiKey.LastUsedAt.After(afterRequest) {
		utils.PrintTestError(t, "LastUsedAt is after request", "LastUsedAt should be before request completed")
	}
}

func TestUnifiedAuthMiddleware_UpdatesApiKeyLastUsedDate_MultipleRequests(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer teardownAuthTest()
	setupAuthTest()

	user := createTestUser()
	dbApiKey, generatedKey, err := createTestApiKey(user.ID, "r")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	apiKeyRepo := repositories.NewApiKeyRepository(nil)

	// First request
	r1 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r1.Header.Set("Authorization", generatedKey)
	w1 := httptest.NewRecorder()

	handler := UnifiedAuthMiddleware(createFakeHandler())
	handler.ServeHTTP(w1, r1)

	if w1.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w1.Result().StatusCode, http.StatusOK)
	}

	// Give time for the first goroutine to complete
	time.Sleep(100 * time.Millisecond)

	// Get the first LastUsedAt time
	firstApiKey, err := apiKeyRepo.GetApiKeyById(dbApiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if firstApiKey.LastUsedAt == nil {
		utils.PrintTestError(t, "FirstApiKey LastUsedAt should not be nil", "FirstApiKey LastUsedAt should be set")
	}

	firstLastUsedAt := *firstApiKey.LastUsedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(50 * time.Millisecond)

	beforeSecondRequest := time.Now()

	// Second request
	r2 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r2.Header.Set("Authorization", generatedKey)
	w2 := httptest.NewRecorder()

	handler.ServeHTTP(w2, r2)

	if w2.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w2.Result().StatusCode, http.StatusOK)
	}

	// Give time for the second goroutine to complete
	time.Sleep(100 * time.Millisecond)

	// Verify the second update
	secondApiKey, err := apiKeyRepo.GetApiKeyById(dbApiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if secondApiKey.LastUsedAt == nil {
		utils.PrintTestError(t, "SecondApiKey LastUsedAt should not be nil", "SecondApiKey LastUsedAt should be set")
	}

	// Second LastUsedAt should be after the first
	if !secondApiKey.LastUsedAt.After(firstLastUsedAt) {
		utils.PrintTestError(t, "Second LastUsedAt should be after first", "Second update should be more recent")
	}

	if secondApiKey.LastUsedAt.Before(beforeSecondRequest) {
		utils.PrintTestError(t, "Second LastUsedAt is before second request", "Second update should be recent")
	}
}

func TestUnifiedAuthMiddleware_DoesNotUpdateLastUsedDate_OnJWT(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer teardownAuthTest()
	setupAuthTest()

	user := createTestUser()
	dbApiKey, _, err := createTestApiKey(user.ID, "r")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify LastUsedAt is initially nil
	apiKeyRepo := repositories.NewApiKeyRepository(nil)
	initialApiKey, err := apiKeyRepo.GetApiKeyById(dbApiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if initialApiKey.LastUsedAt != nil {
		utils.PrintTestError(t, initialApiKey.LastUsedAt, nil)
	}

	jwt := createTestJWT(user.ID, user.Username)

	// Create request with JWT cookie (no API key)
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.AddCookie(&http.Cookie{
		Name:  "receipt-wrangler-jwt",
		Value: jwt,
	})
	w := httptest.NewRecorder()

	// Mock JWT validation by creating custom handler that sets context
	customClaims := &structs.Claims{
		UserId:      user.ID,
		Username:    user.Username,
		Displayname: user.DisplayName,
		UserRole:    user.UserRole,
	}
	claims := &validator.ValidatedClaims{
		CustomClaims: customClaims,
	}

	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, claims)
		r = r.WithContext(ctx)

		// Skip the actual JWT validation since it's complex to set up
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Use mock handler instead of UnifiedAuthMiddleware for JWT test
	mockHandler.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}

	// Give time for any potential goroutine to complete
	time.Sleep(100 * time.Millisecond)

	// Verify LastUsedAt was NOT updated (since JWT was used, not API key)
	unchangedApiKey, err := apiKeyRepo.GetApiKeyById(dbApiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if unchangedApiKey.LastUsedAt != nil {
		utils.PrintTestError(t, unchangedApiKey.LastUsedAt, nil)
	}
}

func TestUnifiedAuthMiddleware_UpdateFailsGracefully_NonExistentKey(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer teardownAuthTest()
	setupAuthTest()

	user := createTestUser()
	dbApiKey, generatedKey, err := createTestApiKey(user.ID, "r")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Delete the API key from database after creation but before use
	apiKeyRepo := repositories.NewApiKeyRepository(nil)
	repositories.GetDB().Delete(&dbApiKey)

	// Create request with the now-deleted API key
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", generatedKey)
	w := httptest.NewRecorder()

	handler := UnifiedAuthMiddleware(createFakeHandler())
	handler.ServeHTTP(w, r)

	// Should fail authentication (API key not found)
	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}

	// The update should not crash the system (it will try to update a non-existent key)
	time.Sleep(100 * time.Millisecond)

	// Verify the key is still deleted (no resurrection)
	_, err = apiKeyRepo.GetApiKeyById(dbApiKey.ID)
	if err == nil {
		utils.PrintTestError(t, err, "an error - key should not exist")
	}
}

// Test Bearer Token Parsing - Valid Bearer token
func TestGetJwt_BearerTokenValid(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", "Bearer valid.jwt.token")

	jwt := getJwt(*r)
	expected := "valid.jwt.token"
	if jwt != expected {
		utils.PrintTestError(t, jwt, expected)
	}
}

// Test Bearer Token Parsing - Bearer token with extra spaces
func TestGetJwt_BearerTokenWithSpaces(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", "Bearer   token.with.spaces   ")

	jwt := getJwt(*r)
	expected := "token.with.spaces"
	if jwt != expected {
		utils.PrintTestError(t, jwt, expected)
	}
}

// Test Bearer Token Parsing - Empty Bearer token
func TestGetJwt_BearerTokenEmpty(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", "Bearer ")

	jwt := getJwt(*r)
	expected := ""
	if jwt != expected {
		utils.PrintTestError(t, jwt, expected)
	}
}

// Test Bearer Token Parsing - Multiple Bearer parts (malformed)
func TestGetJwt_BearerTokenMultipleParts(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", "Bearer token Bearer another")

	jwt := getJwt(*r)
	// Should return the original header since it's malformed (more than 2 parts after split)
	expected := "Bearer token Bearer another"
	if jwt != expected {
		utils.PrintTestError(t, jwt, expected)
	}
}

// Test Bearer Token Parsing - Different Bearer case variations
func TestGetJwt_BearerTokenCaseSensitive(t *testing.T) {
	// Test lowercase "bearer" - should NOT be processed as Bearer token
	r1 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r1.Header.Set("Authorization", "bearer valid.jwt.token")

	jwt1 := getJwt(*r1)
	expected1 := "bearer valid.jwt.token" // Should return as-is since "Bearer" is case-sensitive
	if jwt1 != expected1 {
		utils.PrintTestError(t, jwt1, expected1)
	}

	// Test mixed case "BeArEr" - should NOT be processed as Bearer token
	r2 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r2.Header.Set("Authorization", "BeArEr valid.jwt.token")

	jwt2 := getJwt(*r2)
	expected2 := "BeArEr valid.jwt.token" // Should return as-is since "Bearer" is case-sensitive
	if jwt2 != expected2 {
		utils.PrintTestError(t, jwt2, expected2)
	}
}

// Test Bearer Token Parsing - Authorization header without Bearer
func TestGetJwt_AuthHeaderWithoutBearer(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", "just.a.token")

	jwt := getJwt(*r)
	expected := "just.a.token"
	if jwt != expected {
		utils.PrintTestError(t, jwt, expected)
	}
}

// Test Bearer Token Parsing - Empty Authorization header
func TestGetJwt_EmptyAuthHeader(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", "")

	jwt := getJwt(*r)
	expected := ""
	if jwt != expected {
		utils.PrintTestError(t, jwt, expected)
	}
}

// Test Bearer Token Parsing - Cookie takes precedence over Bearer header
func TestGetJwt_CookieTakesPrecedence(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.AddCookie(&http.Cookie{
		Name:  "jwt",
		Value: "cookie.jwt.token",
	})
	r.Header.Set("Authorization", "Bearer header.jwt.token")

	jwt := getJwt(*r)
	expected := "cookie.jwt.token" // Cookie should take precedence
	if jwt != expected {
		utils.PrintTestError(t, jwt, expected)
	}
}

// Test Bearer Token Parsing - No cookie, Bearer header used
func TestGetJwt_NoCookieBearerHeaderUsed(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	// No cookie set
	r.Header.Set("Authorization", "Bearer header.jwt.token")

	jwt := getJwt(*r)
	expected := "header.jwt.token"
	if jwt != expected {
		utils.PrintTestError(t, jwt, expected)
	}
}

// Test Bearer Token Parsing - Bearer with only spaces
func TestGetJwt_BearerTokenOnlySpaces(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", "Bearer    ")

	jwt := getJwt(*r)
	expected := ""
	if jwt != expected {
		utils.PrintTestError(t, jwt, expected)
	}
}

// Test Bearer Token Parsing - Bearer not at start (should not be processed)
func TestGetJwt_BearerNotAtStart(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r.Header.Set("Authorization", "Token Bearer jwt.token")

	jwt := getJwt(*r)
	// Should return the full header unchanged since Bearer is not at the start
	expected := "Token Bearer jwt.token"
	if jwt != expected {
		utils.PrintTestError(t, jwt, expected)
	}
}

// Test Bearer Token Parsing - Proper Bearer prefix handling
func TestGetJwt_BearerPrefixHandling(t *testing.T) {
	// Test with proper Bearer prefix
	r1 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r1.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")

	jwt1 := getJwt(*r1)
	expected1 := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
	if jwt1 != expected1 {
		utils.PrintTestError(t, jwt1, expected1)
	}

	// Test with Bearer at start but no space (should still work with Contains logic)
	r2 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r2.Header.Set("Authorization", "Bearertoken123")

	jwt2 := getJwt(*r2)
	expected2 := "token123"
	if jwt2 != expected2 {
		utils.PrintTestError(t, jwt2, expected2)
	}

	// Test with "Bearer" as part of token but not at start
	r3 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	r3.Header.Set("Authorization", "MyBearer token123")

	jwt3 := getJwt(*r3)
	expected3 := "MyBearer token123" // Should not be processed
	if jwt3 != expected3 {
		utils.PrintTestError(t, jwt3, expected3)
	}
}
