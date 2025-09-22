package services

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strings"
	"testing"
	"time"
)

func TestApiKeyService_CreateApiKey_Success(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	// Set up pepper for HMAC generation
	pepperService := NewPepperService(nil)
	err := pepperService.InitPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	userId := uint(1)
	command := commands.UpsertApiKeyCommand{
		Name:        "Test API Key",
		Description: "Test description",
		Scope:       "r",
	}

	apiKeyService := NewApiKeyService(nil)
	generatedKey, err := apiKeyService.CreateApiKey(userId, command)

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if generatedKey == "" {
		utils.PrintTestError(t, generatedKey, "a non-empty string")
	}

	// Verify key format: key.1.id.secret
	parts := strings.Split(generatedKey, ".")
	if len(parts) != 4 {
		utils.PrintTestError(t, len(parts), 4)
	}

	if parts[0] != "key" {
		utils.PrintTestError(t, parts[0], "key")
	}

	if parts[1] != "1" {
		utils.PrintTestError(t, parts[1], "1")
	}

	// Verify the API key was saved to the database
	var savedApiKey models.ApiKey
	err = repositories.GetDB().Where("user_id = ? AND name = ?", userId, command.Name).First(&savedApiKey).Error
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if savedApiKey.Name != command.Name {
		utils.PrintTestError(t, savedApiKey.Name, command.Name)
	}

	if savedApiKey.Description != command.Description {
		utils.PrintTestError(t, savedApiKey.Description, command.Description)
	}

	if savedApiKey.Scope != command.Scope {
		utils.PrintTestError(t, savedApiKey.Scope, command.Scope)
	}

	if *savedApiKey.UserID != userId {
		utils.PrintTestError(t, *savedApiKey.UserID, userId)
	}

	if savedApiKey.Prefix != "key" {
		utils.PrintTestError(t, savedApiKey.Prefix, "key")
	}

	if savedApiKey.Version != 1 {
		utils.PrintTestError(t, savedApiKey.Version, 1)
	}

	if savedApiKey.Hmac == "" {
		utils.PrintTestError(t, savedApiKey.Hmac, "a non-empty string")
	}

	if savedApiKey.CreatedBy == nil || *savedApiKey.CreatedBy != userId {
		utils.PrintTestError(t, savedApiKey.CreatedBy, &userId)
	}
}

func TestApiKeyService_CreateApiKey_HmacGenerationError(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "")
	defer repositories.TruncateTestDb()

	userId := uint(1)
	command := commands.UpsertApiKeyCommand{
		Name:        "Test API Key",
		Description: "Test description",
		Scope:       "r",
	}

	apiKeyService := NewApiKeyService(nil)
	_, err := apiKeyService.CreateApiKey(userId, command)

	if err == nil {
		utils.PrintTestError(t, err, "an error")
	}
}

func TestApiKeyService_GenerateApiKeyHmac_Success(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	// Set up pepper for HMAC generation
	pepperService := NewPepperService(nil)
	err := pepperService.InitPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	apiKeyService := NewApiKeyService(nil)
	secret := "test-secret"

	hmac, err := apiKeyService.GenerateApiKeyHmac(secret)

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if hmac == "" {
		utils.PrintTestError(t, hmac, "a non-empty string")
	}

	// Verify it's base64 encoded by checking it doesn't contain invalid characters
	if strings.Contains(hmac, " ") {
		utils.PrintTestError(t, "HMAC contains spaces", "HMAC should be base64 encoded")
	}
}

func TestApiKeyService_GenerateApiKeyHmac_PepperError(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "")
	defer repositories.TruncateTestDb()

	apiKeyService := NewApiKeyService(nil)
	secret := "test-secret"

	_, err := apiKeyService.GenerateApiKeyHmac(secret)

	if err == nil {
		utils.PrintTestError(t, err, "an error")
	}
}

func TestApiKeyService_GenerateApiKeyHmac_NoPepper(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	// Don't create pepper - this should cause an error

	apiKeyService := NewApiKeyService(nil)
	secret := "test-secret"

	_, err := apiKeyService.GenerateApiKeyHmac(secret)

	if err == nil {
		utils.PrintTestError(t, err, "an error")
	}
}

func TestApiKeyService_BuildV1ApiKey(t *testing.T) {
	apiKeyService := NewApiKeyService(nil)

	prefix := "key"
	version := 1
	id := "test-id"
	secret := "test-secret"

	result := apiKeyService.BuildV1ApiKey(prefix, version, id, secret)

	expected := "key.1.test-id.test-secret"
	if result != expected {
		utils.PrintTestError(t, result, expected)
	}
}

func TestApiKeyService_BuildV1ApiKey_DifferentVersion(t *testing.T) {
	apiKeyService := NewApiKeyService(nil)

	prefix := "api"
	version := 2
	id := "different-id"
	secret := "different-secret"

	result := apiKeyService.BuildV1ApiKey(prefix, version, id, secret)

	expected := "api.2.different-id.different-secret"
	if result != expected {
		utils.PrintTestError(t, result, expected)
	}
}

