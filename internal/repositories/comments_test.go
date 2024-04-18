package repositories

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"testing"
)

var commentRepository CommentRepository

func setupCommentTest() {
	CreateTestGroupWithUsers()
	createTestReceipt()
	commentRepository = NewCommentRepository(nil)
}

func createTestReceipt() {
	receipt := models.Receipt{
		Name:         "test",
		PaidByUserID: 1,
		GroupId:      1,
	}

	GetDB().Create(&receipt)
}

func teardownCommentTest() {
	TruncateTestDb()
}

func TestShouldAddCommentAndSendNotificationToAllGroupUsers(t *testing.T) {
	defer teardownCommentTest()
	setupCommentTest()
	userId := uint(1)
	comment := commands.UpsertCommentCommand{
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
}
