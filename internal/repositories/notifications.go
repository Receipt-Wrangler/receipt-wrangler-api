package repositories

import (
	"fmt"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"

	"gorm.io/gorm"
)

type NotificationRepository struct {
	BaseRepository
}

func NewNotificationRepository(tx *gorm.DB) NotificationRepository {
	repository := NotificationRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository NotificationRepository) GetNotificationsForUser(userId uint) ([]models.Notification, error) {
	db := repository.GetDB()
	var notifications []models.Notification

	err := db.Table("notifications").Where("user_id = ?", userId).Find(&notifications).Error

	return notifications, err
}

func (repository NotificationRepository) GetNotificationCountForUser(userId uint) (int64, error) {
	db := repository.GetDB()
	var count int64

	err := db.Table("notifications").Where("user_id = ?", userId).Count(&count).Error

	return count, err
}

func (repository NotificationRepository) GetNotificationById(notificationId string) (models.Notification, error) {
	db := repository.GetDB()
	var notification models.Notification

	err := db.Table("notifications").Where("id = ?", notificationId).Find(&notification).Error

	return notification, err
}

func (repository NotificationRepository) DeleteAllNotificationsForUser(userId uint) error {
	db := repository.GetDB()
	err := db.Delete(models.Notification{}, "user_id = ?", userId).Error

	return err
}

func (repository NotificationRepository) DeleteNotificationById(notificationId string) error {
	db := repository.GetDB()
	err := db.Delete(models.Notification{}, "id = ?", notificationId).Error

	return err
}

func (repository NotificationRepository) SendNotificationToGroup(groupId uint, title string, body string, notificationType models.NotificationType, usersToOmit []interface{}) error {
	db := repository.GetDB()
	notifications, err := BuildNotificationForGroup(groupId, title, body, notificationType, usersToOmit)
	if err != nil {
		return err
	}

	err = db.Table("notifications").CreateInBatches(&notifications, 20).Error

	return err
}

func (repository NotificationRepository) SendNotificationToUsers(userIds []uint, title string, body string, notificationType models.NotificationType, usersToOmit []interface{}) error {
	db := repository.GetDB()
	notifications, err := BuildNotificationsForUsers(userIds, title, body, notificationType, usersToOmit)
	if err != nil {
		return nil
	}

	err = db.Table("notifications").CreateInBatches(&notifications, 20).Error
	if err != nil {
		return err
	}

	return nil
}

func BuildNotificationsForUsers(userIds []uint, title string, body string, notificationType models.NotificationType, usersToOmit []interface{}) ([]models.Notification, error) {
	notifications := make([]models.Notification, 0)

	for _, id := range userIds {
		if !utils.Contains(usersToOmit, id) {
			notification := models.Notification{
				Title:  title,
				Body:   body,
				Type:   notificationType,
				UserId: id,
			}
			notifications = append(notifications, notification)
		}

	}

	return notifications, nil
}

func BuildNotificationForGroup(groupId uint, title string, body string, notificationType models.NotificationType, usersToOmit []interface{}) ([]models.Notification, error) {
	groupMemberRepository := NewGroupMemberRepository(nil)
	groupMembers, err := groupMemberRepository.GetsGroupMembersByGroupId(utils.UintToString(groupId))
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
	return fmt.Sprintf("${%s:%s.%s:%s}", idType, utils.UintToString(id), displayKey, typeOfData)
}
