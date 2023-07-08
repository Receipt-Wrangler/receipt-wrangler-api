package repositories

import (
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"testing"
)

func setupCommentTest() {
	utils.CreateTestGroupWithUsers()
	createTestReceipt()
}

func createTestReceipt() {
	receipt := models.Receipt{
		Name:         "test",
		PaidByUserID: 1,
		GroupId:      1,
	}

	db.GetDB().Create(&receipt)
}

func teardownCommentTest() {
	db := db.GetDB()
	utils.TruncateTable(db, "comments")
	utils.TruncateTable(db, "receipts")
	utils.TruncateTable(db, "groups")
	utils.TruncateTable(db, "users")
}

func TestShouldAddComment(t *testing.T) {
	setupCommentTest()
	userId := uint(1)
	comment := models.Comment{
		Comment:   "test",
		ReceiptId: 1,
		UserId:    &userId,
	}

	_, err := AddComment(comment)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	teardownCommentTest()
}
