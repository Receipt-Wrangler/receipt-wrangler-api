package repositories

import (
	"github.com/shopspring/decimal"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"testing"
	"time"
)

func setupReceiptTest() {
	CreateTestGroupWithUsers()
	CreateTestCategories()
	createTestTags()
}

func createTestTags() {
	db := GetDB()
	tag1 := models.Tag{
		Name: "test-tag-1",
	}
	tag2 := models.Tag{
		Name: "test-tag-2",
	}
	db.Create(&tag1)
	db.Create(&tag2)
}

func createTestReceipts() {
	db := GetDB()

	// Create receipt with ID 1
	receipt1 := models.Receipt{
		Name:         "Test Receipt 1",
		Amount:       decimal.NewFromFloat(100.00),
		Date:         time.Now(),
		PaidByUserID: 1,
		Status:       models.OPEN,
		GroupId:      1,
	}
	db.Create(&receipt1)

	// Create receipt with ID 2
	receipt2 := models.Receipt{
		Name:         "Test Receipt 2",
		Amount:       decimal.NewFromFloat(200.00),
		Date:         time.Now(),
		PaidByUserID: 2,
		Status:       models.RESOLVED,
		GroupId:      1,
	}
	db.Create(&receipt2)
}

func teardownReceiptTest() {
	TruncateTestDb()
}

func TestShouldCreateReceipt(t *testing.T) {
	defer teardownReceiptTest()
	setupReceiptTest()

	repository := NewReceiptRepository(nil)

	// Create command for new receipt
	command := commands.UpsertReceiptCommand{
		Name:         "New Receipt",
		Amount:       decimal.NewFromFloat(50.75),
		Date:         time.Now(),
		PaidByUserID: 1,
		Status:       models.OPEN,
		GroupId:      1,
		Categories: []commands.UpsertCategoryCommand{
			{
				Id: uintPtr(1),
			},
		},
		Tags: []commands.UpsertTagCommand{
			{
				Id: uintPtr(1),
			},
		},
	}

	createdReceipt, err := repository.CreateReceipt(command, 1, true)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Validate created receipt
	if createdReceipt.ID == 0 {
		utils.PrintTestError(t, "Receipt ID should not be 0", nil)
	}

	if createdReceipt.Name != "New Receipt" {
		utils.PrintTestError(t, createdReceipt.Name, "New Receipt")
	}

	if !createdReceipt.Amount.Equal(decimal.NewFromFloat(50.75)) {
		utils.PrintTestError(t, createdReceipt.Amount, decimal.NewFromFloat(50.75))
	}

	if createdReceipt.PaidByUserID != 1 {
		utils.PrintTestError(t, createdReceipt.PaidByUserID, 1)
	}

	if len(createdReceipt.Categories) != 1 {
		utils.PrintTestError(t, len(createdReceipt.Categories), 1)
	}

	if len(createdReceipt.Tags) != 1 {
		utils.PrintTestError(t, len(createdReceipt.Tags), 1)
	}

	if createdReceipt.Status != models.OPEN {
		utils.PrintTestError(t, createdReceipt.Status, models.OPEN)
	}
}

func TestShouldGetReceiptById(t *testing.T) {
	defer teardownReceiptTest()
	setupReceiptTest()
	createTestReceipts()

	repository := NewReceiptRepository(nil)

	receipt, err := repository.GetReceiptById("1")
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if receipt.ID != 1 {
		utils.PrintTestError(t, receipt.ID, 1)
	}

	if receipt.Name != "Test Receipt 1" {
		utils.PrintTestError(t, receipt.Name, "Test Receipt 1")
	}

	if !receipt.Amount.Equal(decimal.NewFromFloat(100.00)) {
		utils.PrintTestError(t, receipt.Amount, decimal.NewFromFloat(100.00))
	}
}

func TestShouldGetFullyLoadedReceiptById(t *testing.T) {
	defer teardownReceiptTest()
	setupReceiptTest()
	createTestReceipts()

	repository := NewReceiptRepository(nil)

	receipt, err := repository.GetFullyLoadedReceiptById("1")
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if receipt.ID != 1 {
		utils.PrintTestError(t, receipt.ID, 1)
	}
}

func TestShouldGetReceiptGroupIdByReceiptId(t *testing.T) {
	defer teardownReceiptTest()
	setupReceiptTest()
	createTestReceipts()

	repository := NewReceiptRepository(nil)

	groupId, err := repository.GetReceiptGroupIdByReceiptId("1")
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if groupId != 1 {
		utils.PrintTestError(t, groupId, 1)
	}
}

