package repositories

import (
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"testing"
	"time"
)

func TestNewApiKeyRepository(t *testing.T) {
	repository := NewApiKeyRepository(nil)

	if repository.DB == nil {
		utils.PrintTestError(t, repository.DB, "a database instance")
	}

	if repository.TX != nil {
		utils.PrintTestError(t, repository.TX, "nil")
	}
}

func TestNewApiKeyRepositoryWithTransaction(t *testing.T) {
	db := GetDB()
	tx := db.Begin()
	defer tx.Rollback()

	repository := NewApiKeyRepository(tx)

	if repository.DB == nil {
		utils.PrintTestError(t, repository.DB, "a database instance")
	}

	if repository.TX == nil {
		utils.PrintTestError(t, repository.TX, "a transaction instance")
	}
}

func TestCreateApiKey_Success(t *testing.T) {
	defer TruncateTestDb()

	userId := uint(1)
	apiKey := models.ApiKey{
		ID:          "test-key-id",
		UserID:      &userId,
		Name:        "Test API Key",
		Description: "Test description",
		Prefix:      "key",
		Hmac:        "test-hmac",
		Version:     1,
		Scope:       "read",
	}

	repository := NewApiKeyRepository(nil)
	createdApiKey, err := repository.CreateApiKey(apiKey)

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if createdApiKey.ID != apiKey.ID {
		utils.PrintTestError(t, createdApiKey.ID, apiKey.ID)
	}

	if createdApiKey.Name != apiKey.Name {
		utils.PrintTestError(t, createdApiKey.Name, apiKey.Name)
	}

	if createdApiKey.Description != apiKey.Description {
		utils.PrintTestError(t, createdApiKey.Description, apiKey.Description)
	}

	if *createdApiKey.UserID != *apiKey.UserID {
		utils.PrintTestError(t, *createdApiKey.UserID, *apiKey.UserID)
	}

	if createdApiKey.Prefix != apiKey.Prefix {
		utils.PrintTestError(t, createdApiKey.Prefix, apiKey.Prefix)
	}

	if createdApiKey.Hmac != apiKey.Hmac {
		utils.PrintTestError(t, createdApiKey.Hmac, apiKey.Hmac)
	}

	if createdApiKey.Version != apiKey.Version {
		utils.PrintTestError(t, createdApiKey.Version, apiKey.Version)
	}

	if createdApiKey.Scope != apiKey.Scope {
		utils.PrintTestError(t, createdApiKey.Scope, apiKey.Scope)
	}

	// Verify the API key was saved to the database
	var savedApiKey models.ApiKey
	err = GetDB().Where("id = ?", apiKey.ID).First(&savedApiKey).Error
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if savedApiKey.ID != apiKey.ID {
		utils.PrintTestError(t, savedApiKey.ID, apiKey.ID)
	}

	if savedApiKey.Name != apiKey.Name {
		utils.PrintTestError(t, savedApiKey.Name, apiKey.Name)
	}
}

func TestCreateApiKey_AllFields(t *testing.T) {
	defer TruncateTestDb()

	userId := uint(1)
	createdBy := uint(2)
	lastUsedAt := time.Now()
	revokedAt := time.Now().Add(time.Hour)

	apiKey := models.ApiKey{
		ID:              "full-test-key-id",
		UserID:          &userId,
		CreatedBy:       &createdBy,
		CreatedByString: "test-creator",
		Name:            "Full Test API Key",
		Description:     "Complete test with all fields",
		Prefix:          "key",
		Hmac:            "full-test-hmac",
		Version:         2,
		Scope:           "read,write",
		LastUsedAt:      &lastUsedAt,
		RevokedAt:       &revokedAt,
	}

	repository := NewApiKeyRepository(nil)
	createdApiKey, err := repository.CreateApiKey(apiKey)

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if createdApiKey.ID != apiKey.ID {
		utils.PrintTestError(t, createdApiKey.ID, apiKey.ID)
	}

	if *createdApiKey.CreatedBy != *apiKey.CreatedBy {
		utils.PrintTestError(t, *createdApiKey.CreatedBy, *apiKey.CreatedBy)
	}

	if createdApiKey.CreatedByString != apiKey.CreatedByString {
		utils.PrintTestError(t, createdApiKey.CreatedByString, apiKey.CreatedByString)
	}

	if createdApiKey.LastUsedAt.IsZero() {
		utils.PrintTestError(t, "LastUsedAt should not be zero", "LastUsedAt should be set")
	}

	if createdApiKey.RevokedAt.IsZero() {
		utils.PrintTestError(t, "RevokedAt should not be zero", "RevokedAt should be set")
	}

	// Verify in database
	var savedApiKey models.ApiKey
	err = GetDB().Where("id = ?", apiKey.ID).First(&savedApiKey).Error
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if savedApiKey.Version != 2 {
		utils.PrintTestError(t, savedApiKey.Version, 2)
	}

	if savedApiKey.Scope != "read,write" {
		utils.PrintTestError(t, savedApiKey.Scope, "read,write")
	}
}

