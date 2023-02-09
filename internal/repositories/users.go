package repositories

import (
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"reflect"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func CreateUser(userData models.User) (models.User, error) {
	db := db.GetDB()

	// Hash password
	bytes, err := bcrypt.GenerateFromPassword([]byte(userData.Password), 14)
	if err != nil {
		return models.User{}, err
	}
	userData.Password = string(bytes)

	err = db.Transaction(func(tx *gorm.DB) error {
		value := reflect.ValueOf(userData.UserRole)

		if !value.IsValid() {
			var usrCnt int64
			// Set user's role
			db.Model(models.User{}).Count(&usrCnt)
			// Save User
			if usrCnt == 0 {
				userData.UserRole = models.ADMIN
			} else {
				userData.UserRole = models.USER
			}
		}

		err = db.Create(&userData).Error

		if err != nil {
			return err
		}

		// Create default group with user as group member
		group := models.Group{
			Name:           "Home",
			IsDefaultGroup: true,
			GroupMembers:   models.GroupMember{UserID: userData.ID, GroupRole: models.OWNER},
		}
		err = db.Create(&group).Error

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return models.User{}, nil
	}

	return userData, nil
}