func TestApiKeyService_CreateApiKey_VerifyHmacGeneration(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	// Set up pepper for HMAC generation
	pepperService := NewPepperService(nil)
	err := pepperService.InitPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	userId := uint(1)
	command := commands.UpsertApiKeyCommand{
		Name:        "HMAC Test Key",
		Description: "Test HMAC generation",
		Scope:       "w",
	}

	apiKeyService := NewApiKeyService(nil)
	generatedKey, err := apiKeyService.CreateApiKey(userId, command)

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Extract secret from generated key
	parts := strings.Split(generatedKey, ".")
	if len(parts) != 4 {
		utils.PrintTestError(t, len(parts), 4)
	}
	b64Secret := parts[3]

	// Decode the base64 secret to get the raw secret
	decodedSecret, err := utils.Base64URLDecode(b64Secret)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Generate HMAC for the raw secret and compare
	expectedHmac, err := apiKeyService.GenerateApiKeyHmac(string(decodedSecret))
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Get the stored API key and compare HMAC
	var savedApiKey models.ApiKey
	err = repositories.GetDB().Where("user_id = ? AND name = ?", userId, command.Name).First(&savedApiKey).Error
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if savedApiKey.Hmac != expectedHmac {
		utils.PrintTestError(t, savedApiKey.Hmac, expectedHmac)
	}
}

func TestApiKeyService_CreateApiKey_MultipleKeys(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	// Set up pepper for HMAC generation
	pepperService := NewPepperService(nil)
	err := pepperService.InitPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	userId := uint(1)
	apiKeyService := NewApiKeyService(nil)

	// Create first API key
	command1 := commands.UpsertApiKeyCommand{
		Name:        "First Key",
		Description: "First test key",
		Scope:       "r",
	}

	key1, err := apiKeyService.CreateApiKey(userId, command1)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Create second API key
	command2 := commands.UpsertApiKeyCommand{
		Name:        "Second Key",
		Description: "Second test key",
		Scope:       "w",
	}

	key2, err := apiKeyService.CreateApiKey(userId, command2)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify keys are different
	if key1 == key2 {
		utils.PrintTestError(t, "Keys should be different", "Keys are identical")
	}

	// Verify both keys have correct structure
	parts1 := strings.Split(key1, ".")
	parts2 := strings.Split(key2, ".")

	if len(parts1) != 4 {
		utils.PrintTestError(t, len(parts1), 4)
	}

	if len(parts2) != 4 {
		utils.PrintTestError(t, len(parts2), 4)
	}

	// Verify IDs are different (base64 encoded)
	if parts1[2] == parts2[2] {
		utils.PrintTestError(t, "IDs should be different", "IDs are identical")
	}

	// Verify secrets are different
	if parts1[3] == parts2[3] {
		utils.PrintTestError(t, "Secrets should be different", "Secrets are identical")
	}

	// Verify both are saved in database
	var count int64
	repositories.GetDB().Model(&models.ApiKey{}).Where("user_id = ?", userId).Count(&count)
	if count != 2 {
		utils.PrintTestError(t, count, 2)
	}
}

func TestApiKeyService_ValidateV1ApiKey_Success(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	// Set up pepper for HMAC generation
	pepperService := NewPepperService(nil)
	err := pepperService.InitPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	userId := uint(1)
	command := commands.UpsertApiKeyCommand{
		Name:        "Validation Test Key",
		Description: "Test key validation",
		Scope:       "r",
	}

	apiKeyService := NewApiKeyService(nil)
	generatedKey, err := apiKeyService.CreateApiKey(userId, command)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
		return
	}

	// Validate the generated key using the service
	validatedKey, err := apiKeyService.ValidateV1ApiKey(generatedKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
		return
	}

	if validatedKey.Name != command.Name {
		utils.PrintTestError(t, validatedKey.Name, command.Name)
	}

	if validatedKey.Description != command.Description {
		utils.PrintTestError(t, validatedKey.Description, command.Description)
	}

	if validatedKey.Scope != command.Scope {
		utils.PrintTestError(t, validatedKey.Scope, command.Scope)
	}

	if validatedKey.UserID == nil || *validatedKey.UserID != userId {
		utils.PrintTestError(t, validatedKey.UserID, userId)
	}
}

func TestApiKeyService_ValidateV1ApiKey_InvalidFormat(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	apiKeyService := NewApiKeyService(nil)

	// Test with wrong number of parts
	_, err := apiKeyService.ValidateV1ApiKey("key.1.invalid")
	if err == nil {
		utils.PrintTestError(t, err, "an error for invalid format")
	}

	// Test with too many parts
	_, err = apiKeyService.ValidateV1ApiKey("key.1.id.secret.extra")
	if err == nil {
		utils.PrintTestError(t, err, "an error for invalid format")
	}

	// Test with empty string
	_, err = apiKeyService.ValidateV1ApiKey("")
	if err == nil {
		utils.PrintTestError(t, err, "an error for empty key")
	}
}