func TestCreateApiKey_MinimalFields(t *testing.T) {
	defer TruncateTestDb()

	apiKey := models.ApiKey{
		ID:      "minimal-key-id",
		Name:    "Minimal Key",
		Prefix:  "key",
		Hmac:    "minimal-hmac",
		Version: 1,
		Scope:   "read",
	}

	repository := NewApiKeyRepository(nil)
	createdApiKey, err := repository.CreateApiKey(apiKey)

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if createdApiKey.ID != apiKey.ID {
		utils.PrintTestError(t, createdApiKey.ID, apiKey.ID)
	}

	if createdApiKey.UserID != nil {
		utils.PrintTestError(t, createdApiKey.UserID, nil)
	}

	if createdApiKey.CreatedBy != nil {
		utils.PrintTestError(t, createdApiKey.CreatedBy, nil)
	}

	if createdApiKey.LastUsedAt != nil {
		utils.PrintTestError(t, createdApiKey.LastUsedAt, nil)
	}

	if createdApiKey.RevokedAt != nil {
		utils.PrintTestError(t, createdApiKey.RevokedAt, nil)
	}
}

func TestCreateApiKey_DuplicateID(t *testing.T) {
	defer TruncateTestDb()

	userId := uint(1)
	apiKey1 := models.ApiKey{
		ID:          "duplicate-id",
		UserID:      &userId,
		Name:        "First Key",
		Description: "First key with duplicate ID",
		Prefix:      "key",
		Hmac:        "first-hmac",
		Version:     1,
		Scope:       "read",
	}

	apiKey2 := models.ApiKey{
		ID:          "duplicate-id",
		UserID:      &userId,
		Name:        "Second Key",
		Description: "Second key with duplicate ID",
		Prefix:      "key",
		Hmac:        "second-hmac",
		Version:     1,
		Scope:       "write",
	}

	repository := NewApiKeyRepository(nil)

	// Create first API key
	_, err := repository.CreateApiKey(apiKey1)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Attempt to create second API key with same ID
	_, err = repository.CreateApiKey(apiKey2)
	if err == nil {
		utils.PrintTestError(t, err, "an error for duplicate ID")
	}

	// Verify only the first key exists
	var count int64
	GetDB().Model(&models.ApiKey{}).Where("id = ?", "duplicate-id").Count(&count)
	if count != 1 {
		utils.PrintTestError(t, count, 1)
	}

	// Verify it's the first key that was saved
	var savedApiKey models.ApiKey
	GetDB().Where("id = ?", "duplicate-id").First(&savedApiKey)
	if savedApiKey.Name != "First Key" {
		utils.PrintTestError(t, savedApiKey.Name, "First Key")
	}
}

func TestCreateApiKey_WithTransaction(t *testing.T) {
	defer TruncateTestDb()

	db := GetDB()
	tx := db.Begin()

	userId := uint(1)
	apiKey := models.ApiKey{
		ID:          "transaction-key-id",
		UserID:      &userId,
		Name:        "Transaction Test Key",
		Description: "Key created within transaction",
		Prefix:      "key",
		Hmac:        "transaction-hmac",
		Version:     1,
		Scope:       "read",
	}

	repository := NewApiKeyRepository(tx)
	createdApiKey, err := repository.CreateApiKey(apiKey)

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if createdApiKey.ID != apiKey.ID {
		utils.PrintTestError(t, createdApiKey.ID, apiKey.ID)
	}

	// Before commit, key should not be visible outside transaction
	var count int64
	db.Model(&models.ApiKey{}).Where("id = ?", apiKey.ID).Count(&count)
	if count != 0 {
		utils.PrintTestError(t, count, 0)
	}

	// Commit transaction
	tx.Commit()

	// After commit, key should be visible
	db.Model(&models.ApiKey{}).Where("id = ?", apiKey.ID).Count(&count)
	if count != 1 {
		utils.PrintTestError(t, count, 1)
	}
}

