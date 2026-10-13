package repositories

import (
	"encoding/json"
	"fmt"
	"os"
	"receipt-wrangler/api/internal/commands"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"

	"gorm.io/gorm"
)

func SetUpTestEnv() {
	os.Args = append(os.Args, "-env=test")
	logging.InitLog()
	config.SetConfigs()
}

func TruncateTable(db *gorm.DB, tableName string) error {
	query := fmt.Sprintf("DELETE FROM %s", tableName)
	err := db.Exec(query).Error
	if err != nil {
		return err
	}

	resetSeqQuery := fmt.Sprintf("DELETE FROM sqlite_sequence WHERE name='%s';", tableName)
	return db.Exec(resetSeqQuery).Error
}

func TestTeardown() {
	os.RemoveAll("./logs")
	RemoveTestDb()
}

func CreateTestUser() {
	db := GetDB()
	user := models.User{
		Username:    "test",
		Password:    "Password",
		DisplayName: "test",
	}

	db.Create(&user)
}

func CreateTestGroup() {
	db := GetDB()
	group := models.Group{
		Name: "test",
	}

	db.Create(&group)
}

func CreateTestGroupWithUsers() {
	db := GetDB()
	group := models.Group{
		BaseModel: models.BaseModel{
			ID: 1,
		},
		Name: "test",
	}

	group2 := models.Group{
		Name: "test2",
	}

	user := models.User{
		Username:    "test",
		DisplayName: "asdf",
		Password:    "1",
	}
	user2 := models.User{
		Username:    "test1",
		DisplayName: "asdf",
		Password:    "1",
	}
	user3 := models.User{
		Username:    "test2",
		DisplayName: "asdf",
		Password:    "1",
	}
	user4 := models.User{
		Username:    "test3",
		DisplayName: "asdf",
		Password:    "1",
	}

	groupMember := models.GroupMember{
		GroupID: 1,
		UserID:  1,
	}
	groupMember2 := models.GroupMember{
		GroupID: 1,
		UserID:  2,
	}
	groupMember3 := models.GroupMember{
		GroupID: 1,
		UserID:  3,
	}
	groupMember4 := models.GroupMember{
		GroupID: 2,
		UserID:  4,
	}

	db.Create(&group)
	db.Create(&group2)

	db.Table("users").Create(&user)
	db.Table("users").Create(&user2)
	db.Table("users").Create(&user3)
	db.Table("users").Create(&user4)

	db.Model(models.GroupMember{}).Create(&groupMember)
	db.Model(models.GroupMember{}).Create(&groupMember2)
	db.Model(models.GroupMember{}).Create(&groupMember3)
	db.Model(models.GroupMember{}).Create(&groupMember4)
}

func CreateTestCategories() {
	db := GetDB()
	category := models.Category{
		Name: "test",
	}

	category2 := models.Category{
		Name: "test2",
	}

	category3 := models.Category{
		Name: "test3",
	}

	db.Create(&category)
	db.Create(&category2)
	db.Create(&category3)
}

func CreateTestRoles() {
	db := GetDB()

	appRole := models.AppRole{
		Name:        "Test App Role",
		Description: "test app role",
		Permissions: []models.AppRolePermission{
			{Permission: "app.users.create"},
		},
	}

	groupRole := models.GroupRoleDefinition{
		Name:        "Test Group Role",
		Description: "test group role",
		Permissions: []models.GroupRolePermission{
			{Permission: "group.receipts.create"},
		},
	}

	db.Create(&appRole)
	db.Create(&groupRole)
}

func TruncateTestDb() {
	db := GetDB()

	// Get all table names
	var tables []string
	db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&tables)

	// Disable foreign key constraints temporarily
	db.Exec("PRAGMA foreign_keys = OFF")

	// Begin transaction
	tx := db.Begin()

	// Truncate all tables and reset sequences
	for _, table := range tables {
		tx.Exec("DELETE FROM " + table)
		tx.Exec("DELETE FROM sqlite_sequence WHERE name='" + table + "'")
	}

	// Commit transaction
	tx.Commit()

	// Re-enable foreign key constraints
	db.Exec("PRAGMA foreign_keys = ON")
}

func RemoveTestDb() {
	os.Remove("./test.db")
}

// SetAppTimeZoneForTests changes the app time zone through the same
// UpdateSystemSettings path an admin's save takes, carrying every other stored
// setting over unchanged. TruncateTestDb removes the settings row, which puts
// the zone back to its UTC default.
func SetAppTimeZoneForTests(name string) error {
	return updateSystemSettingsForTests(func(command *commands.UpsertSystemSettingsCommand) {
		command.TimeZone = &name
	})
}

// SetSecureCookiesForTests turns the secureCookies System Setting on or off the
// same way. TruncateTestDb puts it back to its default, off.
func SetSecureCookiesForTests(enabled bool) error {
	return updateSystemSettingsForTests(func(command *commands.UpsertSystemSettingsCommand) {
		command.SecureCookies = &enabled
	})
}

// updateSystemSettingsForTests round-trips the stored settings into a command,
// lets the caller change one field, and saves it through UpdateSystemSettings.
func updateSystemSettingsForTests(change func(command *commands.UpsertSystemSettingsCommand)) error {
	repository := NewSystemSettingsRepository(nil)
	existing, err := repository.GetSystemSettings()
	if err != nil {
		return err
	}

	bytes, err := json.Marshal(existing)
	if err != nil {
		return err
	}

	var command commands.UpsertSystemSettingsCommand
	if err := json.Unmarshal(bytes, &command); err != nil {
		return err
	}
	change(&command)

	_, err = repository.UpdateSystemSettings(command)
	return err
}
