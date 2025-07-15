package services

import (
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func createTestItems() {
	db := repositories.GetDB()
	
	items := []map[string]interface{}{
		{
			"id":                 1,
			"created_at":         "2024-01-01 10:00:00",
			"updated_at":         "2024-01-01 10:00:00",
			"created_by":         1,
			"created_by_string":  "test user",
			"amount":             "10.50",
			"charged_to_user_id": 1,
			"is_taxed":           true,
			"name":               "Test Item 1",
			"receipt_id":         1,
			"status":             "OPEN",
		},
		{
			"id":                 2,
			"created_at":         "2024-01-01 11:00:00",
			"updated_at":         "2024-01-01 11:00:00",
			"created_by":         1,
			"created_by_string":  "test user",
			"amount":             "25.75",
			"charged_to_user_id": 1,
			"is_taxed":           false,
			"name":               "Test Item 2",
			"receipt_id":         1,
			"status":             "RESOLVED",
		},
	}
	
	for _, item := range items {
		db.Table("items").Create(item)
	}
}

func createTestItemCategories() {
	db := repositories.GetDB()
	
	itemCategories := []ItemCategoryData{
		{ReceiptId: "1", CategoryId: "1"},
		{ReceiptId: "1", CategoryId: "2"},
		{ReceiptId: "2", CategoryId: "1"},
	}
	
	db.Table("item_categories").Create(&itemCategories)
}

func createTestItemTags() {
	db := repositories.GetDB()
	
	itemTags := []ItemTagData{
		{ReceiptId: "1", TagId: "1"},
		{ReceiptId: "1", TagId: "2"},
		{ReceiptId: "2", TagId: "3"},
	}
	
	db.Table("item_tags").Create(&itemTags)
}

func createTestReceipts() {
	db := repositories.GetDB()
	receipt1 := models.Receipt{
		BaseModel: models.BaseModel{ID: 1},
		Name:      "Test Receipt 1",
		Amount:    decimal.NewFromFloat(100.00),
		Date:      time.Now(),
		PaidByUserID: 1,
		Status:    models.OPEN,
		GroupId:   1,
	}
	db.Create(&receipt1)
}

func createTestTags() {
	db := repositories.GetDB()
	tag1 := models.Tag{
		BaseModel: models.BaseModel{ID: 1},
		Name:      "test-tag-1",
	}
	tag2 := models.Tag{
		BaseModel: models.BaseModel{ID: 2},
		Name:      "test-tag-2",
	}
	tag3 := models.Tag{
		BaseModel: models.BaseModel{ID: 3},
		Name:      "test-tag-3",
	}
	db.Create(&tag1)
	db.Create(&tag2)
	db.Create(&tag3)
}

func createOldTables() {
	db := repositories.GetDB()
	
	db.Exec(`CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at TEXT,
		updated_at TEXT,
		created_by INTEGER,
		created_by_string TEXT,
		amount TEXT,
		charged_to_user_id INTEGER,
		is_taxed BOOLEAN,
		name TEXT,
		receipt_id INTEGER,
		status TEXT
	)`)
	
	db.Exec(`CREATE TABLE IF NOT EXISTS item_categories (
		receipt_id TEXT,
		category_id TEXT
	)`)
	
	db.Exec(`CREATE TABLE IF NOT EXISTS item_tags (
		receipt_id TEXT,
		tag_id TEXT
	)`)
}

func countRowsInTable(tableName string) int64 {
	db := repositories.GetDB()
	var count int64
	db.Table(tableName).Count(&count)
	return count
}

func uintPtr(val uint) *uint {
	return &val
}

func TestMigrateItemsToShares_SuccessfulMigration(t *testing.T) {
	defer repositories.TruncateTestDb()
	
	repositories.CreateTestGroupWithUsers()
	repositories.CreateTestCategories()
	createTestTags()
	createTestReceipts()
	createOldTables()
	createTestItems()
	createTestItemCategories()
	createTestItemTags()
	
	itemsCountBefore := countRowsInTable("items")
	itemCategoriesCountBefore := countRowsInTable("item_categories")
	itemTagsCountBefore := countRowsInTable("item_tags")
	
	err := MigrateItemsToShares()
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}
	
	sharesCountAfter := countRowsInTable("shares")
	shareCategoriesCountAfter := countRowsInTable("share_categories")
	shareTagsCountAfter := countRowsInTable("share_tags")
	
	if sharesCountAfter != itemsCountBefore {
		utils.PrintTestError(t, sharesCountAfter, itemsCountBefore)
	}
	
	if shareCategoriesCountAfter != itemCategoriesCountBefore {
		utils.PrintTestError(t, shareCategoriesCountAfter, itemCategoriesCountBefore)
	}
	
	if shareTagsCountAfter != itemTagsCountBefore {
		utils.PrintTestError(t, shareTagsCountAfter, itemTagsCountBefore)
	}
}

func TestMigrateItemsToShares_EmptyTables(t *testing.T) {
	defer repositories.TruncateTestDb()
	
	repositories.CreateTestGroupWithUsers()
	repositories.CreateTestCategories()
	createTestTags()
	createTestReceipts()
	createOldTables()
	
	err := MigrateItemsToShares()
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}
	
	sharesCount := countRowsInTable("shares")
	shareCategoriesCount := countRowsInTable("share_categories")
	shareTagsCount := countRowsInTable("share_tags")
	
	if sharesCount != 0 {
		utils.PrintTestError(t, sharesCount, 0)
	}
	
	if shareCategoriesCount != 0 {
		utils.PrintTestError(t, shareCategoriesCount, 0)
	}
	
	if shareTagsCount != 0 {
		utils.PrintTestError(t, shareTagsCount, 0)
	}
}