func TestCreateApiKey_TransactionRollback(t *testing.T) {
	defer TruncateTestDb()

	db := GetDB()
	tx := db.Begin()

	userId := uint(1)
	apiKey := models.ApiKey{
		ID:          "rollback-key-id",
		UserID:      &userId,
		Name:        "Rollback Test Key",
		Description: "Key to be rolled back",
		Prefix:      "key",
		Hmac:        "rollback-hmac",
		Version:     1,
		Scope:       "read",
	}

	repository := NewApiKeyRepository(tx)
	createdApiKey, err := repository.CreateApiKey(apiKey)

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if createdApiKey.ID != apiKey.ID {
		utils.PrintTestError(t, createdApiKey.ID, apiKey.ID)
	}

	// Rollback transaction
	tx.Rollback()

	// After rollback, key should not exist
	var count int64
	db.Model(&models.ApiKey{}).Where("id = ?", apiKey.ID).Count(&count)
	if count != 0 {
		utils.PrintTestError(t, count, 0)
	}
}

func TestCreateApiKey_MultipleKeys(t *testing.T) {
	defer TruncateTestDb()

	userId1 := uint(1)
	userId2 := uint(2)

	apiKey1 := models.ApiKey{
		ID:          "multi-key-1",
		UserID:      &userId1,
		Name:        "First Multi Key",
		Description: "First of multiple keys",
		Prefix:      "key",
		Hmac:        "multi-hmac-1",
		Version:     1,
		Scope:       "read",
	}

	apiKey2 := models.ApiKey{
		ID:          "multi-key-2",
		UserID:      &userId2,
		Name:        "Second Multi Key",
		Description: "Second of multiple keys",
		Prefix:      "key",
		Hmac:        "multi-hmac-2",
		Version:     1,
		Scope:       "write",
	}

	apiKey3 := models.ApiKey{
		ID:          "multi-key-3",
		UserID:      &userId1,
		Name:        "Third Multi Key",
		Description: "Third of multiple keys",
		Prefix:      "key",
		Hmac:        "multi-hmac-3",
		Version:     2,
		Scope:       "admin",
	}

	repository := NewApiKeyRepository(nil)

	// Create all three keys
	_, err := repository.CreateApiKey(apiKey1)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	_, err = repository.CreateApiKey(apiKey2)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	_, err = repository.CreateApiKey(apiKey3)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify all keys exist
	var totalCount int64
	GetDB().Model(&models.ApiKey{}).Count(&totalCount)
	if totalCount != 3 {
		utils.PrintTestError(t, totalCount, 3)
	}

	// Verify keys for user 1
	var user1Count int64
	GetDB().Model(&models.ApiKey{}).Where("user_id = ?", userId1).Count(&user1Count)
	if user1Count != 2 {
		utils.PrintTestError(t, user1Count, 2)
	}

	// Verify keys for user 2
	var user2Count int64
	GetDB().Model(&models.ApiKey{}).Where("user_id = ?", userId2).Count(&user2Count)
	if user2Count != 1 {
		utils.PrintTestError(t, user2Count, 1)
	}

	// Verify different scopes
	var readCount int64
	GetDB().Model(&models.ApiKey{}).Where("scope = ?", "read").Count(&readCount)
	if readCount != 1 {
		utils.PrintTestError(t, readCount, 1)
	}

	var writeCount int64
	GetDB().Model(&models.ApiKey{}).Where("scope = ?", "write").Count(&writeCount)
	if writeCount != 1 {
		utils.PrintTestError(t, writeCount, 1)
	}

	var adminCount int64
	GetDB().Model(&models.ApiKey{}).Where("scope = ?", "admin").Count(&adminCount)
	if adminCount != 1 {
		utils.PrintTestError(t, adminCount, 1)
	}
}

