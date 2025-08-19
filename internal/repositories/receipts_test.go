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

	var two = uint(2)
	var three = uint(3)

	// Create items for receipt 1
	item1 := models.Item{
		Name:            "Item 1",
		Amount:          decimal.NewFromFloat(50.00),
		ReceiptId:       1,
		ChargedToUserId: &two,
		Status:          models.ITEM_OPEN,
	}
	db.Create(&item1)

	item2 := models.Item{
		Name:            "Item 2",
		Amount:          decimal.NewFromFloat(50.00),
		ReceiptId:       1,
		ChargedToUserId: &three,
		Status:          models.ITEM_OPEN,
	}
	db.Create(&item2)
}

func uintPtr(v uint) *uint {
	return &v
}

func TestShouldCreateReceiptWithLinkedItems(t *testing.T) {
	defer teardownReceiptTest()
	setupReceiptTest()

	repository := NewReceiptRepository(nil)

	// Create command for new receipt with items that have linked items
	command := commands.UpsertReceiptCommand{
		Name:         "Receipt with Linked Items",
		Amount:       decimal.NewFromFloat(100.00),
		Date:         time.Now(),
		PaidByUserID: 1,
		Status:       models.OPEN,
		GroupId:      1,
		Items: []commands.UpsertItemCommand{
			{
				Name:            "Main Item 1",
				Amount:          decimal.NewFromFloat(50.00),
				ChargedToUserId: uintPtr(2),
				Status:          models.ITEM_OPEN,
				Categories: []commands.UpsertCategoryCommand{
					{Id: uintPtr(1)},
				},
				LinkedItems: []commands.UpsertItemCommand{
					{
						Name:            "Linked Item 1",
						Amount:          decimal.NewFromFloat(10.00),
						ChargedToUserId: uintPtr(3),
						Status:          models.ITEM_OPEN,
						Categories: []commands.UpsertCategoryCommand{
							{Id: uintPtr(2)},
						},
						Tags: []commands.UpsertTagCommand{
							{Id: uintPtr(1)},
						},
					},
					{
						Name:            "Linked Item 2",
						Amount:          decimal.NewFromFloat(15.00),
						ChargedToUserId: uintPtr(2),
						Status:          models.ITEM_OPEN,
					},
				},
			},
			{
				Name:            "Main Item 2",
				Amount:          decimal.NewFromFloat(50.00),
				ChargedToUserId: uintPtr(3),
				Status:          models.ITEM_OPEN,
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

	// Get fully loaded receipt to check linked items
	fullyLoadedReceipt, err := repository.GetFullyLoadedReceiptById(utils.UintToString(createdReceipt.ID))
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Check that we have 2 main items (linked items should be filtered out)
	if len(fullyLoadedReceipt.ReceiptItems) != 2 {
		utils.PrintTestError(t, len(fullyLoadedReceipt.ReceiptItems), 2)
	}

	// Find the first item with linked items
	var mainItem1 *models.Item
	for i := range fullyLoadedReceipt.ReceiptItems {
		if fullyLoadedReceipt.ReceiptItems[i].Name == "Main Item 1" {
			mainItem1 = &fullyLoadedReceipt.ReceiptItems[i]
			break
		}
	}

	if mainItem1 == nil {
		utils.PrintTestError(t, "Main Item 1 not found", nil)
		return
	}

	// Check that Main Item 1 has 2 linked items
	if len(mainItem1.LinkedItems) != 2 {
		utils.PrintTestError(t, len(mainItem1.LinkedItems), 2)
	}

	// Verify linked items have correct receipt ID
	for _, linkedItem := range mainItem1.LinkedItems {
		if linkedItem.ReceiptId != createdReceipt.ID {
			utils.PrintTestError(t, linkedItem.ReceiptId, createdReceipt.ID)
		}
	}

	// Check that linked item categories and tags are preserved
	var linkedItem1 *models.Item
	for i := range mainItem1.LinkedItems {
		if mainItem1.LinkedItems[i].Name == "Linked Item 1" {
			linkedItem1 = &mainItem1.LinkedItems[i]
			break
		}
	}

	if linkedItem1 != nil {
		if len(linkedItem1.Categories) != 1 {
			utils.PrintTestError(t, len(linkedItem1.Categories), 1)
		}
		if len(linkedItem1.Tags) != 1 {
			utils.PrintTestError(t, len(linkedItem1.Tags), 1)
		}
	}
}

func TestShouldCreateReceiptWithMultipleItemsHavingLinkedItems(t *testing.T) {
	defer teardownReceiptTest()
	setupReceiptTest()

	repository := NewReceiptRepository(nil)

	// Create command with multiple items having linked items
	command := commands.UpsertReceiptCommand{
		Name:         "Receipt with Multiple Linked Items",
		Amount:       decimal.NewFromFloat(200.00),
		Date:         time.Now(),
		PaidByUserID: 1,
		Status:       models.OPEN,
		GroupId:      1,
		Items: []commands.UpsertItemCommand{
			{
				Name:            "Item A",
				Amount:          decimal.NewFromFloat(100.00),
				ChargedToUserId: uintPtr(2),
				Status:          models.ITEM_OPEN,
				LinkedItems: []commands.UpsertItemCommand{
					{
						Name:            "Item A - Linked 1",
						Amount:          decimal.NewFromFloat(30.00),
						ChargedToUserId: uintPtr(3),
						Status:          models.ITEM_OPEN,
					},
					{
						Name:            "Item A - Linked 2",
						Amount:          decimal.NewFromFloat(20.00),
						ChargedToUserId: uintPtr(2),
						Status:          models.ITEM_OPEN,
					},
				},
			},
			{
				Name:            "Item B",
				Amount:          decimal.NewFromFloat(100.00),
				ChargedToUserId: uintPtr(3),
				Status:          models.ITEM_OPEN,
				LinkedItems: []commands.UpsertItemCommand{
					{
						Name:            "Item B - Linked 1",
						Amount:          decimal.NewFromFloat(50.00),
						ChargedToUserId: uintPtr(2),
						Status:          models.ITEM_OPEN,
					},
				},
			},
		},
	}

	createdReceipt, err := repository.CreateReceipt(command, 1, true)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Get fully loaded receipt
	fullyLoadedReceipt, err := repository.GetFullyLoadedReceiptById(utils.UintToString(createdReceipt.ID))
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Should have 2 main items
	if len(fullyLoadedReceipt.ReceiptItems) != 2 {
		utils.PrintTestError(t, len(fullyLoadedReceipt.ReceiptItems), 2)
	}

	// Count total linked items
	totalLinkedItems := 0
	for _, item := range fullyLoadedReceipt.ReceiptItems {
		totalLinkedItems += len(item.LinkedItems)
	}

	// Should have 3 linked items total (2 for Item A, 1 for Item B)
	if totalLinkedItems != 3 {
		utils.PrintTestError(t, totalLinkedItems, 3)
	}

	// Verify each item has correct number of linked items
	for _, item := range fullyLoadedReceipt.ReceiptItems {
		if item.Name == "Item A" && len(item.LinkedItems) != 2 {
			utils.PrintTestError(t, len(item.LinkedItems), 2)
		}
		if item.Name == "Item B" && len(item.LinkedItems) != 1 {
			utils.PrintTestError(t, len(item.LinkedItems), 1)
		}
	}
}

func TestShouldCreateReceiptWithNestedLinkedItemsCategories(t *testing.T) {
	defer teardownReceiptTest()
	setupReceiptTest()

	repository := NewReceiptRepository(nil)

	// Create command with linked items having categories
	command := commands.UpsertReceiptCommand{
		Name:         "Receipt with Categorized Linked Items",
		Amount:       decimal.NewFromFloat(75.00),
		Date:         time.Now(),
		PaidByUserID: 1,
		Status:       models.OPEN,
		GroupId:      1,
		Items: []commands.UpsertItemCommand{
			{
				Name:            "Main Item with Categorized Links",
				Amount:          decimal.NewFromFloat(75.00),
				ChargedToUserId: uintPtr(2),
				Status:          models.ITEM_OPEN,
				Categories: []commands.UpsertCategoryCommand{
					{Id: uintPtr(1)},
				},
				LinkedItems: []commands.UpsertItemCommand{
					{
						Name:            "Linked with Category 1",
						Amount:          decimal.NewFromFloat(25.00),
						ChargedToUserId: uintPtr(3),
						Status:          models.ITEM_OPEN,
						Categories: []commands.UpsertCategoryCommand{
							{Id: uintPtr(1)},
						},
					},
					{
						Name:            "Linked with Category 2",
						Amount:          decimal.NewFromFloat(25.00),
						ChargedToUserId: uintPtr(2),
						Status:          models.ITEM_OPEN,
						Categories: []commands.UpsertCategoryCommand{
							{Id: uintPtr(2)},
						},
					},
					{
						Name:            "Linked with Both Categories",
						Amount:          decimal.NewFromFloat(25.00),
						ChargedToUserId: uintPtr(3),
						Status:          models.ITEM_OPEN,
						Categories: []commands.UpsertCategoryCommand{
							{Id: uintPtr(1)},
							{Id: uintPtr(2)},
						},
					},
				},
			},
		},
	}

	createdReceipt, err := repository.CreateReceipt(command, 1, true)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Get fully loaded receipt
	fullyLoadedReceipt, err := repository.GetFullyLoadedReceiptById(utils.UintToString(createdReceipt.ID))
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Find the main item
	if len(fullyLoadedReceipt.ReceiptItems) != 1 {
		utils.PrintTestError(t, len(fullyLoadedReceipt.ReceiptItems), 1)
		return
	}

	mainItem := fullyLoadedReceipt.ReceiptItems[0]

	// Check linked items categories
	for _, linkedItem := range mainItem.LinkedItems {
		switch linkedItem.Name {
		case "Linked with Category 1":
			if len(linkedItem.Categories) != 1 {
				utils.PrintTestError(t, len(linkedItem.Categories), 1)
			}
		case "Linked with Category 2":
			if len(linkedItem.Categories) != 1 {
				utils.PrintTestError(t, len(linkedItem.Categories), 1)
			}
		case "Linked with Both Categories":
			if len(linkedItem.Categories) != 2 {
				utils.PrintTestError(t, len(linkedItem.Categories), 2)
			}
		}
	}
}

func TestShouldCreateReceiptWithNestedLinkedItemsTags(t *testing.T) {
	defer teardownReceiptTest()
	setupReceiptTest()

	repository := NewReceiptRepository(nil)

	// Create command with linked items having tags
	command := commands.UpsertReceiptCommand{
		Name:         "Receipt with Tagged Linked Items",
		Amount:       decimal.NewFromFloat(60.00),
		Date:         time.Now(),
		PaidByUserID: 1,
		Status:       models.OPEN,
		GroupId:      1,
		Items: []commands.UpsertItemCommand{
			{
				Name:            "Main Item with Tagged Links",
				Amount:          decimal.NewFromFloat(60.00),
				ChargedToUserId: uintPtr(2),
				Status:          models.ITEM_OPEN,
				Tags: []commands.UpsertTagCommand{
					{Id: uintPtr(1)},
				},
				LinkedItems: []commands.UpsertItemCommand{
					{
						Name:            "Linked with Tag 1",
						Amount:          decimal.NewFromFloat(20.00),
						ChargedToUserId: uintPtr(3),
						Status:          models.ITEM_OPEN,
						Tags: []commands.UpsertTagCommand{
							{Id: uintPtr(1)},
						},
					},
					{
						Name:            "Linked with Tag 2",
						Amount:          decimal.NewFromFloat(20.00),
						ChargedToUserId: uintPtr(2),
						Status:          models.ITEM_OPEN,
						Tags: []commands.UpsertTagCommand{
							{Id: uintPtr(2)},
						},
					},
					{
						Name:            "Linked with Both Tags",
						Amount:          decimal.NewFromFloat(20.00),
						ChargedToUserId: uintPtr(3),
						Status:          models.ITEM_OPEN,
						Tags: []commands.UpsertTagCommand{
							{Id: uintPtr(1)},
							{Id: uintPtr(2)},
						},
					},
				},
			},
		},
	}

	createdReceipt, err := repository.CreateReceipt(command, 1, true)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Get fully loaded receipt
	fullyLoadedReceipt, err := repository.GetFullyLoadedReceiptById(utils.UintToString(createdReceipt.ID))
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Find the main item
	if len(fullyLoadedReceipt.ReceiptItems) != 1 {
		utils.PrintTestError(t, len(fullyLoadedReceipt.ReceiptItems), 1)
		return
	}

	mainItem := fullyLoadedReceipt.ReceiptItems[0]

	// Check linked items tags
	for _, linkedItem := range mainItem.LinkedItems {
		switch linkedItem.Name {
		case "Linked with Tag 1":
			if len(linkedItem.Tags) != 1 {
				utils.PrintTestError(t, len(linkedItem.Tags), 1)
			}
		case "Linked with Tag 2":
			if len(linkedItem.Tags) != 1 {
				utils.PrintTestError(t, len(linkedItem.Tags), 1)
			}
		case "Linked with Both Tags":
			if len(linkedItem.Tags) != 2 {
				utils.PrintTestError(t, len(linkedItem.Tags), 2)
			}
		}
	}
}

func TestShouldFilterLinkedItemsFromReceiptItems(t *testing.T) {
	defer teardownReceiptTest()
	setupReceiptTest()

	repository := NewReceiptRepository(nil)
	db := GetDB()

	// Create a receipt directly in database
	receipt := models.Receipt{
		Name:         "Test Receipt",
		Amount:       decimal.NewFromFloat(100.00),
		Date:         time.Now(),
		PaidByUserID: 1,
		Status:       models.OPEN,
		GroupId:      1,
	}
	db.Create(&receipt)

	// Create items
	item1 := models.Item{
		Name:      "Item 1",
		Amount:    decimal.NewFromFloat(50.00),
		ReceiptId: receipt.ID,
		Status:    models.ITEM_OPEN,
	}
	db.Create(&item1)

	item2 := models.Item{
		Name:      "Item 2",
		Amount:    decimal.NewFromFloat(30.00),
		ReceiptId: receipt.ID,
		Status:    models.ITEM_OPEN,
	}
	db.Create(&item2)

	linkedItem := models.Item{
		Name:      "Linked Item",
		Amount:    decimal.NewFromFloat(20.00),
		ReceiptId: receipt.ID,
		Status:    models.ITEM_OPEN,
	}
	db.Create(&linkedItem)

	// Create the association
	db.Model(&item1).Association("LinkedItems").Append(&linkedItem)

	// Load receipt with all items
	var loadedReceipt models.Receipt
	db.Model(models.Receipt{}).
		Where("id = ?", receipt.ID).
		Preload("ReceiptItems").
		Preload("ReceiptItems.LinkedItems").
		Find(&loadedReceipt)

	// Before filtering, should have 3 items total
	if len(loadedReceipt.ReceiptItems) != 3 {
		utils.PrintTestError(t, len(loadedReceipt.ReceiptItems), 3)
	}

	// Apply filter
	repository.FilterLinkedItemsFromReceiptItems(&loadedReceipt)

	// After filtering, should have 2 items (linked item removed from main list)
	if len(loadedReceipt.ReceiptItems) != 2 {
		utils.PrintTestError(t, len(loadedReceipt.ReceiptItems), 2)
	}

	// Check that Item 1 still has its linked item
	for _, item := range loadedReceipt.ReceiptItems {
		if item.Name == "Item 1" {
			if len(item.LinkedItems) != 1 {
				utils.PrintTestError(t, len(item.LinkedItems), 1)
			}
		}
	}
}