func TestApiKeyService_ValidateV1ApiKey_NonExistentId(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	apiKeyService := NewApiKeyService(nil)

	// Test with non-existent ID
	_, err := apiKeyService.ValidateV1ApiKey("key.1.bm9uZXhpc3RlbnQ=.c2VjcmV0")
	if err == nil {
		utils.PrintTestError(t, err, "an error for non-existent API key")
	}
}

func TestApiKeyService_ValidateV1ApiKey_InvalidSecret(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	// Set up pepper for HMAC generation
	pepperService := NewPepperService(nil)
	err := pepperService.InitPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	userId := uint(1)
	command := commands.UpsertApiKeyCommand{
		Name:        "Invalid Secret Test",
		Description: "Test invalid secret",
		Scope:       "r",
	}

	apiKeyService := NewApiKeyService(nil)
	generatedKey, err := apiKeyService.CreateApiKey(userId, command)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Modify the secret part to be invalid
	parts := strings.Split(generatedKey, ".")
	if len(parts) != 4 {
		utils.PrintTestError(t, len(parts), 4)
	}

	// Create key with wrong secret
	invalidKey := strings.Join([]string{parts[0], parts[1], parts[2], "d3JvbmdzZWNyZXQ="}, ".")

	// Try to validate with wrong secret
	_, err = apiKeyService.ValidateV1ApiKey(invalidKey)
	if err == nil {
		utils.PrintTestError(t, err, "an error for invalid secret")
	}

	expectedMsg := "invalid api key secret"
	if !strings.Contains(err.Error(), expectedMsg) {
		utils.PrintTestError(t, err.Error(), "should contain '"+expectedMsg+"'")
	}
}

func TestApiKeyService_ValidateV1ApiKey_Base64DecodeError(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	apiKeyService := NewApiKeyService(nil)

	// Test with invalid Base64 in secret
	_, err := apiKeyService.ValidateV1ApiKey("key.1.dGVzdC1pZA==.invalid-base64!")
	if err == nil {
		utils.PrintTestError(t, err, "an error for invalid base64")
	}
}

