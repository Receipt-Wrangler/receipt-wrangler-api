package repositories

import (
	"fmt"
	"receipt-wrangler/api/internal/commands"
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
		CreatedBy:   &userId,
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
}

func TestCreateApiKey_DuplicateID(t *testing.T) {
	defer TruncateTestDb()

	userId := uint(1)
	apiKey1 := models.ApiKey{
		ID:          "duplicate-id",
		UserID:      &userId,
		CreatedBy:   &userId,
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
		CreatedBy:   &userId,
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
		CreatedBy:   &userId,
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
		CreatedBy:   &userId,
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
		CreatedBy:   &userId1,
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
		CreatedBy:   &userId2,
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
		CreatedBy:   &userId1,
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
		CreatedBy:   &userId,
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

// Helper function to create test API keys for pagination tests
func createTestApiKeysForPagination() {
	user1 := uint(1)
	user2 := uint(2)
	now := time.Now()
	pastTime := now.Add(-time.Hour)

	apiKeys := []models.ApiKey{
		{
			ID:          "key-1",
			UserID:      &user1,
			CreatedBy:   &user1,
			Name:        "Alpha Key",
			Description: "First key alphabetically",
			Prefix:      "key",
			Hmac:        "hmac-1",
			Version:     1,
			Scope:       "read",
			CreatedAt:   pastTime,
			UpdatedAt:   pastTime,
		},
		{
			ID:          "key-2",
			UserID:      &user1,
			CreatedBy:   &user1,
			Name:        "Beta Key",
			Description: "Second key alphabetically",
			Prefix:      "key",
			Hmac:        "hmac-2",
			Version:     2,
			Scope:       "write",
			CreatedAt:   now,
			UpdatedAt:   now,
			LastUsedAt:  &now,
		},
		{
			ID:          "key-3",
			UserID:      &user2,
			CreatedBy:   &user2,
			Name:        "Gamma Key",
			Description: "Third key alphabetically",
			Prefix:      "key",
			Hmac:        "hmac-3",
			Version:     1,
			Scope:       "admin",
			CreatedAt:   pastTime.Add(-time.Hour),
			UpdatedAt:   pastTime.Add(-time.Hour),
		},
		{
			ID:          "key-4",
			UserID:      &user2,
			CreatedBy:   &user2,
			Name:        "Delta Key",
			Description: "Fourth key alphabetically",
			Prefix:      "key",
			Hmac:        "hmac-4",
			Version:     1,
			Scope:       "read",
			CreatedAt:   now.Add(-30 * time.Minute),
			UpdatedAt:   now.Add(-30 * time.Minute),
		},
		{
			ID:          "key-5",
			UserID:      &user1,
			CreatedBy:   &user1,
			Name:        "Echo Key",
			Description: "Fifth key alphabetically",
			Prefix:      "key",
			Hmac:        "hmac-5",
			Version:     1,
			Scope:       "read",
			CreatedAt:   now.Add(-15 * time.Minute),
			UpdatedAt:   now.Add(-15 * time.Minute),
		},
	}

	db := GetDB()
	for _, apiKey := range apiKeys {
		db.Create(&apiKey)
	}
}

func TestGetPagedApiKeys_BasicPagination(t *testing.T) {
	defer TruncateTestDb()
	createTestApiKeysForPagination()

	repository := NewApiKeyRepository(nil)
	command := commands.PagedApiKeyRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:          1,
			PageSize:      2,
			OrderBy:       "name",
			SortDirection: commands.ASCENDING,
		},
		ApiKeyFilter: commands.ApiKeyFilter{
			AssociatedApiKeys: commands.ASSOCIATED_API_KEYS_ALL,
		},
	}

	results, count, err := repository.GetPagedApiKeys(command, "1")

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Should return 2 results (page size)
	if len(results) != 2 {
		utils.PrintTestError(t, len(results), 2)
	}

	// Count should be 5
	if count != 5 {
		utils.PrintTestError(t, count, int64(5))
	}

	// Results should be sorted by name ascending
	if results[0].Name != "Alpha Key" {
		utils.PrintTestError(t, results[0].Name, "Alpha Key")
	}

	if results[1].Name != "Beta Key" {
		utils.PrintTestError(t, results[1].Name, "Beta Key")
	}
}

func TestGetPagedApiKeys_SecondPage(t *testing.T) {
	defer TruncateTestDb()
	createTestApiKeysForPagination()

	repository := NewApiKeyRepository(nil)
	command := commands.PagedApiKeyRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:          2,
			PageSize:      2,
			OrderBy:       "name",
			SortDirection: commands.ASCENDING,
		},
		ApiKeyFilter: commands.ApiKeyFilter{
			AssociatedApiKeys: commands.ASSOCIATED_API_KEYS_ALL,
		},
	}

	results, count, err := repository.GetPagedApiKeys(command, "1")

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Should return 2 results (page size)
	if len(results) != 2 {
		utils.PrintTestError(t, len(results), 2)
	}

	// Count should be 5
	if count != 5 {
		utils.PrintTestError(t, count, int64(5))
	}

	// Results should be the next 2 in alphabetical order
	if results[0].Name != "Delta Key" {
		utils.PrintTestError(t, results[0].Name, "Delta Key")
	}

	if results[1].Name != "Echo Key" {
		utils.PrintTestError(t, results[1].Name, "Echo Key")
	}
}