func TestMigrateItemsToShares_MissingOldTables(t *testing.T) {
	defer repositories.TruncateTestDb()
	
	repositories.CreateTestGroupWithUsers()
	repositories.CreateTestCategories()
	createTestTags()
	createTestReceipts()
	
	err := MigrateItemsToShares()
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}
	
	sharesCount := countRowsInTable("shares")
	shareCategoriesCount := countRowsInTable("share_categories")
	shareTagsCount := countRowsInTable("share_tags")
	
	if sharesCount != 0 {
		utils.PrintTestError(t, sharesCount, 0)
	}
	
	if shareCategoriesCount != 0 {
		utils.PrintTestError(t, shareCategoriesCount, 0)
	}
	
	if shareTagsCount != 0 {
		utils.PrintTestError(t, shareTagsCount, 0)
	}
}

func TestMigrateItemsToShares_PartialTables(t *testing.T) {
	defer repositories.TruncateTestDb()
	
	repositories.CreateTestGroupWithUsers()
	repositories.CreateTestCategories()
	createTestTags()
	createTestReceipts()
	
	db := repositories.GetDB()
	db.Exec(`CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at TEXT,
		updated_at TEXT,
		created_by INTEGER,
		created_by_string TEXT,
		amount TEXT,
		charged_to_user_id INTEGER,
		is_taxed BOOLEAN,
		name TEXT,
		receipt_id INTEGER,
		status TEXT
	)`)
	
	createTestItems()
	
	err := MigrateItemsToShares()
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}
	
	sharesCount := countRowsInTable("shares")
	shareCategoriesCount := countRowsInTable("share_categories")
	shareTagsCount := countRowsInTable("share_tags")
	
	if sharesCount != 2 {
		utils.PrintTestError(t, sharesCount, 2)
	}
	
	if shareCategoriesCount != 0 {
		utils.PrintTestError(t, shareCategoriesCount, 0)
	}
	
	if shareTagsCount != 0 {
		utils.PrintTestError(t, shareTagsCount, 0)
	}
}

func TestMigrateItemsToShares_DataIntegrity(t *testing.T) {
	defer repositories.TruncateTestDb()
	
	repositories.CreateTestGroupWithUsers()
	repositories.CreateTestCategories()
	createTestTags()
	createTestReceipts()
	createOldTables()
	createTestItems()
	
	err := MigrateItemsToShares()
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}
	
	db := repositories.GetDB()
	var originalItem map[string]interface{}
	var migratedShare models.Share
	
	db.Table("items").Where("id = ?", 1).Take(&originalItem)
	db.Table("shares").Where("id = ?", 1).First(&migratedShare)
	
	if uint(originalItem["id"].(int64)) != migratedShare.ID {
		utils.PrintTestError(t, migratedShare.ID, originalItem["id"])
	}
	
	if originalItem["name"].(string) != migratedShare.Name {
		utils.PrintTestError(t, migratedShare.Name, originalItem["name"])
	}
	
	originalAmount, _ := decimal.NewFromString(originalItem["amount"].(string))
	if !originalAmount.Equal(migratedShare.Amount) {
		utils.PrintTestError(t, migratedShare.Amount, originalAmount)
	}
	
	if uint(originalItem["charged_to_user_id"].(int64)) != migratedShare.ChargedToUserId {
		utils.PrintTestError(t, migratedShare.ChargedToUserId, originalItem["charged_to_user_id"])
	}
	
	if originalItem["is_taxed"].(bool) != migratedShare.IsTaxed {
		utils.PrintTestError(t, migratedShare.IsTaxed, originalItem["is_taxed"])
	}
	
	if models.ShareStatus(originalItem["status"].(string)) != migratedShare.Status {
		utils.PrintTestError(t, migratedShare.Status, originalItem["status"])
	}
}

func TestMigrateItemsToShares_LargeDataset(t *testing.T) {
	defer repositories.TruncateTestDb()
	
	repositories.CreateTestGroupWithUsers()
	repositories.CreateTestCategories()
	createTestTags()
	createTestReceipts()
	createOldTables()
	
	db := repositories.GetDB()
	
	for i := 1; i <= 2500; i++ {
		item := map[string]interface{}{
			"id":                 i,
			"created_at":         "2024-01-01 10:00:00",
			"updated_at":         "2024-01-01 10:00:00",
			"created_by":         1,
			"created_by_string":  "test user",
			"amount":             "10.50",
			"charged_to_user_id": 1,
			"is_taxed":           true,
			"name":               "Test Item",
			"receipt_id":         1,
			"status":             "OPEN",
		}
		db.Table("items").Create(item)
	}
	
	itemsCountBefore := countRowsInTable("items")
	
	err := MigrateItemsToShares()
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}
	
	sharesCountAfter := countRowsInTable("shares")
	
	if sharesCountAfter != itemsCountBefore {
		utils.PrintTestError(t, sharesCountAfter, itemsCountBefore)
	}
	
	if sharesCountAfter != 2500 {
		utils.PrintTestError(t, sharesCountAfter, 2500)
	}
}