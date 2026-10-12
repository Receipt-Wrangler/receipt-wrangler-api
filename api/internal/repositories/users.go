package repositories

import (
	"errors"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/permissions"
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

// CountByIds returns how many of the given user ids exist. Used to validate that
// a group role's paid-by user grants reference real users. Duplicate ids in the
// input are de-duplicated by the IN clause, so callers should pass a unique set.
func (repository UserRepository) CountByIds(ids []uint) (int64, error) {
	db := repository.GetDB()

	var count int64
	err := db.Model(&models.User{}).Where("id IN ?", ids).Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
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

		if userData.AppRoleID != nil {
			// The admin-protected create endpoint chose a modern app role
			// explicitly. Honor it.
			user.AppRoleID = userData.AppRoleID
		} else {
			// Assign the modern app role so the account is not locked out under
			// permission enforcement: Legacy Admin for the first/bootstrap user
			// (an administrator, so it is never the configurable default), the
			// configurable default app role otherwise.
			var usrCnt int64
			tx.Model(models.User{}).Count(&usrCnt)
			isAdmin := usrCnt == 0

			appRoleId, roleErr := repository.resolveAppRoleId(tx, isAdmin)
			if roleErr != nil {
				repository.ClearTransaction()
				return roleErr
			}
			user.AppRoleID = appRoleId
		}

		err = repository.GetDB().Create(&user).Error

		if err != nil {
			repository.ClearTransaction()
			return err
		}

		// An app role may opt its users out of the personal "My Receipts" group, for
		// accounts that are only ever meant to live in specific shared groups. The
		// virtual "All" group is always created, so the account still has a working
		// dashboard.
		skipDefaultGroup, err := repository.appRoleSkipsDefaultGroup(tx, user.AppRoleID)
		if err != nil {
			repository.ClearTransaction()
			return err
		}

		if !skipDefaultGroup {
			groupCommand := commands.UpsertGroupCommand{
				Name:           "My Receipts",
				IsDefaultGroup: true,
			}

			_, err := groupRepository.CreateGroup(groupCommand, user.ID)
			if err != nil {
				repository.ClearTransaction()
				return err
			}
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

// appRoleSkipsDefaultGroup reports whether a newly-created user's app role opts
// out of the personal "My Receipts" group. Best-effort in the same way as
// resolveAppRoleId: an unassigned (nil) or missing role falls back to creating
// the group rather than failing user creation. Any *other* lookup error
// propagates and rolls the creation back — a transient DB failure must not
// silently decide which groups an account is created with.
func (repository UserRepository) appRoleSkipsDefaultGroup(tx *gorm.DB, appRoleId *uint) (bool, error) {
	if appRoleId == nil {
		return false, nil
	}

	skip, err := NewRoleRepository(tx).AppRoleSkipsDefaultGroup(*appRoleId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return skip, nil
}

// resolveAppRoleId picks the modern app role for a newly-created user: the
// Legacy Admin role for an administrator (the first/bootstrap user, so it is
// never assigned the configurable default and locked out of administration) and
// the configurable default app role for everyone else. Returns nil when the role
// can't be resolved (e.g. an unseeded test database), leaving the user without a
// modern role rather than failing creation.
func (repository UserRepository) resolveAppRoleId(tx *gorm.DB, isAdmin bool) (*uint, error) {
	roleRepository := NewRoleRepository(tx)
	if isAdmin {
		return roleRepository.GetAppRoleIdByName(LegacyAdminRoleName)
	}
	return roleRepository.GetDefaultAppRoleId()
}

// UpdateUser updates the editable fields of a user. The app role column is only
// written when an id is supplied, so callers that omit it never clear an existing
// assignment.
func (repository UserRepository) UpdateUser(id string, command commands.SignUpCommand) error {
	db := repository.GetDB()

	updates := map[string]interface{}{
		"username":     command.Username,
		"display_name": command.DisplayName,
	}

	if command.AppRoleID != nil {
		updates["app_role_id"] = *command.AppRoleID
	}

	return db.Table("users").Where("id = ?", id).Updates(updates).Error
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

func (repository UserRepository) GetPagedUsers(command commands.PagedRequestCommand) ([]structs.UserView, int64, error) {
	db := repository.GetDB()
	var results []structs.UserView
	var count int64

	query := db.Model(&models.User{})

	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	validColumn := repository.isUserSortColumn(command.OrderBy)
	if !validColumn {
		return nil, 0, errors.New("invalid column: " + command.OrderBy)
	}

	query = repository.Sort(query, command.OrderBy, command.SortDirection)
	query = query.Scopes(repository.Paginate(command.Page, command.PageSize))

	err = query.Find(&results).Error
	if err != nil {
		return nil, 0, err
	}

	return results, count, nil
}

func (repository UserRepository) isUserSortColumn(orderBy string) bool {
	return orderBy == "username" ||
		orderBy == "display_name" ||
		orderBy == "created_at" ||
		orderBy == "updated_at"
}

func (repository UserRepository) GetUserById(userId uint) (structs.UserView, error) {
	var user structs.UserView

	err := repository.GetDB().Model(models.User{}).Where("id = ?", userId).First(&user).Error
	if err != nil {
		return structs.UserView{}, err
	}

	return user, nil
}

// GetUserByUsername resolves a user by their unique username. Mirrors
// GetUserById's view type so callers that only need identity do not have to
// handle the password hash.
func (repository UserRepository) GetUserByUsername(username string) (structs.UserView, error) {
	var user structs.UserView

	err := repository.GetDB().Model(models.User{}).Where("username = ?", username).First(&user).Error
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

// IsFirstAdminToLogin reports whether no administrator has logged in yet, where
// "administrator" is any user whose app role grants app.users.read (the modern
// replacement for the removed UserRole == ADMIN check). Returns true when no such
// user has a last_login_date.
func (repository UserRepository) IsFirstAdminToLogin() (bool, error) {
	roleRepository := NewRoleRepository(repository.TX)
	adminRoleIds, err := roleRepository.appRoleIdsWithPermission(permissions.AppUsersRead)
	if err != nil {
		return false, err
	}
	if len(adminRoleIds) == 0 {
		return true, nil
	}

	foundUser := models.User{}
	err = repository.
		GetDB().
		Limit(1).
		Select("id").
		Model(models.User{}).
		Where("app_role_id IN ? AND last_login_date IS NOT NULL", adminRoleIds).
		Find(&foundUser).
		Error

	if err != nil {
		return false, err
	}

	return foundUser.ID == 0, nil
}