func TestGetPagedApiKeys_NoLimit(t *testing.T) {
	defer TruncateTestDb()
	createTestApiKeysForPagination()

	repository := NewApiKeyRepository(nil)
	command := commands.PagedApiKeyRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:          1,
			PageSize:      -1, // No limit
			OrderBy:       "name",
			SortDirection: commands.ASCENDING,
		},
		ApiKeyFilter: commands.ApiKeyFilter{
			AssociatedApiKeys: commands.ASSOCIATED_API_KEYS_ALL,
		},
	}

	results, count, err := repository.GetPagedApiKeys(command, "1")

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Should return all results
	if len(results) != 5 {
		utils.PrintTestError(t, len(results), 5)
	}

	if count != 5 {
		utils.PrintTestError(t, count, int64(5))
	}
}

func TestGetPagedApiKeys_FilterMineOnly(t *testing.T) {
	defer TruncateTestDb()
	createTestApiKeysForPagination()

	repository := NewApiKeyRepository(nil)
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

	// Request as user 1
	results, count, err := repository.GetPagedApiKeys(command, "1")

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Should return 3 results (user 1 has 3 keys)
	if len(results) != 3 {
		utils.PrintTestError(t, len(results), 3)
	}

	if count != 3 {
		utils.PrintTestError(t, count, int64(3))
	}

	// All results should belong to user 1
	for _, result := range results {
		if *result.UserID != 1 {
			utils.PrintTestError(t, *result.UserID, 1)
		}
	}

	// Results should be Alpha Key, Beta Key, and Echo Key (in alphabetical order)
	if results[0].Name != "Alpha Key" {
		utils.PrintTestError(t, results[0].Name, "Alpha Key")
	}

	if results[1].Name != "Beta Key" {
		utils.PrintTestError(t, results[1].Name, "Beta Key")
	}

	if results[2].Name != "Echo Key" {
		utils.PrintTestError(t, results[2].Name, "Echo Key")
	}
}

func TestGetPagedApiKeys_FilterMineOnlyUser2(t *testing.T) {
	defer TruncateTestDb()
	createTestApiKeysForPagination()

	repository := NewApiKeyRepository(nil)
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

	// Request as user 2
	results, count, err := repository.GetPagedApiKeys(command, "2")

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Should return 2 results (user 2 has 2 keys)
	if len(results) != 2 {
		utils.PrintTestError(t, len(results), 2)
	}

	if count != 2 {
		utils.PrintTestError(t, count, int64(2))
	}

	// Both results should belong to user 2
	for _, result := range results {
		if *result.UserID != 2 {
			utils.PrintTestError(t, *result.UserID, 2)
		}
	}
}

func TestGetPagedApiKeys_SortByCreatedAtDescending(t *testing.T) {
	defer TruncateTestDb()
	createTestApiKeysForPagination()

	repository := NewApiKeyRepository(nil)
	command := commands.PagedApiKeyRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:          1,
			PageSize:      10,
			OrderBy:       "created_at",
			SortDirection: commands.DESCENDING,
		},
		ApiKeyFilter: commands.ApiKeyFilter{
			AssociatedApiKeys: commands.ASSOCIATED_API_KEYS_ALL,
		},
	}

	results, count, err := repository.GetPagedApiKeys(command, "1")

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if len(results) != 5 {
		utils.PrintTestError(t, len(results), 5)
	}

	if count != 5 {
		utils.PrintTestError(t, count, int64(5))
	}

	// Should be sorted by created_at descending
	// Most recent first: Beta Key, Echo Key, Delta Key, Alpha Key, Gamma Key
	if results[0].Name != "Beta Key" {
		utils.PrintTestError(t, results[0].Name, "Beta Key")
	}

	if results[1].Name != "Echo Key" {
		utils.PrintTestError(t, results[1].Name, "Echo Key")
	}

	if results[2].Name != "Delta Key" {
		utils.PrintTestError(t, results[2].Name, "Delta Key")
	}

	if results[3].Name != "Alpha Key" {
		utils.PrintTestError(t, results[3].Name, "Alpha Key")
	}

	if results[4].Name != "Gamma Key" {
		utils.PrintTestError(t, results[4].Name, "Gamma Key")
	}
}

func TestGetPagedApiKeys_SortByDescription(t *testing.T) {
	defer TruncateTestDb()
	createTestApiKeysForPagination()

	repository := NewApiKeyRepository(nil)
	command := commands.PagedApiKeyRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:          1,
			PageSize:      10,
			OrderBy:       "description",
			SortDirection: commands.ASCENDING,
		},
		ApiKeyFilter: commands.ApiKeyFilter{
			AssociatedApiKeys: commands.ASSOCIATED_API_KEYS_ALL,
		},
	}

	results, count, err := repository.GetPagedApiKeys(command, "1")

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if len(results) != 5 {
		utils.PrintTestError(t, len(results), 5)
	}

	if count != 5 {
		utils.PrintTestError(t, count, int64(5))
	}

	// Should be sorted by description
	if results[0].Description != "Fifth key alphabetically" {
		utils.PrintTestError(t, results[0].Description, "Fifth key alphabetically")
	}

	if results[1].Description != "First key alphabetically" {
		utils.PrintTestError(t, results[1].Description, "First key alphabetically")
	}

	if results[2].Description != "Fourth key alphabetically" {
		utils.PrintTestError(t, results[2].Description, "Fourth key alphabetically")
	}
}

