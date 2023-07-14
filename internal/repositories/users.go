package repositories

import (
	"receipt-wrangler/api/internal/commands"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"

	"gorm.io/gorm"
)

func CreateUser(userData commands.SignUpCommand) (models.User, error) {
	db := db.GetDB()
	user := models.User{
		Username:    userData.Username,
		DisplayName: userData.Displayname,
		Password:    userData.Password,
	}

	// Hash password
	bytes, err := utils.HashPassword(user.Password)
	if err != nil {
		return models.User{}, err
	}
	user.Password = string(bytes)

	err = db.Transaction(func(tx *gorm.DB) error {
		value := user.UserRole

		if len(value) == 0 {
			var usrCnt int64
			// Set user's role
			tx.Model(models.User{}).Count(&usrCnt)
			// Save User
			if usrCnt == 0 {
				user.UserRole = models.ADMIN
			} else {
				user.UserRole = models.USER
			}
		}

		err = tx.Create(&user).Error

		if err != nil {
			return err
		}

		var groupMembers = make([]models.GroupMember, 1)
		groupMembers = append(groupMembers, models.GroupMember{UserID: user.ID, GroupRole: models.OWNER})
		// Create default group with user as group member
		group := models.Group{
			Name:           "Home",
			IsDefaultGroup: true,
			GroupMembers:   groupMembers,
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

	return user, nil
}