func TestCreateApiKey_VerifyTimestamps(t *testing.T) {
	defer TruncateTestDb()

	beforeCreate := time.Now()

	userId := uint(1)
	apiKey := models.ApiKey{
		ID:          "timestamp-key-id",
		UserID:      &userId,
		Name:        "Timestamp Test Key",
		Description: "Key to test timestamp behavior",
		Prefix:      "key",
		Hmac:        "timestamp-hmac",
		Version:     1,
		Scope:       "read",
	}

	repository := NewApiKeyRepository(nil)
	createdApiKey, err := repository.CreateApiKey(apiKey)

	afterCreate := time.Now()

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify CreatedAt is set and within reasonable bounds
	if createdApiKey.CreatedAt.IsZero() {
		utils.PrintTestError(t, "CreatedAt should not be zero", "CreatedAt should be set")
	}

	if createdApiKey.CreatedAt.Before(beforeCreate) {
		utils.PrintTestError(t, "CreatedAt is before creation started", "CreatedAt should be after creation started")
	}

	if createdApiKey.CreatedAt.After(afterCreate) {
		utils.PrintTestError(t, "CreatedAt is after creation completed", "CreatedAt should be before creation completed")
	}

	// Verify UpdatedAt is set and within reasonable bounds
	if createdApiKey.UpdatedAt.IsZero() {
		utils.PrintTestError(t, "UpdatedAt should not be zero", "UpdatedAt should be set")
	}

	if createdApiKey.UpdatedAt.Before(beforeCreate) {
		utils.PrintTestError(t, "UpdatedAt is before creation started", "UpdatedAt should be after creation started")
	}

	if createdApiKey.UpdatedAt.After(afterCreate) {
		utils.PrintTestError(t, "UpdatedAt is after creation completed", "UpdatedAt should be before creation completed")
	}

	// Verify timestamps in database match returned values
	var savedApiKey models.ApiKey
	GetDB().Where("id = ?", apiKey.ID).First(&savedApiKey)

	if !savedApiKey.CreatedAt.Equal(createdApiKey.CreatedAt) {
		utils.PrintTestError(t, savedApiKey.CreatedAt, createdApiKey.CreatedAt)
	}

	if !savedApiKey.UpdatedAt.Equal(createdApiKey.UpdatedAt) {
		utils.PrintTestError(t, savedApiKey.UpdatedAt, createdApiKey.UpdatedAt)
	}
}