func TestGetPagedApiKeys_SortByLastUsedAt(t *testing.T) {
	defer TruncateTestDb()
	createTestApiKeysForPagination()

	repository := NewApiKeyRepository(nil)
	command := commands.PagedApiKeyRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:          1,
			PageSize:      10,
			OrderBy:       "last_used_at",
			SortDirection: commands.DESCENDING,
		},
		ApiKeyFilter: commands.ApiKeyFilter{
			AssociatedApiKeys: commands.ASSOCIATED_API_KEYS_ALL,
		},
	}

	results, count, err := repository.GetPagedApiKeys(command, "1")

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if len(results) != 5 {
		utils.PrintTestError(t, len(results), 5)
	}

	if count != 5 {
		utils.PrintTestError(t, count, int64(5))
	}

	// Beta Key should be first (only one with LastUsedAt set)
	if results[0].Name != "Beta Key" {
		utils.PrintTestError(t, results[0].Name, "Beta Key")
	}
}

func TestGetPagedApiKeys_InvalidColumn(t *testing.T) {
	defer TruncateTestDb()
	createTestApiKeysForPagination()

	repository := NewApiKeyRepository(nil)
	command := commands.PagedApiKeyRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:          1,
			PageSize:      10,
			OrderBy:       "invalid_column",
			SortDirection: commands.ASCENDING,
		},
		ApiKeyFilter: commands.ApiKeyFilter{
			AssociatedApiKeys: commands.ASSOCIATED_API_KEYS_ALL,
		},
	}

	_, _, err := repository.GetPagedApiKeys(command, "1")

	if err == nil {
		utils.PrintTestError(t, err, "an error for invalid column")
	}

	if err.Error() != "invalid column" {
		utils.PrintTestError(t, err.Error(), "invalid column")
	}
}

func TestGetPagedApiKeys_EmptyDatabase(t *testing.T) {
	defer TruncateTestDb()

	repository := NewApiKeyRepository(nil)
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

	results, count, err := repository.GetPagedApiKeys(command, "1")

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if len(results) != 0 {
		utils.PrintTestError(t, len(results), 0)
	}

	if count != 0 {
		utils.PrintTestError(t, count, int64(0))
	}
}

func TestGetPagedApiKeys_WithTestKeys(t *testing.T) {
	defer TruncateTestDb()

	user1 := uint(1)

	// Create test key
	testKey := models.ApiKey{
		ID:          "test-key-1",
		UserID:      &user1,
		CreatedBy:   &user1,
		Name:        "Revoked Key",
		Description: "This is a test key",
		Prefix:      "key",
		Hmac:        "revoked-hmac-1",
		Version:     1,
		Scope:       "read",
	}

	GetDB().Create(&testKey)

	repository := NewApiKeyRepository(nil)
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

	results, count, err := repository.GetPagedApiKeys(command, "1")

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Should return 1 result
	if len(results) != 1 {
		utils.PrintTestError(t, len(results), 1)
	}

	if count != 1 {
		utils.PrintTestError(t, count, int64(1))
	}

	// Verify it's the test key
	if results[0].Name != "Revoked Key" {
		utils.PrintTestError(t, results[0].Name, "Revoked Key")
	}
}

func TestGetPagedApiKeys_WithTransaction(t *testing.T) {
	defer TruncateTestDb()

	db := GetDB()
	tx := db.Begin()
	defer tx.Rollback()

	user1 := uint(1)
	apiKey := models.ApiKey{
		ID:          "tx-key-1",
		UserID:      &user1,
		CreatedBy:   &user1,
		Name:        "Transaction Key",
		Description: "Key created in transaction",
		Prefix:      "key",
		Hmac:        "tx-hmac-1",
		Version:     1,
		Scope:       "read",
	}

	// Create key within transaction
	tx.Create(&apiKey)

	repository := NewApiKeyRepository(tx)
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

	// Should find the key within the transaction
	results, count, err := repository.GetPagedApiKeys(command, "1")

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if len(results) != 1 {
		utils.PrintTestError(t, len(results), 1)
	}

	if count != 1 {
		utils.PrintTestError(t, count, int64(1))
	}

	if results[0].Name != "Transaction Key" {
		utils.PrintTestError(t, results[0].Name, "Transaction Key")
	}

	// Outside the transaction, the key should not be visible
	repositoryOutside := NewApiKeyRepository(nil)
	resultsOutside, countOutside, err := repositoryOutside.GetPagedApiKeys(command, "1")

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if len(resultsOutside) != 0 {
		utils.PrintTestError(t, len(resultsOutside), 0)
	}

	if countOutside != 0 {
		utils.PrintTestError(t, countOutside, int64(0))
	}
}

func TestGetPagedApiKeys_LargePageSize(t *testing.T) {
	defer TruncateTestDb()
	createTestApiKeysForPagination()

	repository := NewApiKeyRepository(nil)
	command := commands.PagedApiKeyRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:          1,
			PageSize:      1000, // Should be capped to 100
			OrderBy:       "name",
			SortDirection: commands.ASCENDING,
		},
		ApiKeyFilter: commands.ApiKeyFilter{
			AssociatedApiKeys: commands.ASSOCIATED_API_KEYS_ALL,
		},
	}

	results, count, err := repository.GetPagedApiKeys(command, "1")

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Should return all 5 keys (since we have fewer than 100)
	if len(results) != 5 {
		utils.PrintTestError(t, len(results), 5)
	}

	if count != 5 {
		utils.PrintTestError(t, count, int64(5))
	}
}

