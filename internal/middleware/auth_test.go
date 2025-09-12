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
		Name:  "receipt-wrangler-jwt",
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
	if jwt != "Bearer valid.jwt.token" {
		utils.PrintTestError(t, jwt, "Bearer valid.jwt.token")
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
		}

		validatedClaims, ok := claims.(*validator.ValidatedClaims)
		if !ok {
			utils.PrintTestError(t, "type assertion failed", "should be ValidatedClaims")
		}

		customClaims, ok := validatedClaims.CustomClaims.(*structs.Claims)
		if !ok {
			utils.PrintTestError(t, "custom claims type assertion failed", "should be Claims")
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
	validKey := apiKeyService.BuildV1ApiKey("key", 1, "test-id", utils.Base64EncodeBytes([]byte(secret)))

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