func TestShouldUpdateReceipt(t *testing.T) {
	defer teardownReceiptTest()
	setupReceiptTest()
	createTestReceipts()

	repository := NewReceiptRepository(nil)

	// Create command for updating receipt
	command := commands.UpsertReceiptCommand{
		Name:         "Updated Receipt",
		Amount:       decimal.NewFromFloat(150.25),
		Date:         time.Now(),
		PaidByUserID: 2,
		Status:       models.NEEDS_ATTENTION,
		GroupId:      1,
		Categories: []commands.UpsertCategoryCommand{
			{
				Id: uintPtr(1),
			},
			{
				Id: uintPtr(2),
			},
		},
		Tags: []commands.UpsertTagCommand{
			{
				Id: uintPtr(1),
			},
		},
	}

	updatedReceipt, err := repository.UpdateReceipt("1", command, 1)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Validate updated receipt
	if updatedReceipt.Name != "Updated Receipt" {
		utils.PrintTestError(t, updatedReceipt.Name, "Updated Receipt")
	}

	if !updatedReceipt.Amount.Equal(decimal.NewFromFloat(150.25)) {
		utils.PrintTestError(t, updatedReceipt.Amount, decimal.NewFromFloat(150.25))
	}

	if updatedReceipt.PaidByUserID != 2 {
		utils.PrintTestError(t, updatedReceipt.PaidByUserID, 2)
	}

	if updatedReceipt.Status != models.NEEDS_ATTENTION {
		utils.PrintTestError(t, updatedReceipt.Status, models.NEEDS_ATTENTION)
	}

	if len(updatedReceipt.Categories) != 2 {
		utils.PrintTestError(t, len(updatedReceipt.Categories), 2)
	}
}

func TestShouldUpdateItemsToStatus(t *testing.T) {
	defer teardownReceiptTest()
	setupReceiptTest()
	createTestReceipts()
	createTestItems()

	repository := NewReceiptRepository(nil)

	db := GetDB()
	var receipt models.Receipt
	db.First(&receipt, 1)

	err := repository.UpdateItemsToStatus(&receipt, models.ITEM_RESOLVED)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Verify items have been updated
	var items []models.Item
	db.Where("receipt_id = ?", receipt.ID).Find(&items)

	for _, item := range items {
		if item.Status != models.ITEM_RESOLVED {
			utils.PrintTestError(t, item.Status, models.ITEM_RESOLVED)
		}
	}
}

func TestShouldGetPagedReceiptsByGroupId(t *testing.T) {
	defer teardownReceiptTest()
	setupReceiptTest()
	createTestReceipts()

	repository := NewReceiptRepository(nil)

	// Create paged request command
	pagedRequest := commands.ReceiptPagedRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:          1,
			PageSize:      10,
			OrderBy:       "created_at",
			SortDirection: commands.DESCENDING,
		},
	}

	receipts, count, err := repository.GetPagedReceiptsByGroupId(1, "1", pagedRequest, nil)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if count != 2 {
		utils.PrintTestError(t, count, 2)
	}

	if len(receipts) != 2 {
		utils.PrintTestError(t, len(receipts), 2)
	}
}

func TestShouldBuildGormFilterQuery(t *testing.T) {
	defer teardownReceiptTest()
	setupReceiptTest()

	repository := NewReceiptRepository(nil)

	// Test simple filter
	pagedRequest := commands.ReceiptPagedRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			OrderBy: "created_at",
		},
		Filter: commands.ReceiptPagedRequestFilter{
			Name: commands.PagedRequestField{
				Operation: commands.CONTAINS,
				Value:     "test",
			},
		},
	}

	query, err := repository.BuildGormFilterQuery(pagedRequest)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if query == nil {
		utils.PrintTestError(t, "Query should not be nil", nil)
	}
}

// Helper functions
func createTestItems() {
	db := GetDB()

	// Create items for receipt 1
	item1 := models.Item{
		Name:            "Item 1",
		Amount:          decimal.NewFromFloat(50.00),
		ReceiptId:       1,
		ChargedToUserId: 2,
		Status:          models.ITEM_OPEN,
	}
	db.Create(&item1)

	item2 := models.Item{
		Name:            "Item 2",
		Amount:          decimal.NewFromFloat(50.00),
		ReceiptId:       1,
		ChargedToUserId: 3,
		Status:          models.ITEM_OPEN,
	}
	db.Create(&item2)
}

func uintPtr(v uint) *uint {
	return &v
}