func TestGetPagedApiKeys_ZeroPage(t *testing.T) {
	defer TruncateTestDb()
	createTestApiKeysForPagination()

	repository := NewApiKeyRepository(nil)
	command := commands.PagedApiKeyRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:          0, // Should default to page 1
			PageSize:      2,
			OrderBy:       "name",
			SortDirection: commands.ASCENDING,
		},
		ApiKeyFilter: commands.ApiKeyFilter{
			AssociatedApiKeys: commands.ASSOCIATED_API_KEYS_ALL,
		},
	}

	results, count, err := repository.GetPagedApiKeys(command, "1")

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Should return first 2 results (same as page 1)
	if len(results) != 2 {
		utils.PrintTestError(t, len(results), 2)
	}

	if count != 5 {
		utils.PrintTestError(t, count, int64(5))
	}

	if results[0].Name != "Alpha Key" {
		utils.PrintTestError(t, results[0].Name, "Alpha Key")
	}
}

func TestGetPagedApiKeys_AllValidColumns(t *testing.T) {
	defer TruncateTestDb()
	createTestApiKeysForPagination()

	repository := NewApiKeyRepository(nil)
	validColumns := []string{"name", "description", "created_at", "updated_at", "last_used_at"}

	for _, column := range validColumns {
		command := commands.PagedApiKeyRequestCommand{
			PagedRequestCommand: commands.PagedRequestCommand{
				Page:          1,
				PageSize:      10,
				OrderBy:       column,
				SortDirection: commands.ASCENDING,
			},
			ApiKeyFilter: commands.ApiKeyFilter{
				AssociatedApiKeys: commands.ASSOCIATED_API_KEYS_ALL,
			},
		}

		_, _, err := repository.GetPagedApiKeys(command, "1")

		if err != nil {
			utils.PrintTestError(t, err, "no error for column "+column)
		}
	}
}

