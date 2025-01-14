package repositories

import (
	"fmt"
	"os"
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

func TruncateTestDb() {
	db := GetDB()
	TruncateTable(db, "system_settings")
	TruncateTable(db, "receipt_processing_settings")
	TruncateTable(db, "notifications")
	TruncateTable(db, "prompts")
	TruncateTable(db, "system_emails")
	TruncateTable(db, "system_tasks")
	TruncateTable(db, "comments")
	TruncateTable(db, "receipts")
	TruncateTable(db, "group_members")
	TruncateTable(db, "group_receipt_settings")
	TruncateTable(db, "group_settings")
	TruncateTable(db, "groups")
	TruncateTable(db, "user_prefernces")
	TruncateTable(db, "categories")
	TruncateTable(db, "tags")
	TruncateTable(db, "users")
}

func RemoveTestDb() {
	os.Remove("./test.db")
}
