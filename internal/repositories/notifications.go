package repositories

import (
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
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