func TestGetApiKeyById_Success(t *testing.T) {
	defer TruncateTestDb()

	userId := uint(1)
	apiKey := models.ApiKey{
		ID:          "test-get-key-id",
		UserID:      &userId,
		CreatedBy:   &userId,
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
		CreatedBy:   &userId,
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
		CreatedBy:   &userId,
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

func TestUpdateApiKeyLastUsedDate_Success(t *testing.T) {
	defer TruncateTestDb()

	userId := uint(1)
	apiKey := models.ApiKey{
		ID:          "update-last-used-key",
		UserID:      &userId,
		CreatedBy:   &userId,
		Name:        "Last Used Test Key",
		Description: "Test updating last used date",
		Prefix:      "key",
		Hmac:        "last-used-hmac",
		Version:     1,
		Scope:       "read",
	}

	repository := NewApiKeyRepository(nil)

	// Create the API key first
	_, err := repository.CreateApiKey(apiKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify LastUsedAt is initially nil
	retrievedApiKey, err := repository.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if retrievedApiKey.LastUsedAt != nil {
		utils.PrintTestError(t, retrievedApiKey.LastUsedAt, nil)
	}

	beforeUpdate := time.Now()

	// Update the last used date
	err = repository.UpdateApiKeyLastUsedDate(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	afterUpdate := time.Now()

	// Verify the last used date was updated
	updatedApiKey, err := repository.GetApiKeyById(apiKey.ID)
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

func TestUpdateApiKeyLastUsedDate_NonExistentKey(t *testing.T) {
	defer TruncateTestDb()

	repository := NewApiKeyRepository(nil)

	// Try to update non-existent API key
	err := repository.UpdateApiKeyLastUsedDate("non-existent-key")

	// Should not return an error (GORM returns no error for UPDATE on non-existent records)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}
}

func TestUpdateApiKeyLastUsedDate_WithTransaction(t *testing.T) {
	defer TruncateTestDb()

	db := GetDB()
	tx := db.Begin()
	defer tx.Rollback()

	userId := uint(1)
	apiKey := models.ApiKey{
		ID:          "tx-update-last-used-key",
		UserID:      &userId,
		CreatedBy:   &userId,
		Name:        "Transaction Last Used Test",
		Description: "Test updating last used date within transaction",
		Prefix:      "key",
		Hmac:        "tx-last-used-hmac",
		Version:     1,
		Scope:       "read",
	}

	repository := NewApiKeyRepository(tx)

	// Create API key within transaction
	_, err := repository.CreateApiKey(apiKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Update last used date within transaction
	err = repository.UpdateApiKeyLastUsedDate(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify update within transaction
	updatedApiKey, err := repository.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if updatedApiKey.LastUsedAt == nil {
		utils.PrintTestError(t, "LastUsedAt should not be nil", "LastUsedAt should be set")
	}

	// Commit transaction
	tx.Commit()

	// Verify update persisted after commit
	repositoryOutside := NewApiKeyRepository(nil)
	persistedApiKey, err := repositoryOutside.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if persistedApiKey.LastUsedAt == nil {
		utils.PrintTestError(t, "LastUsedAt should not be nil after commit", "LastUsedAt should be set")
	}

	if !persistedApiKey.LastUsedAt.Equal(*updatedApiKey.LastUsedAt) {
		utils.PrintTestError(t, persistedApiKey.LastUsedAt, updatedApiKey.LastUsedAt)
	}
}

func TestUpdateApiKeyLastUsedDate_VerifyTimestamp(t *testing.T) {
	defer TruncateTestDb()

	userId := uint(1)
	apiKey := models.ApiKey{
		ID:          "timestamp-update-key",
		UserID:      &userId,
		CreatedBy:   &userId,
		Name:        "Timestamp Update Test",
		Description: "Test timestamp precision",
		Prefix:      "key",
		Hmac:        "timestamp-update-hmac",
		Version:     1,
		Scope:       "read",
	}

	repository := NewApiKeyRepository(nil)

	// Create the API key
	_, err := repository.CreateApiKey(apiKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Wait a small amount to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	beforeUpdate := time.Now()

	// Update last used date
	err = repository.UpdateApiKeyLastUsedDate(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	afterUpdate := time.Now()

	// Get the updated key
	updatedApiKey, err := repository.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify timestamp is after creation
	if !updatedApiKey.LastUsedAt.After(updatedApiKey.CreatedAt) {
		utils.PrintTestError(t, "LastUsedAt should be after CreatedAt", "LastUsedAt should be more recent")
	}

	// Verify timestamp is within expected range
	if updatedApiKey.LastUsedAt.Before(beforeUpdate) {
		utils.PrintTestError(t, "LastUsedAt is before update time", "LastUsedAt should be after update started")
	}

	if updatedApiKey.LastUsedAt.After(afterUpdate) {
		utils.PrintTestError(t, "LastUsedAt is after update time", "LastUsedAt should be before update completed")
	}

	// Test multiple updates
	time.Sleep(10 * time.Millisecond)
	firstUpdateTime := *updatedApiKey.LastUsedAt

	beforeSecondUpdate := time.Now()
	err = repository.UpdateApiKeyLastUsedDate(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	secondUpdatedApiKey, err := repository.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify second update is later than first
	if !secondUpdatedApiKey.LastUsedAt.After(firstUpdateTime) {
		utils.PrintTestError(t, "Second LastUsedAt should be after first", "Second update should be more recent")
	}

	if secondUpdatedApiKey.LastUsedAt.Before(beforeSecondUpdate) {
		utils.PrintTestError(t, "Second LastUsedAt is before second update", "Second update should be recent")
	}
}

// UpdateApiKey Repository Tests

func TestUpdateApiKey_Success(t *testing.T) {
	defer TruncateTestDb()

	userId := uint(1)
	apiKey := models.ApiKey{
		ID:          "update-success-key",
		UserID:      &userId,
		CreatedBy:   &userId,
		Name:        "Original Name",
		Description: "Original description",
		Prefix:      "key",
		Hmac:        "original-hmac",
		Version:     1,
		Scope:       "r",
	}

	repository := NewApiKeyRepository(nil)

	// Create the API key first
	_, err := repository.CreateApiKey(apiKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Update the API key
	err = repository.UpdateApiKey(apiKey.ID, userId, "Updated Name", "Updated description", "rw")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify the update
	updatedApiKey, err := repository.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if updatedApiKey.Name != "Updated Name" {
		utils.PrintTestError(t, updatedApiKey.Name, "Updated Name")
	}

	if updatedApiKey.Description != "Updated description" {
		utils.PrintTestError(t, updatedApiKey.Description, "Updated description")
	}

	if updatedApiKey.Scope != "rw" {
		utils.PrintTestError(t, updatedApiKey.Scope, "rw")
	}

	// Verify other fields remain unchanged
	if updatedApiKey.ID != apiKey.ID {
		utils.PrintTestError(t, updatedApiKey.ID, apiKey.ID)
	}

	if updatedApiKey.Prefix != apiKey.Prefix {
		utils.PrintTestError(t, updatedApiKey.Prefix, apiKey.Prefix)
	}

	if updatedApiKey.Hmac != apiKey.Hmac {
		utils.PrintTestError(t, updatedApiKey.Hmac, apiKey.Hmac)
	}

	if updatedApiKey.Version != apiKey.Version {
		utils.PrintTestError(t, updatedApiKey.Version, apiKey.Version)
	}

	if *updatedApiKey.UserID != *apiKey.UserID {
		utils.PrintTestError(t, *updatedApiKey.UserID, *apiKey.UserID)
	}
}

func TestUpdateApiKey_OnlyName(t *testing.T) {
	defer TruncateTestDb()

	userId := uint(1)
	apiKey := models.ApiKey{
		ID:          "update-name-only-key",
		UserID:      &userId,
		CreatedBy:   &userId,
		Name:        "Original Name",
		Description: "Original description",
		Prefix:      "key",
		Hmac:        "original-hmac",
		Version:     1,
		Scope:       "w",
	}

	repository := NewApiKeyRepository(nil)

	// Create the API key first
	_, err := repository.CreateApiKey(apiKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Update only the name
	err = repository.UpdateApiKey(apiKey.ID, userId, "New Name Only", apiKey.Description, apiKey.Scope)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify the update
	updatedApiKey, err := repository.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if updatedApiKey.Name != "New Name Only" {
		utils.PrintTestError(t, updatedApiKey.Name, "New Name Only")
	}

	// Verify other fields remain unchanged
	if updatedApiKey.Description != apiKey.Description {
		utils.PrintTestError(t, updatedApiKey.Description, apiKey.Description)
	}

	if updatedApiKey.Scope != apiKey.Scope {
		utils.PrintTestError(t, updatedApiKey.Scope, apiKey.Scope)
	}
}

func TestUpdateApiKey_OnlyDescription(t *testing.T) {
	defer TruncateTestDb()

	userId := uint(1)
	apiKey := models.ApiKey{
		ID:          "update-desc-only-key",
		UserID:      &userId,
		CreatedBy:   &userId,
		Name:        "Test Name",
		Description: "Original description",
		Prefix:      "key",
		Hmac:        "original-hmac",
		Version:     1,
		Scope:       "r",
	}

	repository := NewApiKeyRepository(nil)

	// Create the API key first
	_, err := repository.CreateApiKey(apiKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Update only the description
	err = repository.UpdateApiKey(apiKey.ID, userId, apiKey.Name, "New Description Only", apiKey.Scope)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify the update
	updatedApiKey, err := repository.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if updatedApiKey.Description != "New Description Only" {
		utils.PrintTestError(t, updatedApiKey.Description, "New Description Only")
	}

	// Verify other fields remain unchanged
	if updatedApiKey.Name != apiKey.Name {
		utils.PrintTestError(t, updatedApiKey.Name, apiKey.Name)
	}

	if updatedApiKey.Scope != apiKey.Scope {
		utils.PrintTestError(t, updatedApiKey.Scope, apiKey.Scope)
	}
}

func TestUpdateApiKey_OnlyScope(t *testing.T) {
	defer TruncateTestDb()

	userId := uint(1)
	apiKey := models.ApiKey{
		ID:          "update-scope-only-key",
		UserID:      &userId,
		CreatedBy:   &userId,
		Name:        "Test Name",
		Description: "Test description",
		Prefix:      "key",
		Hmac:        "original-hmac",
		Version:     1,
		Scope:       "r",
	}

	repository := NewApiKeyRepository(nil)

	// Create the API key first
	_, err := repository.CreateApiKey(apiKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Update only the scope
	err = repository.UpdateApiKey(apiKey.ID, userId, apiKey.Name, apiKey.Description, "w")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify the update
	updatedApiKey, err := repository.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if updatedApiKey.Scope != "w" {
		utils.PrintTestError(t, updatedApiKey.Scope, "w")
	}

	// Verify other fields remain unchanged
	if updatedApiKey.Name != apiKey.Name {
		utils.PrintTestError(t, updatedApiKey.Name, apiKey.Name)
	}

	if updatedApiKey.Description != apiKey.Description {
		utils.PrintTestError(t, updatedApiKey.Description, apiKey.Description)
	}
}

func TestUpdateApiKey_AllValidScopes(t *testing.T) {
	defer TruncateTestDb()

	userId := uint(1)
	validScopes := []string{"r", "w", "rw"}

	for i, scope := range validScopes {
		keyId := fmt.Sprintf("scope-test-key-%d", i)
		apiKey := models.ApiKey{
			ID:          keyId,
			UserID:      &userId,
			CreatedBy:   &userId,
			Name:        fmt.Sprintf("Scope Test Key %d", i),
			Description: fmt.Sprintf("Testing scope %s", scope),
			Prefix:      "key",
			Hmac:        fmt.Sprintf("scope-hmac-%d", i),
			Version:     1,
			Scope:       "r", // Start with read scope
		}

		repository := NewApiKeyRepository(nil)

		// Create the API key
		_, err := repository.CreateApiKey(apiKey)
		if err != nil {
			utils.PrintTestError(t, err, "no error")
		}

		// Update to the test scope
		err = repository.UpdateApiKey(keyId, userId, apiKey.Name, apiKey.Description, scope)
		if err != nil {
			utils.PrintTestError(t, err, fmt.Sprintf("no error for scope %s", scope))
		}

		// Verify the scope was updated
		updatedApiKey, err := repository.GetApiKeyById(keyId)
		if err != nil {
			utils.PrintTestError(t, err, "no error")
		}

		if updatedApiKey.Scope != scope {
			utils.PrintTestError(t, updatedApiKey.Scope, scope)
		}
	}
}

func TestUpdateApiKey_WrongUser(t *testing.T) {
	defer TruncateTestDb()

	userId1 := uint(1)
	userId2 := uint(2)
	apiKey := models.ApiKey{
		ID:          "wrong-user-key",
		UserID:      &userId1,
		CreatedBy:   &userId1,
		Name:        "User 1 Key",
		Description: "Belongs to user 1",
		Prefix:      "key",
		Hmac:        "wrong-user-hmac",
		Version:     1,
		Scope:       "r",
	}

	repository := NewApiKeyRepository(nil)

	// Create the API key for user 1
	_, err := repository.CreateApiKey(apiKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Try to update with user 2 (should not update anything)
	err = repository.UpdateApiKey(apiKey.ID, userId2, "Hacked Name", "Hacked description", "rw")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify the API key was NOT updated
	unchangedApiKey, err := repository.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if unchangedApiKey.Name != apiKey.Name {
		utils.PrintTestError(t, unchangedApiKey.Name, apiKey.Name)
	}

	if unchangedApiKey.Description != apiKey.Description {
		utils.PrintTestError(t, unchangedApiKey.Description, apiKey.Description)
	}

	if unchangedApiKey.Scope != apiKey.Scope {
		utils.PrintTestError(t, unchangedApiKey.Scope, apiKey.Scope)
	}
}

func TestUpdateApiKey_NonExistentKey(t *testing.T) {
	defer TruncateTestDb()

	repository := NewApiKeyRepository(nil)

	// Try to update non-existent API key
	err := repository.UpdateApiKey("non-existent-key", 1, "New Name", "New description", "rw")

	// Should not return an error (GORM returns no error for UPDATE on non-existent records)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}
}

func TestUpdateApiKey_EmptyId(t *testing.T) {
	defer TruncateTestDb()

	repository := NewApiKeyRepository(nil)

	// Try to update with empty ID
	err := repository.UpdateApiKey("", 1, "New Name", "New description", "rw")

	// Should not return an error (GORM returns no error for UPDATE on non-existent records)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}
}

func TestUpdateApiKey_EmptyFields(t *testing.T) {
	defer TruncateTestDb()

	userId := uint(1)
	apiKey := models.ApiKey{
		ID:          "empty-fields-key",
		UserID:      &userId,
		CreatedBy:   &userId,
		Name:        "Original Name",
		Description: "Original description",
		Prefix:      "key",
		Hmac:        "empty-fields-hmac",
		Version:     1,
		Scope:       "r",
	}

	repository := NewApiKeyRepository(nil)

	// Create the API key first
	_, err := repository.CreateApiKey(apiKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Update with empty values
	err = repository.UpdateApiKey(apiKey.ID, userId, "", "", "")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify the fields were updated to empty values
	updatedApiKey, err := repository.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if updatedApiKey.Name != "" {
		utils.PrintTestError(t, updatedApiKey.Name, "")
	}

	if updatedApiKey.Description != "" {
		utils.PrintTestError(t, updatedApiKey.Description, "")
	}

	if updatedApiKey.Scope != "" {
		utils.PrintTestError(t, updatedApiKey.Scope, "")
	}
}

func TestUpdateApiKey_InvalidUserId(t *testing.T) {
	defer TruncateTestDb()

	userId := uint(1)
	apiKey := models.ApiKey{
		ID:          "invalid-user-key",
		UserID:      &userId,
		CreatedBy:   &userId,
		Name:        "Valid User Key",
		Description: "Belongs to valid user",
		Prefix:      "key",
		Hmac:        "invalid-user-hmac",
		Version:     1,
		Scope:       "r",
	}

	repository := NewApiKeyRepository(nil)

	// Create the API key
	_, err := repository.CreateApiKey(apiKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Try to update with invalid user ID (0)
	err = repository.UpdateApiKey(apiKey.ID, 0, "Invalid Update", "Invalid description", "w")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify the API key was NOT updated
	unchangedApiKey, err := repository.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if unchangedApiKey.Name != apiKey.Name {
		utils.PrintTestError(t, unchangedApiKey.Name, apiKey.Name)
	}

	// Try to update with very large user ID
	err = repository.UpdateApiKey(apiKey.ID, 999999, "Invalid Update", "Invalid description", "w")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify the API key was still NOT updated
	stillUnchangedApiKey, err := repository.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if stillUnchangedApiKey.Name != apiKey.Name {
		utils.PrintTestError(t, stillUnchangedApiKey.Name, apiKey.Name)
	}
}

func TestUpdateApiKey_WithTransaction(t *testing.T) {
	defer TruncateTestDb()

	db := GetDB()
	tx := db.Begin()
	defer tx.Rollback()

	userId := uint(1)
	apiKey := models.ApiKey{
		ID:          "transaction-update-key",
		UserID:      &userId,
		CreatedBy:   &userId,
		Name:        "Transaction Test Key",
		Description: "Test updating within transaction",
		Prefix:      "key",
		Hmac:        "transaction-update-hmac",
		Version:     1,
		Scope:       "r",
	}

	repository := NewApiKeyRepository(tx)

	// Create API key within transaction
	_, err := repository.CreateApiKey(apiKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Update API key within transaction
	err = repository.UpdateApiKey(apiKey.ID, userId, "Updated in TX", "Updated description in TX", "w")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify update within transaction
	updatedApiKey, err := repository.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if updatedApiKey.Name != "Updated in TX" {
		utils.PrintTestError(t, updatedApiKey.Name, "Updated in TX")
	}

	if updatedApiKey.Description != "Updated description in TX" {
		utils.PrintTestError(t, updatedApiKey.Description, "Updated description in TX")
	}

	if updatedApiKey.Scope != "w" {
		utils.PrintTestError(t, updatedApiKey.Scope, "w")
	}

	// Commit transaction
	tx.Commit()

	// Verify update persisted after commit
	repositoryOutside := NewApiKeyRepository(nil)
	persistedApiKey, err := repositoryOutside.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if persistedApiKey.Name != "Updated in TX" {
		utils.PrintTestError(t, persistedApiKey.Name, "Updated in TX")
	}

	if persistedApiKey.Description != "Updated description in TX" {
		utils.PrintTestError(t, persistedApiKey.Description, "Updated description in TX")
	}

	if persistedApiKey.Scope != "w" {
		utils.PrintTestError(t, persistedApiKey.Scope, "w")
	}
}

func TestUpdateApiKey_TransactionRollback(t *testing.T) {
	defer TruncateTestDb()

	userId := uint(1)
	apiKey := models.ApiKey{
		ID:          "rollback-update-key",
		UserID:      &userId,
		CreatedBy:   &userId,
		Name:        "Rollback Test Key",
		Description: "Test rollback of update",
		Prefix:      "key",
		Hmac:        "rollback-update-hmac",
		Version:     1,
		Scope:       "r",
	}

	// First create the key outside of transaction
	repository := NewApiKeyRepository(nil)
	_, err := repository.CreateApiKey(apiKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	db := GetDB()
	tx := db.Begin()

	repositoryTx := NewApiKeyRepository(tx)

	// Update API key within transaction
	err = repositoryTx.UpdateApiKey(apiKey.ID, userId, "Updated in TX", "Updated description in TX", "w")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify update within transaction
	updatedApiKey, err := repositoryTx.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if updatedApiKey.Name != "Updated in TX" {
		utils.PrintTestError(t, updatedApiKey.Name, "Updated in TX")
	}

	// Rollback transaction
	tx.Rollback()

	// Verify update was rolled back
	unchangedApiKey, err := repository.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if unchangedApiKey.Name != apiKey.Name {
		utils.PrintTestError(t, unchangedApiKey.Name, apiKey.Name)
	}

	if unchangedApiKey.Description != apiKey.Description {
		utils.PrintTestError(t, unchangedApiKey.Description, apiKey.Description)
	}

	if unchangedApiKey.Scope != apiKey.Scope {
		utils.PrintTestError(t, unchangedApiKey.Scope, apiKey.Scope)
	}
}

func TestUpdateApiKey_VerifyTimestampUpdate(t *testing.T) {
	defer TruncateTestDb()

	userId := uint(1)
	apiKey := models.ApiKey{
		ID:          "timestamp-update-key",
		UserID:      &userId,
		CreatedBy:   &userId,
		Name:        "Timestamp Test Key",
		Description: "Test timestamp update",
		Prefix:      "key",
		Hmac:        "timestamp-update-hmac",
		Version:     1,
		Scope:       "r",
	}

	repository := NewApiKeyRepository(nil)

	// Create the API key
	_, err := repository.CreateApiKey(apiKey)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Get initial timestamps
	initialApiKey, err := repository.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	initialUpdatedAt := initialApiKey.UpdatedAt

	// Wait a small amount to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	beforeUpdate := time.Now()

	// Update the API key
	err = repository.UpdateApiKey(apiKey.ID, userId, "Updated Name", "Updated description", "w")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	afterUpdate := time.Now()

	// Get updated API key
	updatedApiKey, err := repository.GetApiKeyById(apiKey.ID)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify UpdatedAt was changed
	if updatedApiKey.UpdatedAt.Equal(initialUpdatedAt) {
		utils.PrintTestError(t, "UpdatedAt should be different", "UpdatedAt should be updated")
	}

	// Verify UpdatedAt is after the update time
	if updatedApiKey.UpdatedAt.Before(beforeUpdate) {
		utils.PrintTestError(t, "UpdatedAt is before update time", "UpdatedAt should be after update started")
	}

	if updatedApiKey.UpdatedAt.After(afterUpdate) {
		utils.PrintTestError(t, "UpdatedAt is after update time", "UpdatedAt should be before update completed")
	}

	// Verify CreatedAt was not changed
	if !updatedApiKey.CreatedAt.Equal(initialApiKey.CreatedAt) {
		utils.PrintTestError(t, updatedApiKey.CreatedAt, initialApiKey.CreatedAt)
	}
}

func TestUpdateApiKey_MultipleKeys(t *testing.T) {
	defer TruncateTestDb()

	userId1 := uint(1)
	userId2 := uint(2)

	// Create multiple API keys for different users
	apiKeys := []models.ApiKey{
		{
			ID:          "multi-key-1",
			UserID:      &userId1,
			CreatedBy:   &userId1,
			Name:        "User 1 Key A",
			Description: "First key for user 1",
			Prefix:      "key",
			Hmac:        "multi-hmac-1",
			Version:     1,
			Scope:       "r",
		},
		{
			ID:          "multi-key-2",
			UserID:      &userId1,
			CreatedBy:   &userId1,
			Name:        "User 1 Key B",
			Description: "Second key for user 1",
			Prefix:      "key",
			Hmac:        "multi-hmac-2",
			Version:     1,
			Scope:       "w",
		},
		{
			ID:          "multi-key-3",
			UserID:      &userId2,
			CreatedBy:   &userId2,
			Name:        "User 2 Key A",
			Description: "First key for user 2",
			Prefix:      "key",
			Hmac:        "multi-hmac-3",
			Version:     1,
			Scope:       "rw",
		},
	}

	repository := NewApiKeyRepository(nil)

	// Create all API keys
	for _, apiKey := range apiKeys {
		_, err := repository.CreateApiKey(apiKey)
		if err != nil {
			utils.PrintTestError(t, err, "no error")
		}
	}

	// Update one specific key
	err := repository.UpdateApiKey("multi-key-2", userId1, "Updated User 1 Key B", "Updated second key", "rw")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify only the targeted key was updated
	updatedKey, err := repository.GetApiKeyById("multi-key-2")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if updatedKey.Name != "Updated User 1 Key B" {
		utils.PrintTestError(t, updatedKey.Name, "Updated User 1 Key B")
	}

	if updatedKey.Description != "Updated second key" {
		utils.PrintTestError(t, updatedKey.Description, "Updated second key")
	}

	if updatedKey.Scope != "rw" {
		utils.PrintTestError(t, updatedKey.Scope, "rw")
	}

	// Verify other keys remain unchanged
	unchangedKey1, err := repository.GetApiKeyById("multi-key-1")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if unchangedKey1.Name != "User 1 Key A" {
		utils.PrintTestError(t, unchangedKey1.Name, "User 1 Key A")
	}

	unchangedKey3, err := repository.GetApiKeyById("multi-key-3")
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if unchangedKey3.Name != "User 2 Key A" {
		utils.PrintTestError(t, unchangedKey3.Name, "User 2 Key A")
	}
}
