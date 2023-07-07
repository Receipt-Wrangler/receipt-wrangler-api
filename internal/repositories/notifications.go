package repositories

import (
	"fmt"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/utils"
)

func GetNotificationsForUser(userId uint) ([]models.Notification, error) {
	db := db.GetDB()
	var notifications []models.Notification

	err := db.Table("notifications").Where("user_id = ?", userId).Find(&notifications).Error

	return notifications, err
}

func GetNotificationCountForUser(userId uint) (int64, error) {
	db := db.GetDB()
	var count int64

	err := db.Table("notifications").Where("user_id = ?", userId).Count(&count).Error

	return count, err
}

func GetNotificationById(notificationId string) (models.Notification, error) {
	db := db.GetDB()
	var notification models.Notification

	err := db.Table("notifications").Where("id = ?", notificationId).Find(&notification).Error

	return notification, err
}

func DeleteAllNotificationsForUser(userId uint) error {
	db := db.GetDB()
	err := db.Delete(models.Notification{}, "user_id = ?", userId).Error

	return err
}

func DeleteNotificationById(notificationId string) error {
	db := db.GetDB()
	err := db.Delete(models.Notification{}, "id = ?", notificationId).Error

	return err
}

func SendNotificationToGroup(groupId uint, title string, body string, notificationType models.NotificationType, usersToOmit []interface{}) error {
	db := db.GetDB()
	notifications, err := BuildNotificationForGroup(groupId, title, body, notificationType, usersToOmit)
	if err != nil {
		return err
	}

	err = db.Table("notifications").CreateInBatches(notifications, 20).Error

	return err
}

func BuildNotificationForGroup(groupId uint, title string, body string, notificationType models.NotificationType, usersToOmit []interface{}) ([]models.Notification, error) {
	groupMembers, err := GetsGroupMembersByGroupId(simpleutils.UintToString(groupId))
	if err != nil {
		return nil, err
	}

	notifications := make([]models.Notification, 0)
	for i := 0; i < len(groupMembers); i++ {
		groupMember := groupMembers[i]
		if !utils.Contains(usersToOmit, groupMember.UserID) {
			notification := models.Notification{
				Title:  title,
				Body:   body,
				Type:   notificationType,
				UserId: groupMember.UserID,
			}

			notifications = append(notifications, notification)
		}

	}

	return notifications, nil
}

func BuildParamaterisedString(idType string, id uint, displayKey string, typeOfData string) string {
	return fmt.Sprintf("${%s:%s.%s:%s}", idType, simpleutils.UintToString(id), displayKey, typeOfData)
}
