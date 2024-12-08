package repositories

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"testing"
)

func TestShouldCreateAdminUserWithGroup(t *testing.T) {
	defer TruncateTestDb()
	db := GetDB()
	userToCreate := commands.SignUpCommand{
		Username:    "test",
		DisplayName: "test",
		Password:    "a really secure password",
	}
	userRepository := NewUserRepository(nil)
	createdUser, err := userRepository.CreateUser(userToCreate)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	validateUser(t, createdUser, userToCreate, models.ADMIN, 1)

	var group models.Group
	db.Table("groups").Where("id = 1").Preload("GroupMembers").First(&group)

	validateGroup(t, group, 1, 1)
}

func TestShouldCreateNonAdminUserWithGroup(t *testing.T) {
	defer TruncateTestDb()
	db := GetDB()
	CreateTestUser()
	CreateTestGroup()
	userToCreate := commands.SignUpCommand{
		Username:    "test2",
		DisplayName: "test",
		Password:    "a really secure password",
	}
	userRepository := NewUserRepository(nil)
	createdUser, err := userRepository.CreateUser(userToCreate)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	validateUser(t, createdUser, userToCreate, models.USER, 2)

	var group models.Group
	db.Table("groups").Where("id = 2").Preload("GroupMembers").First(&group)

	validateGroup(t, group, 2, 2)
}

func TestShouldReturnErrorWhenCreatingUserWithDuplicateUsername(t *testing.T) {
	defer TruncateTestDb()
	CreateTestUser()
	CreateTestGroup()
	userToCreate := commands.SignUpCommand{
		Username:    "test",
		DisplayName: "test",
		Password:    "a really secure password",
	}
	userRepository := NewUserRepository(nil)
	_, err := userRepository.CreateUser(userToCreate)
	if err == nil {
		utils.PrintTestError(t, err, "error")
	}
}

func TestShouldBeFirstAdminToLogin(t *testing.T) {
	defer TruncateTestDb()
	userToCreate := commands.SignUpCommand{
		Username:    "test",
		DisplayName: "test",
		Password:    "a really secure password",
	}
	userRepository := NewUserRepository(nil)
	createdUser, err := userRepository.CreateUser(userToCreate)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	validateUser(t, createdUser, userToCreate, models.ADMIN, 1)

	firstAdminToLogin, err := userRepository.IsFirstAdminToLogin()

	if firstAdminToLogin != true {
		utils.PrintTestError(t, firstAdminToLogin, true)
	}
}

func TestShouldNotBeFirstAdminToLogin(t *testing.T) {
	defer TruncateTestDb()
	userToCreate := commands.SignUpCommand{
		Username:    "test",
		DisplayName: "test",
		Password:    "a really secure password",
	}
	userRepository := NewUserRepository(nil)
	createdUser, err := userRepository.CreateUser(userToCreate)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	validateUser(t, createdUser, userToCreate, models.ADMIN, 1)

	userRepository.UpdateUserLastLoginDate(1)

	firstAdminToLogin, err := userRepository.IsFirstAdminToLogin()
	if firstAdminToLogin != false {
		utils.PrintTestError(t, firstAdminToLogin, false)
	}
}

func validateUser(t *testing.T, createdUser models.User, userToCreate commands.SignUpCommand, role models.UserRole, id uint) {
	if createdUser.ID != id {
		utils.PrintTestError(t, createdUser.ID, id)
	}
	if createdUser.Password == userToCreate.Password {
		utils.PrintTestError(t, createdUser.Password, "hashed password")
	}
	if createdUser.DefaultAvatarColor != "#27b1ff" {
		utils.PrintTestError(t, createdUser.DefaultAvatarColor, "#27b1ff")
	}
	if createdUser.UserRole != role {
		utils.PrintTestError(t, createdUser.UserRole, models.ADMIN)
	}
}

func validateGroup(t *testing.T, group models.Group, id uint, userId uint) {
	if group.ID != id {
		utils.PrintTestError(t, group.ID, id)
	}
	if group.GroupMembers[0].UserID != userId {
		utils.PrintTestError(t, group.GroupMembers[0].UserID, userId)
	}
	if group.Name != "My Receipts" {
		utils.PrintTestError(t, group.Name, "My Receipts")
	}

}
