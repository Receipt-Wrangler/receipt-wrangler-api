package repositories

import (
	"fmt"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"

	"gorm.io/gorm"
)

type CommentRepository struct {
	BaseRepository
}

func NewCommentRepository(tx *gorm.DB) CommentRepository {
	repository := CommentRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository CommentRepository) AddComment(command commands.UpsertCommentCommand) (models.Comment, error) {
	db := repository.GetDB()
	comment := models.Comment{
		Comment:   command.Comment,
		ReceiptId: command.ReceiptId,
		UserId:    command.UserId,
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		repository.SetTransaction(tx)

		err := tx.Model(&comment).Create(&comment).Error
		if err != nil {
			return err
		}

		err = repository.sendNotificationsToUsers(comment)
		if err != nil {
			return err
		}

		repository.ClearTransaction()
		return nil
	})

	if err != nil {
		return models.Comment{}, err
	}

	return comment, nil
}

func (repository CommentRepository) GetUsersInCommentThread(comment models.Comment) ([]uint, error) {
	db := repository.GetDB()
	userIds := make([]interface{}, 0)
	result := make([]uint, 0)

	if *comment.UserId > 0 {
		userIds = append(userIds, *comment.UserId)
		result = append(result, *comment.UserId)
	}

	if *comment.CommentId > 0 {
		var threadComments []models.Comment
		var parentComment models.Comment

		err := db.Model(models.Comment{}).Where("comment_id = ?", comment.CommentId).Find(&threadComments).Error
		if err != nil {
			return nil, err
		}

		err = db.Model(models.Comment{}).Where("id = ?", comment.CommentId).Find(&parentComment).Error
		if err != nil {
			return nil, err
		}

		if *parentComment.UserId > 0 && !utils.Contains(userIds, *parentComment.UserId) {
			userIds = append(userIds, *parentComment.UserId)
			result = append(result, *parentComment.UserId)
		}

		for _, comment := range threadComments {
			if comment.ID > 0 && !utils.Contains(userIds, *comment.UserId) {
				userIds = append(userIds, *comment.UserId)
				result = append(result, *comment.UserId)

			}
		}
	}

	return result, nil
}

func (repository CommentRepository) DeleteComment(commentId string, tokenUserId uint) error {
	db := repository.GetDB()
	var comment models.Comment

	err := db.Model(models.Comment{}).Where("id = ?", commentId).First(&comment).Error
	if err != nil {
		return err
	}

	if *comment.UserId == tokenUserId {
		err = db.Model(models.Comment{}).Where("id = ?", commentId).Delete(&comment).Error
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("not allowed to delete another user's comment")
	}

	return nil
}

func (repository CommentRepository) sendNotificationsToUsers(comment models.Comment) error {
	var receipt models.Receipt
	usersToOmit := make([]interface{}, 0)
	usersToOmit = append(usersToOmit, *comment.UserId)
	notificationRepository := NewNotificationRepository(repository.TX)
	receiptRepository := NewReceiptRepository(repository.TX)

	receipt, err := receiptRepository.GetReceiptById(utils.UintToString(comment.ReceiptId))
	if err != nil {
		return err
	}

	if comment.CommentId == nil {
		err := notificationRepository.SendNotificationToGroup(receipt.GroupId, "Comment Added", fmt.Sprintf("%s has added a comment to a receipt in group %s. %s", BuildParamaterisedString("userId", *comment.UserId, "displayName", "string"), BuildParamaterisedString("groupId", receipt.GroupId, "name", "string"), BuildParamaterisedString("receiptId", comment.ReceiptId, "noop", "link")), models.NOTIFICATION_TYPE_NORMAL, usersToOmit)
		if err != nil {
			return err
		}
	} else {
		threadUsers, err := repository.GetUsersInCommentThread(comment)
		if err != nil {
			return err
		}

		err = notificationRepository.SendNotificationToUsers(threadUsers, "Comment Replied", fmt.Sprintf("%s has replied to a thread that you are a part of.", BuildParamaterisedString("userId", *comment.UserId, "displayName", "string")), models.NOTIFICATION_TYPE_NORMAL, usersToOmit)
		if err != nil {
			return err
		}
	}

	return nil
}
