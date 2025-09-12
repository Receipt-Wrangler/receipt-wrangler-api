package repositories

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"time"

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
		userPreferencesRepository := NewUserPreferencesRepository(tx)

		userRole := user.UserRole

		if len(userRole) == 0 {
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

		groupCommand := commands.UpsertGroupCommand{
			Name:           "My Receipts",
			IsDefaultGroup: true,
		}

		_, err := groupRepository.CreateGroup(groupCommand, user.ID)
		if err != nil {
			repository.ClearTransaction()
			return err
		}

		_, err = groupRepository.CreateAllGroup(user.ID)
		if err != nil {
			repository.ClearTransaction()
			return err
		}

		userPreferences := models.UserPrefernces{UserId: user.ID}
		_, err = userPreferencesRepository.CreateUserPreferences(userPreferences)
		if err != nil {
			return err
		}

		repository.ClearTransaction()
		userPreferencesRepository.ClearTransaction()
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

func (repository UserRepository) GetAllUserViews() ([]structs.UserView, error) {
	var users []structs.UserView

	err := repository.GetDB().Model(models.User{}).Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (repository UserRepository) GetUserById(userId uint) (structs.UserView, error) {
	var user structs.UserView

	err := repository.GetDB().Model(models.User{}).Where("id = ?", userId).First(&user).Error
	if err != nil {
		return structs.UserView{}, err
	}

	return user, nil
}

func (repository UserRepository) UpdateUserLastLoginDate(userId uint) (time.Time, error) {
	now := time.Now()
	err := repository.GetDB().Model(models.User{}).Where("id = ?", userId).Update("last_login_date", now).Error

	if err != nil {
		return time.Time{}, err
	}

	return now, nil
}

func (repository UserRepository) IsFirstAdminToLogin() (bool, error) {
	foundUser := models.User{}

	err := repository.
		GetDB().
		Limit(1).
		Select("id").
		Model(models.User{}).
		Where("user_role = ? AND last_login_date IS NOT NULL", models.ADMIN).
		Find(&foundUser).
		Error

	if err != nil {
		return false, err
	}

	return foundUser.ID == 0, nil
}
