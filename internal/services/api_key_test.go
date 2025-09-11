package services

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"
	"strings"
	"testing"
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
		Scope:       "read",
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
}

func TestApiKeyService_CreateApiKey_HmacGenerationError(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "")
	defer repositories.TruncateTestDb()

	userId := uint(1)
	command := commands.UpsertApiKeyCommand{
		Name:        "Test API Key",
		Description: "Test description",
		Scope:       "read",
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
		Scope:       "write",
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
	secret := parts[3]

	// Generate HMAC for the same secret and compare
	expectedHmac, err := apiKeyService.GenerateApiKeyHmac(secret)
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
		Scope:       "read",
	}

	key1, err := apiKeyService.CreateApiKey(userId, command1)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Create second API key
	command2 := commands.UpsertApiKeyCommand{
		Name:        "Second Key",
		Description: "Second test key",
		Scope:       "write",
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
