package repositories

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"

	"gorm.io/gorm"
)

type UserRepository struct {
	BaseRepository
}

func NewUserRepository(tx *gorm.DB) UserRepository {
	repository := UserRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository UserRepository) CreateUser(userData commands.SignUpCommand) (models.User, error) {
	db := repository.GetDB()
	user := models.User{
		Username:           userData.Username,
		DisplayName:        userData.DisplayName,
		Password:           userData.Password,
		IsDummyUser:        userData.IsDummyUser,
		DefaultAvatarColor: "#27b1ff",
	}

	// Hash password
	bytes, err := utils.HashPassword(user.Password)
	if err != nil {
		return models.User{}, err
	}
	user.Password = string(bytes)

	err = db.Transaction(func(tx *gorm.DB) error {
		groupRepository := NewGroupRepository(tx)
		repository.SetTransaction(tx)
		value := user.UserRole

		if len(value) == 0 {
			var usrCnt int64
			tx.Model(models.User{}).Count(&usrCnt)
			if usrCnt == 0 {
				user.UserRole = models.ADMIN
			} else {
				user.UserRole = models.USER
			}
		}

		err = repository.GetDB().Create(&user).Error

		if err != nil {
			repository.ClearTransaction()
			return err
		}

		group := models.Group{
			Name:           "My Receipts",
			IsDefaultGroup: true,
			GroupSettings:  models.GroupSettings{},
		}

		_, err := groupRepository.CreateGroup(group, user.ID)
		if err != nil {
			repository.ClearTransaction()
			return err
		}

		err = repository.GetDB().Create(&group).Error

		if err != nil {
			repository.ClearTransaction()
			return err
		}

		repository.ClearTransaction()
		return nil
	})
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (repository UserRepository) CreateUserIfNoneExist() error {
	repository.GetDB()
	var userCount int64

	err := repository.GetDB().Model(models.User{}).Count(&userCount).Error
	if err != nil {
		return err
	}

	if userCount == 0 {
		_, err = repository.CreateUser(commands.GetDefaultAdminSignUpCommand())
		if err != nil {
			return err
		}
	}

	return nil
}