func TestApiKeyService_GetClaimsFromApiKey_Success(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	// Create test user
	user := models.User{
		Username:           "testuser",
		Password:           "hashedpassword",
		DisplayName:        "Test User",
		UserRole:           models.ADMIN,
		DefaultAvatarColor: "#FF0000",
	}
	repositories.GetDB().Create(&user)

	// Set up pepper for HMAC generation
	pepperService := NewPepperService(nil)
	err := pepperService.InitPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	command := commands.UpsertApiKeyCommand{
		Name:        "Claims Test Key",
		Description: "Test claims generation",
		Scope:       "rw",
	}

	apiKeyService := NewApiKeyService(nil)
	generatedKey, err := apiKeyService.CreateApiKey(user.ID, command)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Validate the key to get the database model
	dbApiKey, err := apiKeyService.ValidateV1ApiKey(generatedKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Get claims from the API key
	claims, err := apiKeyService.GetClaimsFromApiKey(dbApiKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify claims are populated correctly
	customClaims, ok := claims.CustomClaims.(*structs.Claims)
	if !ok {
		utils.PrintTestError(t, "type assertion failed", "should be Claims")
	}

	if customClaims.UserId != user.ID {
		utils.PrintTestError(t, customClaims.UserId, user.ID)
	}

	if customClaims.Username != user.Username {
		utils.PrintTestError(t, customClaims.Username, user.Username)
	}

	if customClaims.Displayname != user.DisplayName {
		utils.PrintTestError(t, customClaims.Displayname, user.DisplayName)
	}

	if customClaims.UserRole != user.UserRole {
		utils.PrintTestError(t, customClaims.UserRole, user.UserRole)
	}

	if customClaims.DefaultAvatarColor != user.DefaultAvatarColor {
		utils.PrintTestError(t, customClaims.DefaultAvatarColor, user.DefaultAvatarColor)
	}

	if customClaims.ApiKeyScope != models.ApiKeyScope(command.Scope) {
		utils.PrintTestError(t, customClaims.ApiKeyScope, command.Scope)
	}
}

func TestApiKeyService_GetClaimsFromApiKey_UserNotFound(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	nonExistentUserId := uint(999)
	apiKey := models.ApiKey{
		ID:          "test-id",
		UserID:      &nonExistentUserId,
		Name:        "Orphaned Key",
		Description: "Key without user",
		Scope:       "r",
		Prefix:      "key",
		Hmac:        "test-hmac",
		Version:     1,
	}

	apiKeyService := NewApiKeyService(nil)
	_, err := apiKeyService.GetClaimsFromApiKey(apiKey)

	if err == nil {
		utils.PrintTestError(t, err, "an error for non-existent user")
	}
}

func TestApiKeyService_GetClaimsFromApiKey_AllFieldsPopulated(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	// Create user with all fields populated
	user := models.User{
		Username:           "fulluser",
		Password:           "hashedpassword",
		DisplayName:        "Full Test User",
		UserRole:           models.USER,
		DefaultAvatarColor: "#00FF00",
	}
	repositories.GetDB().Create(&user)

	// Set up pepper
	pepperService := NewPepperService(nil)
	err := pepperService.InitPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Create API key with specific scope
	command := commands.UpsertApiKeyCommand{
		Name:        "Full Fields Test",
		Description: "Test all fields",
		Scope:       "w",
	}

	apiKeyService := NewApiKeyService(nil)
	generatedKey, err := apiKeyService.CreateApiKey(user.ID, command)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	dbApiKey, err := apiKeyService.ValidateV1ApiKey(generatedKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	claims, err := apiKeyService.GetClaimsFromApiKey(dbApiKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	customClaims, ok := claims.CustomClaims.(*structs.Claims)
	if !ok {
		utils.PrintTestError(t, "type assertion failed", "should be Claims")
	}

	// Verify all fields are correctly populated
	if customClaims.UserId != user.ID {
		utils.PrintTestError(t, customClaims.UserId, user.ID)
	}

	if customClaims.Username != "fulluser" {
		utils.PrintTestError(t, customClaims.Username, "fulluser")
	}

	if customClaims.Displayname != "Full Test User" {
		utils.PrintTestError(t, customClaims.Displayname, "Full Test User")
	}

	if customClaims.UserRole != models.USER {
		utils.PrintTestError(t, customClaims.UserRole, models.USER)
	}

	if customClaims.DefaultAvatarColor != "#00FF00" {
		utils.PrintTestError(t, customClaims.DefaultAvatarColor, "#00FF00")
	}

	if customClaims.ApiKeyScope != "w" {
		utils.PrintTestError(t, customClaims.ApiKeyScope, "w")
	}

	// Verify registered claims exist but are empty (as expected)
	if claims.RegisteredClaims.Issuer != "" {
		utils.PrintTestError(t, "RegisteredClaims should be empty", "RegisteredClaims populated unexpectedly")
	}
}

func TestApiKeyService_UpdateApiKeyLastUsedDate_Success(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	// Set up pepper for HMAC generation
	pepperService := NewPepperService(nil)
	err := pepperService.InitPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	userId := uint(1)
	command := commands.UpsertApiKeyCommand{
		Name:        "Last Used Test API Key",
		Description: "Test updating last used date",
		Scope:       "r",
	}

	apiKeyService := NewApiKeyService(nil)

	// Create an API key first
	generatedKey, err := apiKeyService.CreateApiKey(userId, command)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Extract the ID from the generated key
	parts := strings.Split(generatedKey, ".")
	if len(parts) != 4 {
		utils.PrintTestError(t, len(parts), 4)
	}
	apiKeyId := parts[2]

	// Verify initial LastUsedAt is nil
	apiKeyRepo := repositories.NewApiKeyRepository(nil)
	initialApiKey, err := apiKeyRepo.GetApiKeyById(apiKeyId)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if initialApiKey.LastUsedAt != nil {
		utils.PrintTestError(t, initialApiKey.LastUsedAt, nil)
	}

	beforeUpdate := time.Now()

	// Update the last used date using the service
	err = apiKeyService.UpdateApiKeyLastUsedDate(apiKeyId)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	afterUpdate := time.Now()

	// Verify the last used date was updated
	updatedApiKey, err := apiKeyRepo.GetApiKeyById(apiKeyId)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if updatedApiKey.LastUsedAt == nil {
		utils.PrintTestError(t, "LastUsedAt should not be nil", "LastUsedAt should be set")
	}

	if updatedApiKey.LastUsedAt.Before(beforeUpdate) {
		utils.PrintTestError(t, "LastUsedAt is before update started", "LastUsedAt should be after update started")
	}

	if updatedApiKey.LastUsedAt.After(afterUpdate) {
		utils.PrintTestError(t, "LastUsedAt is after update completed", "LastUsedAt should be before update completed")
	}
}

func TestApiKeyService_UpdateApiKeyLastUsedDate_NonExistentKey(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	apiKeyService := NewApiKeyService(nil)

	// Try to update a non-existent API key
	err := apiKeyService.UpdateApiKeyLastUsedDate("non-existent-key-id")

	// Should not return an error (GORM UPDATE on non-existent record doesn't error)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}
}

func TestApiKeyService_UpdateApiKeyLastUsedDate_WithTransaction(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	// Set up pepper for HMAC generation
	pepperService := NewPepperService(nil)
	err := pepperService.InitPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	db := repositories.GetDB()
	tx := db.Begin()
	defer tx.Rollback()

	userId := uint(1)
	command := commands.UpsertApiKeyCommand{
		Name:        "Transaction Last Used Test",
		Description: "Test updating last used date within transaction",
		Scope:       "r",
	}

	// Create API key service with transaction
	apiKeyServiceTx := NewApiKeyService(tx)

	// Create an API key within the transaction
	generatedKey, err := apiKeyServiceTx.CreateApiKey(userId, command)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Extract the ID from the generated key
	parts := strings.Split(generatedKey, ".")
	if len(parts) != 4 {
		utils.PrintTestError(t, len(parts), 4)
	}
	apiKeyId := parts[2]

	// Update last used date within the transaction
	err = apiKeyServiceTx.UpdateApiKeyLastUsedDate(apiKeyId)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify the update within the transaction
	apiKeyRepoTx := repositories.NewApiKeyRepository(tx)
	updatedApiKey, err := apiKeyRepoTx.GetApiKeyById(apiKeyId)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if updatedApiKey.LastUsedAt == nil {
		utils.PrintTestError(t, "LastUsedAt should not be nil", "LastUsedAt should be set")
	}

	// Verify the key is not visible outside the transaction yet
	apiKeyServiceOutside := NewApiKeyService(nil)
	apiKeyRepoOutside := repositories.NewApiKeyRepository(nil)
	_, err = apiKeyRepoOutside.GetApiKeyById(apiKeyId)
	if err == nil {
		utils.PrintTestError(t, err, "an error - key should not be visible outside transaction")
	}

	// Commit the transaction
	tx.Commit()

	// Now verify the update persisted after commit
	persistedApiKey, err := apiKeyRepoOutside.GetApiKeyById(apiKeyId)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if persistedApiKey.LastUsedAt == nil {
		utils.PrintTestError(t, "LastUsedAt should not be nil after commit", "LastUsedAt should be set")
	}

	// Test updating with service outside transaction
	time.Sleep(10 * time.Millisecond) // Ensure timestamp difference
	beforeSecondUpdate := time.Now()

	err = apiKeyServiceOutside.UpdateApiKeyLastUsedDate(apiKeyId)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	secondUpdatedApiKey, err := apiKeyRepoOutside.GetApiKeyById(apiKeyId)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if !secondUpdatedApiKey.LastUsedAt.After(*persistedApiKey.LastUsedAt) {
		utils.PrintTestError(t, "Second update should be after first", "Second LastUsedAt should be more recent")
	}

	if secondUpdatedApiKey.LastUsedAt.Before(beforeSecondUpdate) {
		utils.PrintTestError(t, "Second LastUsedAt is before update", "Second update should be recent")
	}
}

func TestApiKeyService_UpdateApiKey_Success(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	// Set up pepper for HMAC generation
	pepperService := NewPepperService(nil)
	err := pepperService.InitPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	userId := uint(1)
	originalCommand := commands.UpsertApiKeyCommand{
		Name:        "Original API Key",
		Description: "Original description",
		Scope:       "r",
	}

	apiKeyService := NewApiKeyService(nil)

	// Create an API key first
	generatedKey, err := apiKeyService.CreateApiKey(userId, originalCommand)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Extract the ID from the generated key
	parts := strings.Split(generatedKey, ".")
	if len(parts) != 4 {
		utils.PrintTestError(t, len(parts), 4)
	}
	apiKeyId := parts[2]

	// Update the API key
	updateCommand := commands.UpsertApiKeyCommand{
		Name:        "Updated API Key",
		Description: "Updated description",
		Scope:       "rw",
	}

	err = apiKeyService.UpdateApiKey(apiKeyId, userId, updateCommand)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify the API key was updated
	apiKeyRepo := repositories.NewApiKeyRepository(nil)
	updatedApiKey, err := apiKeyRepo.GetApiKeyById(apiKeyId)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if updatedApiKey.Name != updateCommand.Name {
		utils.PrintTestError(t, updatedApiKey.Name, updateCommand.Name)
	}

	if updatedApiKey.Description != updateCommand.Description {
		utils.PrintTestError(t, updatedApiKey.Description, updateCommand.Description)
	}

	if updatedApiKey.Scope != updateCommand.Scope {
		utils.PrintTestError(t, updatedApiKey.Scope, updateCommand.Scope)
	}

	// Verify UserID and other fields remain unchanged
	if *updatedApiKey.UserID != userId {
		utils.PrintTestError(t, *updatedApiKey.UserID, userId)
	}
}

func TestApiKeyService_UpdateApiKey_NotFound(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	userId := uint(1)
	updateCommand := commands.UpsertApiKeyCommand{
		Name:        "Updated API Key",
		Description: "Updated description",
		Scope:       "rw",
	}

	apiKeyService := NewApiKeyService(nil)

	// Try to update a non-existent API key
	err := apiKeyService.UpdateApiKey("non-existent-key-id", userId, updateCommand)

	if err == nil {
		utils.PrintTestError(t, err, "an error")
	}

	expectedMsg := "API key not found"
	if err.Error() != expectedMsg {
		utils.PrintTestError(t, err.Error(), expectedMsg)
	}
}

func TestApiKeyService_UpdateApiKey_WrongUser(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	// Set up pepper for HMAC generation
	pepperService := NewPepperService(nil)
	err := pepperService.InitPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	userA := uint(1)
	userB := uint(2)
	originalCommand := commands.UpsertApiKeyCommand{
		Name:        "User A API Key",
		Description: "User A's key",
		Scope:       "r",
	}

	apiKeyService := NewApiKeyService(nil)

	// Create an API key for user A
	generatedKey, err := apiKeyService.CreateApiKey(userA, originalCommand)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Extract the ID from the generated key
	parts := strings.Split(generatedKey, ".")
	if len(parts) != 4 {
		utils.PrintTestError(t, len(parts), 4)
	}
	apiKeyId := parts[2]

	// Try to update the API key as user B
	updateCommand := commands.UpsertApiKeyCommand{
		Name:        "Malicious Update",
		Description: "User B trying to update User A's key",
		Scope:       "rw",
	}

	err = apiKeyService.UpdateApiKey(apiKeyId, userB, updateCommand)

	if err == nil {
		utils.PrintTestError(t, err, "an error")
	}

	expectedMsg := "API key not found"
	if err.Error() != expectedMsg {
		utils.PrintTestError(t, err.Error(), expectedMsg)
	}

	// Verify the original API key is unchanged
	apiKeyRepo := repositories.NewApiKeyRepository(nil)
	unchangedApiKey, err := apiKeyRepo.GetApiKeyById(apiKeyId)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if unchangedApiKey.Name != originalCommand.Name {
		utils.PrintTestError(t, unchangedApiKey.Name, originalCommand.Name)
	}

	if unchangedApiKey.Description != originalCommand.Description {
		utils.PrintTestError(t, unchangedApiKey.Description, originalCommand.Description)
	}

	if unchangedApiKey.Scope != originalCommand.Scope {
		utils.PrintTestError(t, unchangedApiKey.Scope, originalCommand.Scope)
	}

	if *unchangedApiKey.UserID != userA {
		utils.PrintTestError(t, *unchangedApiKey.UserID, userA)
	}
}

func TestApiKeyService_UpdateApiKey_VerifyOnlyAllowedFieldsUpdate(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	// Set up pepper for HMAC generation
	pepperService := NewPepperService(nil)
	err := pepperService.InitPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	userId := uint(1)
	originalCommand := commands.UpsertApiKeyCommand{
		Name:        "Original API Key",
		Description: "Original description",
		Scope:       "r",
	}

	apiKeyService := NewApiKeyService(nil)

	// Create an API key first
	generatedKey, err := apiKeyService.CreateApiKey(userId, originalCommand)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Extract the ID from the generated key
	parts := strings.Split(generatedKey, ".")
	if len(parts) != 4 {
		utils.PrintTestError(t, len(parts), 4)
	}
	apiKeyId := parts[2]

	// Get the original API key to verify unchanging fields
	apiKeyRepo := repositories.NewApiKeyRepository(nil)
	originalApiKey, err := apiKeyRepo.GetApiKeyById(apiKeyId)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Store original values that should not change
	originalID := originalApiKey.ID
	originalUserID := originalApiKey.UserID
	originalCreatedBy := originalApiKey.CreatedBy
	originalCreatedAt := originalApiKey.CreatedAt
	originalPrefix := originalApiKey.Prefix
	originalHmac := originalApiKey.Hmac
	originalVersion := originalApiKey.Version

	// Update the API key
	updateCommand := commands.UpsertApiKeyCommand{
		Name:        "Updated API Key",
		Description: "Updated description",
		Scope:       "rw",
	}

	err = apiKeyService.UpdateApiKey(apiKeyId, userId, updateCommand)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify the API key was updated
	updatedApiKey, err := apiKeyRepo.GetApiKeyById(apiKeyId)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify allowed fields were updated
	if updatedApiKey.Name != updateCommand.Name {
		utils.PrintTestError(t, updatedApiKey.Name, updateCommand.Name)
	}

	if updatedApiKey.Description != updateCommand.Description {
		utils.PrintTestError(t, updatedApiKey.Description, updateCommand.Description)
	}

	if updatedApiKey.Scope != updateCommand.Scope {
		utils.PrintTestError(t, updatedApiKey.Scope, updateCommand.Scope)
	}

	// Verify fields that should NOT change remained the same
	if updatedApiKey.ID != originalID {
		utils.PrintTestError(t, updatedApiKey.ID, originalID)
	}

	if updatedApiKey.UserID == nil || *updatedApiKey.UserID != *originalUserID {
		utils.PrintTestError(t, *updatedApiKey.UserID, *originalUserID)
	}

	if updatedApiKey.CreatedBy == nil || *updatedApiKey.CreatedBy != *originalCreatedBy {
		utils.PrintTestError(t, *updatedApiKey.CreatedBy, *originalCreatedBy)
	}

	if !updatedApiKey.CreatedAt.Equal(originalCreatedAt) {
		utils.PrintTestError(t, updatedApiKey.CreatedAt, originalCreatedAt)
	}

	if updatedApiKey.Prefix != originalPrefix {
		utils.PrintTestError(t, updatedApiKey.Prefix, originalPrefix)
	}

	if updatedApiKey.Hmac != originalHmac {
		utils.PrintTestError(t, updatedApiKey.Hmac, originalHmac)
	}

	if updatedApiKey.Version != originalVersion {
		utils.PrintTestError(t, updatedApiKey.Version, originalVersion)
	}

	// Verify UpdatedAt field was actually updated
	if !updatedApiKey.UpdatedAt.After(originalApiKey.UpdatedAt) {
		utils.PrintTestError(t, "UpdatedAt should be after original", "UpdatedAt should have been updated")
	}
}

func TestApiKeyService_UpdateApiKey_WithTransaction(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	// Set up pepper for HMAC generation
	pepperService := NewPepperService(nil)
	err := pepperService.InitPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	db := repositories.GetDB()
	tx := db.Begin()
	defer tx.Rollback()

	userId := uint(1)
	originalCommand := commands.UpsertApiKeyCommand{
		Name:        "Transaction Test Key",
		Description: "Test update within transaction",
		Scope:       "r",
	}

	// Create API key service with transaction
	apiKeyServiceTx := NewApiKeyService(tx)

	// Create an API key within the transaction
	generatedKey, err := apiKeyServiceTx.CreateApiKey(userId, originalCommand)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Extract the ID from the generated key
	parts := strings.Split(generatedKey, ".")
	if len(parts) != 4 {
		utils.PrintTestError(t, len(parts), 4)
	}
	apiKeyId := parts[2]

	// Update the API key within the transaction
	updateCommand := commands.UpsertApiKeyCommand{
		Name:        "Updated in Transaction",
		Description: "Updated description in transaction",
		Scope:       "rw",
	}

	err = apiKeyServiceTx.UpdateApiKey(apiKeyId, userId, updateCommand)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify the update within the transaction
	apiKeyRepoTx := repositories.NewApiKeyRepository(tx)
	updatedApiKey, err := apiKeyRepoTx.GetApiKeyById(apiKeyId)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if updatedApiKey.Name != updateCommand.Name {
		utils.PrintTestError(t, updatedApiKey.Name, updateCommand.Name)
	}

	if updatedApiKey.Description != updateCommand.Description {
		utils.PrintTestError(t, updatedApiKey.Description, updateCommand.Description)
	}

	if updatedApiKey.Scope != updateCommand.Scope {
		utils.PrintTestError(t, updatedApiKey.Scope, updateCommand.Scope)
	}

	// Verify the key is not visible outside the transaction yet
	apiKeyRepoOutside := repositories.NewApiKeyRepository(nil)
	_, err = apiKeyRepoOutside.GetApiKeyById(apiKeyId)
	if err == nil {
		utils.PrintTestError(t, err, "an error - key should not be visible outside transaction")
	}

	// Commit the transaction
	tx.Commit()

	// Now verify the update persisted after commit
	persistedApiKey, err := apiKeyRepoOutside.GetApiKeyById(apiKeyId)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if persistedApiKey.Name != updateCommand.Name {
		utils.PrintTestError(t, persistedApiKey.Name, updateCommand.Name)
	}

	if persistedApiKey.Description != updateCommand.Description {
		utils.PrintTestError(t, persistedApiKey.Description, updateCommand.Description)
	}

	if persistedApiKey.Scope != updateCommand.Scope {
		utils.PrintTestError(t, persistedApiKey.Scope, updateCommand.Scope)
	}
}

func TestApiKeyService_UpdateApiKey_MultipleUpdates(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	// Set up pepper for HMAC generation
	pepperService := NewPepperService(nil)
	err := pepperService.InitPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	userId := uint(1)
	originalCommand := commands.UpsertApiKeyCommand{
		Name:        "Multiple Updates Test",
		Description: "Original description",
		Scope:       "r",
	}

	apiKeyService := NewApiKeyService(nil)

	// Create an API key first
	generatedKey, err := apiKeyService.CreateApiKey(userId, originalCommand)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Extract the ID from the generated key
	parts := strings.Split(generatedKey, ".")
	if len(parts) != 4 {
		utils.PrintTestError(t, len(parts), 4)
	}
	apiKeyId := parts[2]

	// First update
	firstUpdateCommand := commands.UpsertApiKeyCommand{
		Name:        "First Update",
		Description: "First updated description",
		Scope:       "w",
	}

	err = apiKeyService.UpdateApiKey(apiKeyId, userId, firstUpdateCommand)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify first update
	apiKeyRepo := repositories.NewApiKeyRepository(nil)
	firstUpdatedApiKey, err := apiKeyRepo.GetApiKeyById(apiKeyId)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if firstUpdatedApiKey.Name != firstUpdateCommand.Name {
		utils.PrintTestError(t, firstUpdatedApiKey.Name, firstUpdateCommand.Name)
	}

	if firstUpdatedApiKey.Description != firstUpdateCommand.Description {
		utils.PrintTestError(t, firstUpdatedApiKey.Description, firstUpdateCommand.Description)
	}

	if firstUpdatedApiKey.Scope != firstUpdateCommand.Scope {
		utils.PrintTestError(t, firstUpdatedApiKey.Scope, firstUpdateCommand.Scope)
	}

	// Store first update time for comparison
	firstUpdateTime := firstUpdatedApiKey.UpdatedAt

	// Sleep briefly to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Second update
	secondUpdateCommand := commands.UpsertApiKeyCommand{
		Name:        "Second Update",
		Description: "Second updated description",
		Scope:       "rw",
	}

	err = apiKeyService.UpdateApiKey(apiKeyId, userId, secondUpdateCommand)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify second update
	secondUpdatedApiKey, err := apiKeyRepo.GetApiKeyById(apiKeyId)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if secondUpdatedApiKey.Name != secondUpdateCommand.Name {
		utils.PrintTestError(t, secondUpdatedApiKey.Name, secondUpdateCommand.Name)
	}

	if secondUpdatedApiKey.Description != secondUpdateCommand.Description {
		utils.PrintTestError(t, secondUpdatedApiKey.Description, secondUpdateCommand.Description)
	}

	if secondUpdatedApiKey.Scope != secondUpdateCommand.Scope {
		utils.PrintTestError(t, secondUpdatedApiKey.Scope, secondUpdateCommand.Scope)
	}

	// Verify the UpdatedAt timestamp was updated again
	if !secondUpdatedApiKey.UpdatedAt.After(firstUpdateTime) {
		utils.PrintTestError(t, "Second update should be after first", "Second UpdatedAt should be more recent")
	}

	// Verify unchanged fields remain the same
	if secondUpdatedApiKey.ID != firstUpdatedApiKey.ID {
		utils.PrintTestError(t, secondUpdatedApiKey.ID, firstUpdatedApiKey.ID)
	}

	if *secondUpdatedApiKey.UserID != *firstUpdatedApiKey.UserID {
		utils.PrintTestError(t, *secondUpdatedApiKey.UserID, *firstUpdatedApiKey.UserID)
	}

	if secondUpdatedApiKey.Hmac != firstUpdatedApiKey.Hmac {
		utils.PrintTestError(t, secondUpdatedApiKey.Hmac, firstUpdatedApiKey.Hmac)
	}

	if secondUpdatedApiKey.Version != firstUpdatedApiKey.Version {
		utils.PrintTestError(t, secondUpdatedApiKey.Version, firstUpdatedApiKey.Version)
	}
}

func TestApiKeyService_UpdateApiKey_DatabaseError(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	// Set up pepper for HMAC generation
	pepperService := NewPepperService(nil)
	err := pepperService.InitPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	userId := uint(1)
	originalCommand := commands.UpsertApiKeyCommand{
		Name:        "Database Error Test",
		Description: "Test database error handling",
		Scope:       "r",
	}

	apiKeyService := NewApiKeyService(nil)

	// Create an API key first
	generatedKey, err := apiKeyService.CreateApiKey(userId, originalCommand)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Extract the ID from the generated key
	parts := strings.Split(generatedKey, ".")
	if len(parts) != 4 {
		utils.PrintTestError(t, len(parts), 4)
	}
	apiKeyId := parts[2]

	// Manually corrupt the API key in the database to have a nil UserID
	// This simulates a data integrity issue that would cause the ownership check to fail
	db := repositories.GetDB()
	err = db.Model(&models.ApiKey{}).Where("id = ?", apiKeyId).Update("user_id", nil).Error
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Try to update the API key - should fail due to nil UserID
	updateCommand := commands.UpsertApiKeyCommand{
		Name:        "Should Not Update",
		Description: "This should fail",
		Scope:       "rw",
	}

	err = apiKeyService.UpdateApiKey(apiKeyId, userId, updateCommand)

	if err == nil {
		utils.PrintTestError(t, err, "an error")
	}

	expectedMsg := "API key not found"
	if err.Error() != expectedMsg {
		utils.PrintTestError(t, err.Error(), expectedMsg)
	}

	// Verify the API key was not updated
	apiKeyRepo := repositories.NewApiKeyRepository(nil)
	unchangedApiKey, err := apiKeyRepo.GetApiKeyById(apiKeyId)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Should still have the original values (except UserID which is now nil)
	if unchangedApiKey.Name != originalCommand.Name {
		utils.PrintTestError(t, unchangedApiKey.Name, originalCommand.Name)
	}

	if unchangedApiKey.Description != originalCommand.Description {
		utils.PrintTestError(t, unchangedApiKey.Description, originalCommand.Description)
	}

	if unchangedApiKey.Scope != originalCommand.Scope {
		utils.PrintTestError(t, unchangedApiKey.Scope, originalCommand.Scope)
	}

	// Verify UserID is indeed nil (as we corrupted it)
	if unchangedApiKey.UserID != nil {
		utils.PrintTestError(t, unchangedApiKey.UserID, nil)
	}
}
