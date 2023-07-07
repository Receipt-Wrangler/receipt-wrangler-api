package repositories

import (
	"fmt"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/simpleutils"

	"gorm.io/gorm"
)

func AddComment(comment models.Comment) (models.Comment, error) {
	db := db.GetDB()
	var receipt models.Receipt

	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&comment).Create(&comment).Error
		receipt, err = GetReceiptById(simpleutils.UintToString(comment.ReceiptId))
		if err != nil {
			return err
		}

		usersToOmit := make([]interface{}, 0)
		usersToOmit = append(usersToOmit, *comment.UserId)

		if comment.CommentId == nil {
			err := SendNotificationToGroup(receipt.GroupId, "Comment Added", fmt.Sprintf("%s has added a comment to a receipt in group %s. %s", BuildParamaterisedString("userId", *comment.UserId, "displayName", "string"), BuildParamaterisedString("groupId", receipt.GroupId, "name", "string"), BuildParamaterisedString("receiptId", comment.ReceiptId, "noop", "link")), models.NOTIFICATION_TYPE_NORMAL, usersToOmit)
			if err != nil {
				return err
			}
		} else {
			threadUsers, err := GetUsersInCommentThread(comment)
			if err != nil {
				return err
			}

			err = SendNotificationToUsers(threadUsers, "Comment Replied", fmt.Sprintf("%s has replied to a comment thread that you are a part of.", BuildParamaterisedString("userId", *comment.UserId, "displayName", "string")), models.NOTIFICATION_TYPE_NORMAL, usersToOmit)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return models.Comment{}, err
	}

	return comment, nil
}

func GetUsersInCommentThread(comment models.Comment) ([]uint, error) {
	db := db.GetDB()
	userIds := make([]uint, 0)

	if *comment.UserId > 0 {
		userIds = append(userIds, *comment.UserId)
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

		if *parentComment.UserId > 0 {
			userIds = append(userIds, *parentComment.UserId)
		}

		for _, comment := range threadComments {
			if comment.ID > 0 {
				userIds = append(userIds, *comment.UserId)
			}
		}
	}

	return userIds, nil
}
