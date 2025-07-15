package services

import (
	"fmt"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"
	"strconv"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
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

func createTestItemsWithEdgeCases() {
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
		{
			"id":                 3,
			"created_at":         "2024-12-31 23:59:59",
			"updated_at":         "2024-12-31 23:59:59",
			"created_by":         1,
			"created_by_string":  "",
			"amount":             "0.001",
			"charged_to_user_id": 1,
			"is_taxed":           true,
			"name":               "Edge Case Item - Empty String",
			"receipt_id":         1,
			"status":             "OPEN",
		},
		{
			"id":                 4,
			"created_at":         "2024-06-15 12:30:45",
			"updated_at":         "2024-06-15 12:30:45",
			"created_by":         1,
			"created_by_string":  "Special chars: !@#$%^&*()_+{}|:<>?[]\\;',./",
			"amount":             "999.999",
			"charged_to_user_id": 1,
			"is_taxed":           false,
			"name":               "Item with special chars: !@#$%^&*()_+{}|:<>?[]\\;',./",
			"receipt_id":         1,
			"status":             "RESOLVED",
		},
		{
			"id":                 5,
			"created_at":         "2024-02-29 00:00:00",
			"updated_at":         "2024-02-29 00:00:00",
			"created_by":         1,
			"created_by_string":  "leap year test",
			"amount":             "1000000.999",
			"charged_to_user_id": 1,
			"is_taxed":           true,
			"name":               "Leap Year Test Item",
			"receipt_id":         1,
			"status":             "OPEN",
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

func createTestItemCategoriesWithEdgeCases() {
	db := repositories.GetDB()

	itemCategories := []ItemCategoryData{
		{ReceiptId: "1", CategoryId: "1"},
		{ReceiptId: "1", CategoryId: "2"},
		{ReceiptId: "2", CategoryId: "1"},
		{ReceiptId: "3", CategoryId: "1"},
		{ReceiptId: "3", CategoryId: "2"},
		{ReceiptId: "4", CategoryId: "1"},
		{ReceiptId: "5", CategoryId: "2"},
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

func createTestItemTagsWithEdgeCases() {
	db := repositories.GetDB()

	itemTags := []ItemTagData{
		{ReceiptId: "1", TagId: "1"},
		{ReceiptId: "1", TagId: "2"},
		{ReceiptId: "2", TagId: "3"},
		{ReceiptId: "3", TagId: "1"},
		{ReceiptId: "4", TagId: "2"},
		{ReceiptId: "4", TagId: "3"},
		{ReceiptId: "5", TagId: "1"},
		{ReceiptId: "5", TagId: "2"},
		{ReceiptId: "5", TagId: "3"},
	}

	db.Table("item_tags").Create(&itemTags)
}

func createTestReceipts() {
	db := repositories.GetDB()
	receipt1 := models.Receipt{
		BaseModel:    models.BaseModel{ID: 1},
		Name:         "Test Receipt 1",
		Amount:       decimal.NewFromFloat(100.00),
		Date:         time.Now(),
		PaidByUserID: 1,
		Status:       models.OPEN,
		GroupId:      1,
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

func validateSharesMatchItems(t *testing.T, db *gorm.DB) {
	var items []map[string]interface{}
	var shares []models.Share

	err := db.Table("items").Find(&items).Error
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	err = db.Table("shares").Find(&shares).Error
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if len(items) != len(shares) {
		utils.PrintTestError(t, "Record count mismatch", map[string]interface{}{
			"items_count":  len(items),
			"shares_count": len(shares),
		})
		return
	}

	for _, item := range items {
		var matchingShare *models.Share
		for i := range shares {
			if shares[i].ID == uint(item["id"].(int64)) {
				matchingShare = &shares[i]
				break
			}
		}

		if matchingShare == nil {
			utils.PrintTestError(t, "Share not found for item ID", item["id"])
			continue
		}

		compareItemToShare(t, item, *matchingShare)
	}
}

func compareItemToShare(t *testing.T, item map[string]interface{}, share models.Share) {
	itemID := uint(item["id"].(int64))

	if itemID != share.ID {
		utils.PrintTestError(t, "ID mismatch for item", map[string]interface{}{
			"item_id":  itemID,
			"share_id": share.ID,
		})
	}

	if item["name"].(string) != share.Name {
		utils.PrintTestError(t, "Name mismatch for item", map[string]interface{}{
			"item_id":    itemID,
			"item_name":  item["name"].(string),
			"share_name": share.Name,
		})
	}

	originalAmount, err := decimal.NewFromString(item["amount"].(string))
	if err != nil {
		utils.PrintTestError(t, "Invalid amount in item", map[string]interface{}{
			"item_id": itemID,
			"amount":  item["amount"].(string),
		})
		return
	}
	if !originalAmount.Equal(share.Amount) {
		utils.PrintTestError(t, "Amount mismatch for item", map[string]interface{}{
			"item_id":      itemID,
			"item_amount":  originalAmount,
			"share_amount": share.Amount,
		})
	}

	if uint(item["charged_to_user_id"].(int64)) != share.ChargedToUserId {
		utils.PrintTestError(t, "ChargedToUserId mismatch for item", map[string]interface{}{
			"item_id":                  itemID,
			"item_charged_to_user_id":  item["charged_to_user_id"].(int64),
			"share_charged_to_user_id": share.ChargedToUserId,
		})
	}

	var itemIsTaxed bool
	switch v := item["is_taxed"].(type) {
	case bool:
		itemIsTaxed = v
	case int64:
		itemIsTaxed = v != 0
	default:
		utils.PrintTestError(t, "Invalid is_taxed type in item", map[string]interface{}{
			"item_id": itemID,
			"type":    fmt.Sprintf("%T", v),
		})
		return
	}

	if itemIsTaxed != share.IsTaxed {
		utils.PrintTestError(t, "IsTaxed mismatch for item", map[string]interface{}{
			"item_id":        itemID,
			"item_is_taxed":  itemIsTaxed,
			"share_is_taxed": share.IsTaxed,
		})
	}

	if uint(item["receipt_id"].(int64)) != share.ReceiptId {
		utils.PrintTestError(t, "ReceiptId mismatch for item", map[string]interface{}{
			"item_id":          itemID,
			"item_receipt_id":  item["receipt_id"].(int64),
			"share_receipt_id": share.ReceiptId,
		})
	}

	if models.ShareStatus(item["status"].(string)) != share.Status {
		utils.PrintTestError(t, "Status mismatch for item", map[string]interface{}{
			"item_id":      itemID,
			"item_status":  item["status"].(string),
			"share_status": share.Status,
		})
	}

	expectedCreatedAt, err := time.Parse("2006-01-02 15:04:05", item["created_at"].(string))
	if err != nil {
		utils.PrintTestError(t, "Invalid created_at in item", map[string]interface{}{
			"item_id":    itemID,
			"created_at": item["created_at"].(string),
		})
		return
	}
	if !expectedCreatedAt.Equal(share.CreatedAt) {
		utils.PrintTestError(t, "CreatedAt mismatch for item", map[string]interface{}{
			"item_id":          itemID,
			"item_created_at":  expectedCreatedAt,
			"share_created_at": share.CreatedAt,
		})
	}

	expectedUpdatedAt, err := time.Parse("2006-01-02 15:04:05", item["updated_at"].(string))
	if err != nil {
		utils.PrintTestError(t, "Invalid updated_at in item", map[string]interface{}{
			"item_id":    itemID,
			"updated_at": item["updated_at"].(string),
		})
		return
	}
	if !expectedUpdatedAt.Equal(share.UpdatedAt) {
		utils.PrintTestError(t, "UpdatedAt mismatch for item", map[string]interface{}{
			"item_id":          itemID,
			"item_updated_at":  expectedUpdatedAt,
			"share_updated_at": share.UpdatedAt,
		})
	}

	expectedCreatedBy := uint(item["created_by"].(int64))
	if share.CreatedBy == nil || *share.CreatedBy != expectedCreatedBy {
		utils.PrintTestError(t, "CreatedBy mismatch for item", map[string]interface{}{
			"item_id":          itemID,
			"item_created_by":  expectedCreatedBy,
			"share_created_by": share.CreatedBy,
		})
	}

	if item["created_by_string"].(string) != share.CreatedByString {
		utils.PrintTestError(t, "CreatedByString mismatch for item", map[string]interface{}{
			"item_id":                 itemID,
			"item_created_by_string":  item["created_by_string"].(string),
			"share_created_by_string": share.CreatedByString,
		})
	}
}

func validateShareCategoriesMatchItemCategories(t *testing.T, db *gorm.DB) {
	var itemCategories []ItemCategoryData
	var shareCategories []ShareCategoryData

	err := db.Table("item_categories").Find(&itemCategories).Error
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	err = db.Table("share_categories").Find(&shareCategories).Error
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if len(itemCategories) != len(shareCategories) {
		utils.PrintTestError(t, "Category association count mismatch", map[string]interface{}{
			"item_categories_count":  len(itemCategories),
			"share_categories_count": len(shareCategories),
		})
		return
	}

	for _, itemCategory := range itemCategories {
		shareId, _ := strconv.ParseUint(itemCategory.ReceiptId, 10, 32)
		categoryId, _ := strconv.ParseUint(itemCategory.CategoryId, 10, 32)

		found := false
		for _, shareCategory := range shareCategories {
			if shareCategory.ShareId == uint(shareId) && shareCategory.CategoryId == uint(categoryId) {
				found = true
				break
			}
		}

		if !found {
			utils.PrintTestError(t, "Category association not found", map[string]interface{}{
				"share_id":    shareId,
				"category_id": categoryId,
			})
		}
	}
}

func validateShareTagsMatchItemTags(t *testing.T, db *gorm.DB) {
	var itemTags []ItemTagData
	var shareTags []ShareTagData

	err := db.Table("item_tags").Find(&itemTags).Error
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	err = db.Table("share_tags").Find(&shareTags).Error
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if len(itemTags) != len(shareTags) {
		utils.PrintTestError(t, "Tag association count mismatch", map[string]interface{}{
			"item_tags_count":  len(itemTags),
			"share_tags_count": len(shareTags),
		})
		return
	}

	for _, itemTag := range itemTags {
		shareId, _ := strconv.ParseUint(itemTag.ReceiptId, 10, 32)
		tagId, _ := strconv.ParseUint(itemTag.TagId, 10, 32)

		found := false
		for _, shareTag := range shareTags {
			if shareTag.ShareId == uint(shareId) && shareTag.TagId == uint(tagId) {
				found = true
				break
			}
		}

		if !found {
			utils.PrintTestError(t, "Tag association not found", map[string]interface{}{
				"share_id": shareId,
				"tag_id":   tagId,
			})
		}
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
	validateSharesMatchItems(t, db)
}

func TestMigrateItemsToShares_CompleteDataIntegrity(t *testing.T) {
	defer repositories.TruncateTestDb()

	repositories.CreateTestGroupWithUsers()
	repositories.CreateTestCategories()
	createTestTags()
	createTestReceipts()
	createOldTables()
	createTestItems()
	createTestItemCategories()
	createTestItemTags()

	err := MigrateItemsToShares()
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	db := repositories.GetDB()
	validateSharesMatchItems(t, db)
	validateShareCategoriesMatchItemCategories(t, db)
	validateShareTagsMatchItemTags(t, db)
}

func TestMigrateItemsToShares_AssociationIntegrity(t *testing.T) {
	defer repositories.TruncateTestDb()

	repositories.CreateTestGroupWithUsers()
	repositories.CreateTestCategories()
	createTestTags()
	createTestReceipts()
	createOldTables()
	createTestItems()
	createTestItemCategories()
	createTestItemTags()

	err := MigrateItemsToShares()
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	db := repositories.GetDB()

	var itemCategoriesCount int64
	var shareCategoriesCount int64
	var itemTagsCount int64
	var shareTagsCount int64

	db.Table("item_categories").Count(&itemCategoriesCount)
	db.Table("share_categories").Count(&shareCategoriesCount)
	db.Table("item_tags").Count(&itemTagsCount)
	db.Table("share_tags").Count(&shareTagsCount)

	if itemCategoriesCount != shareCategoriesCount {
		utils.PrintTestError(t, "Category association count mismatch", map[string]interface{}{
			"item_categories_count":  itemCategoriesCount,
			"share_categories_count": shareCategoriesCount,
		})
	}

	if itemTagsCount != shareTagsCount {
		utils.PrintTestError(t, "Tag association count mismatch", map[string]interface{}{
			"item_tags_count":  itemTagsCount,
			"share_tags_count": shareTagsCount,
		})
	}

	validateShareCategoriesMatchItemCategories(t, db)
	validateShareTagsMatchItemTags(t, db)
}

func TestMigrateItemsToShares_EdgeCases(t *testing.T) {
	defer repositories.TruncateTestDb()

	repositories.CreateTestGroupWithUsers()
	repositories.CreateTestCategories()
	createTestTags()
	createTestReceipts()
	createOldTables()
	createTestItemsWithEdgeCases()
	createTestItemCategoriesWithEdgeCases()
	createTestItemTagsWithEdgeCases()

	err := MigrateItemsToShares()
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	db := repositories.GetDB()
	validateSharesMatchItems(t, db)
	validateShareCategoriesMatchItemCategories(t, db)
	validateShareTagsMatchItemTags(t, db)

	var shares []models.Share
	db.Find(&shares)

	if len(shares) != 5 {
		utils.PrintTestError(t, "Expected 5 shares for edge case test", len(shares))
		return
	}

	for _, share := range shares {
		switch share.ID {
		case 3:
			if share.CreatedByString != "" {
				utils.PrintTestError(t, "Empty string not preserved", share.CreatedByString)
			}
			if share.Amount.String() != "0.001" {
				utils.PrintTestError(t, "Small decimal not preserved", share.Amount.String())
			}
		case 4:
			expectedSpecialChars := "Special chars: !@#$%^&*()_+{}|:<>?[]\\;',./"
			if share.CreatedByString != expectedSpecialChars {
				utils.PrintTestError(t, "Special characters not preserved", map[string]interface{}{
					"expected": expectedSpecialChars,
					"actual":   share.CreatedByString,
				})
			}
			if share.Amount.String() != "999.999" {
				utils.PrintTestError(t, "Large decimal not preserved", share.Amount.String())
			}
		case 5:
			expectedTime, _ := time.Parse("2006-01-02 15:04:05", "2024-02-29 00:00:00")
			if !share.CreatedAt.Equal(expectedTime) {
				utils.PrintTestError(t, "Leap year date not preserved", map[string]interface{}{
					"expected": expectedTime,
					"actual":   share.CreatedAt,
				})
			}
			if share.Amount.String() != "1000000.999" {
				utils.PrintTestError(t, "Very large decimal not preserved", share.Amount.String())
			}
		}
	}

	// Verify specific association counts for edge case test
	var shareCategoriesCount int64
	var shareTagsCount int64
	db.Table("share_categories").Count(&shareCategoriesCount)
	db.Table("share_tags").Count(&shareTagsCount)

	if shareCategoriesCount != 7 {
		utils.PrintTestError(t, "Expected 7 share categories for edge case test", shareCategoriesCount)
	}

	if shareTagsCount != 9 {
		utils.PrintTestError(t, "Expected 9 share tags for edge case test", shareTagsCount)
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
