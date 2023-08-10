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
	config.SetConfigs()
	logging.InitLog()
}

func TruncateTable(db *gorm.DB, tableName string) error {
	query := fmt.Sprintf("DELETE FROM %s", tableName)
	return db.Exec(query).Error
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
		Name: "test",
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

	db.Create(&group)

	db.Table("users").Create(&user)
	db.Table("users").Create(&user2)
	db.Table("users").Create(&user3)

	db.Model(models.GroupMember{}).Create(&groupMember)
	db.Model(models.GroupMember{}).Create(&groupMember2)
	db.Model(models.GroupMember{}).Create(&groupMember3)
}

func RemoveTestDb() {
	os.Remove("./test.db")
}
