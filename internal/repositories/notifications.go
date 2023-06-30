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