func TestGetApiKeyById_Success(t *testing.T) {
	defer TruncateTestDb()

	userId := uint(1)
	apiKey := models.ApiKey{
		ID:          "test-get-key-id",
		UserID:      &userId,
		Name:        "Get Test API Key",
		Description: "Test retrieving API key",
		Prefix:      "key",
		Hmac:        "test-get-hmac",
		Version:     1,
		Scope:       "read",
	}

	repository := NewApiKeyRepository(nil)

	// Create the API key first
	_, err := repository.CreateApiKey(apiKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Retrieve the API key by ID
	retrievedApiKey, err := repository.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify all fields match
	if retrievedApiKey.ID != apiKey.ID {
		utils.PrintTestError(t, retrievedApiKey.ID, apiKey.ID)
	}

	if retrievedApiKey.Name != apiKey.Name {
		utils.PrintTestError(t, retrievedApiKey.Name, apiKey.Name)
	}

	if retrievedApiKey.Description != apiKey.Description {
		utils.PrintTestError(t, retrievedApiKey.Description, apiKey.Description)
	}

	if *retrievedApiKey.UserID != *apiKey.UserID {
		utils.PrintTestError(t, *retrievedApiKey.UserID, *apiKey.UserID)
	}

	if retrievedApiKey.Prefix != apiKey.Prefix {
		utils.PrintTestError(t, retrievedApiKey.Prefix, apiKey.Prefix)
	}

	if retrievedApiKey.Hmac != apiKey.Hmac {
		utils.PrintTestError(t, retrievedApiKey.Hmac, apiKey.Hmac)
	}

	if retrievedApiKey.Version != apiKey.Version {
		utils.PrintTestError(t, retrievedApiKey.Version, apiKey.Version)
	}

	if retrievedApiKey.Scope != apiKey.Scope {
		utils.PrintTestError(t, retrievedApiKey.Scope, apiKey.Scope)
	}

	// Verify timestamps are set
	if retrievedApiKey.CreatedAt.IsZero() {
		utils.PrintTestError(t, "CreatedAt should not be zero", "CreatedAt should be set")
	}

	if retrievedApiKey.UpdatedAt.IsZero() {
		utils.PrintTestError(t, "UpdatedAt should not be zero", "UpdatedAt should be set")
	}
}

func TestGetApiKeyById_NotFound(t *testing.T) {
	defer TruncateTestDb()

	repository := NewApiKeyRepository(nil)

	// Try to retrieve non-existent API key
	_, err := repository.GetApiKeyById("non-existent-id")
	if err == nil {
		utils.PrintTestError(t, err, "an error for non-existent API key")
	}
}

func TestGetApiKeyById_EmptyId(t *testing.T) {
	defer TruncateTestDb()

	repository := NewApiKeyRepository(nil)

	// Try to retrieve with empty ID
	_, err := repository.GetApiKeyById("")
	if err == nil {
		utils.PrintTestError(t, err, "an error for empty ID")
	}
}

func TestGetApiKeyById_WithTransaction(t *testing.T) {
	defer TruncateTestDb()

	db := GetDB()
	tx := db.Begin()
	defer tx.Rollback()

	userId := uint(1)
	apiKey := models.ApiKey{
		ID:          "transaction-get-key-id",
		UserID:      &userId,
		Name:        "Transaction Get Test",
		Description: "Test retrieving with transaction",
		Prefix:      "key",
		Hmac:        "transaction-get-hmac",
		Version:     1,
		Scope:       "read",
	}

	repository := NewApiKeyRepository(tx)

	// Create API key within transaction
	_, err := repository.CreateApiKey(apiKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Retrieve API key within same transaction
	retrievedApiKey, err := repository.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if retrievedApiKey.ID != apiKey.ID {
		utils.PrintTestError(t, retrievedApiKey.ID, apiKey.ID)
	}

	if retrievedApiKey.Name != apiKey.Name {
		utils.PrintTestError(t, retrievedApiKey.Name, apiKey.Name)
	}

	// Verify key is not visible outside transaction
	repositoryOutside := NewApiKeyRepository(nil)
	_, err = repositoryOutside.GetApiKeyById(apiKey.ID)
	if err == nil {
		utils.PrintTestError(t, err, "an error - key should not be visible outside transaction")
	}
}

func TestGetApiKeyById_AfterUpdate(t *testing.T) {
	defer TruncateTestDb()

	userId := uint(1)
	apiKey := models.ApiKey{
		ID:          "update-get-key-id",
		UserID:      &userId,
		Name:        "Original Name",
		Description: "Original description",
		Prefix:      "key",
		Hmac:        "original-hmac",
		Version:     1,
		Scope:       "read",
	}

	repository := NewApiKeyRepository(nil)

	// Create API key
	_, err := repository.CreateApiKey(apiKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Update the API key directly in database
	GetDB().Model(&models.ApiKey{}).Where("id = ?", apiKey.ID).Updates(map[string]interface{}{
		"name":        "Updated Name",
		"description": "Updated description",
		"scope":       "write",
	})

	// Retrieve updated API key
	retrievedApiKey, err := repository.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify updated values
	if retrievedApiKey.Name != "Updated Name" {
		utils.PrintTestError(t, retrievedApiKey.Name, "Updated Name")
	}

	if retrievedApiKey.Description != "Updated description" {
		utils.PrintTestError(t, retrievedApiKey.Description, "Updated description")
	}

	if retrievedApiKey.Scope != "write" {
		utils.PrintTestError(t, retrievedApiKey.Scope, "write")
	}

	// Verify unchanged values
	if retrievedApiKey.ID != apiKey.ID {
		utils.PrintTestError(t, retrievedApiKey.ID, apiKey.ID)
	}

	if retrievedApiKey.Prefix != apiKey.Prefix {
		utils.PrintTestError(t, retrievedApiKey.Prefix, apiKey.Prefix)
	}

	if retrievedApiKey.Version != apiKey.Version {
		utils.PrintTestError(t, retrievedApiKey.Version, apiKey.Version)
	}
}
