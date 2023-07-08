package repositories

import (
	"fmt"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"testing"
)

var commentRepository CommentRepository

func setupCommentTest() {
	utils.CreateTestGroupWithUsers()
	createTestReceipt()
	commentRepository = NewCommentRepository(nil)
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
	utils.TruncateTable(db, "notifications")
	utils.TruncateTable(db, "comments")
	utils.TruncateTable(db, "receipts")
	utils.TruncateTable(db, "group_members")
	utils.TruncateTable(db, "groups")
	utils.TruncateTable(db, "users")
}

func TestShouldAddCommentAndSendNotificationToAllGroupUsers(t *testing.T) {
	setupCommentTest()
	userId := uint(1)
	comment := models.Comment{
		Comment:   "test",
		ReceiptId: 1,
		UserId:    &userId,
	}

	newComment, err := commentRepository.AddComment(comment)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if newComment.ID != 1 {
		utils.PrintTestError(t, newComment.ID, 1)
	}

	notificationRepository := NewNotificationRepository(nil)

	user1Notifications, _ := notificationRepository.GetNotificationsForUser(1)
	if len(user1Notifications) > 0 {
		utils.PrintTestError(t, len(user1Notifications), 0)
	}

	user2Notifications, _ := notificationRepository.GetNotificationsForUser(2)
	if len(user2Notifications) != 1 {
		utils.PrintTestError(t, len(user2Notifications), 1)
	}

	user3Notifications, _ := notificationRepository.GetNotificationsForUser(3)
	if len(user3Notifications) != 1 {
		utils.PrintTestError(t, len(user3Notifications), 1)
	}

	teardownCommentTest()
}

func TestShouldAddCommentAndSendNotificationToThreadUsers(t *testing.T) {
	fmt.Println("test 2")
	db := db.GetDB()
	setupCommentTest()
	user1Id := uint(1)
	user2Id := uint(2)

	parentComment := models.Comment{
		Comment:   "test",
		ReceiptId: 1,
		UserId:    &user1Id,
	}
	db.Create(&parentComment)

	comment := models.Comment{
		Comment:   "test2",
		ReceiptId: 1,
		UserId:    &user2Id,
		CommentId: &parentComment.ID,
	}

	newComment, err := commentRepository.AddComment(comment)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if newComment.ID != 2 {
		utils.PrintTestError(t, newComment.ID, 2)
		return
	}

	notificationRepository := NewNotificationRepository(nil)

	user1Notifications, _ := notificationRepository.GetNotificationsForUser(1)
	if len(user1Notifications) != 1 {
		utils.PrintTestError(t, len(user1Notifications), 1)
	}

	user2Notifications, _ := notificationRepository.GetNotificationsForUser(2)
	if len(user2Notifications) != 0 {
		utils.PrintTestError(t, len(user2Notifications), 0)
	}

	user3Notifications, _ := notificationRepository.GetNotificationsForUser(3)
	if len(user3Notifications) != 0 {
		utils.PrintTestError(t, len(user3Notifications), 0)
	}

	teardownCommentTest()
}
