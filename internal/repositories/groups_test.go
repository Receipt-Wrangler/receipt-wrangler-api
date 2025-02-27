package repositories

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"testing"
)

func setUpGroupTest() {
	CreateTestUser()
}

func setupGroupRepository() GroupRepository {
	return NewGroupRepository(nil)
}

func teardownGroupTest() {
	TruncateTestDb()
}

func TestShouldCreateGroupSuccessfully(t *testing.T) {
	defer teardownGroupTest()
	groupToCreate := commands.UpsertGroupCommand{Name: "test"}
	setUpGroupTest()
	groupRepository := setupGroupRepository()
	createdGroup, err := groupRepository.CreateGroup(groupToCreate, 1)

	if err != nil {
		utils.PrintTestError(t, err, "Expected no error")
	}

	if createdGroup.ID != 1 {
		utils.PrintTestError(t, createdGroup.ID, "1")
	}
	if createdGroup.Name != "test" {
		utils.PrintTestError(t, createdGroup.Name, "test")
	}
	if createdGroup.Status != models.GROUP_ACTIVE {
		utils.PrintTestError(t, createdGroup.Status, "Active")
	}
	if len(createdGroup.GroupMembers) != 1 {
		utils.PrintTestError(t, createdGroup.GroupMembers, "1")
	}
	if createdGroup.GroupMembers[0].UserID != 1 {
		utils.PrintTestError(t, createdGroup.GroupMembers[0].UserID, "1")
	}
}

func TestShouldGetGroupById(t *testing.T) {
	defer teardownGroupTest()
	CreateTestGroup()
	setUpGroupTest()
	groupRepository := setupGroupRepository()
	testGroup, err := groupRepository.GetGroupById("1", false, true, true)

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}
	if testGroup.ID != 1 {
		utils.PrintTestError(t, err, "1")
	}
	if testGroup.Name != "test" {
		utils.PrintTestError(t, err, "1")
	}
}

func TestShouldGetAGroupWithGroupMembers(t *testing.T) {
	defer teardownGroupTest()
	CreateTestGroupWithUsers()
	groupRepository := setupGroupRepository()

	testGroup, err := groupRepository.GetGroupById("1", true, true, true)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}
	if testGroup.ID != 1 {
		utils.PrintTestError(t, testGroup.ID, "1")
	}
	if len(testGroup.GroupMembers) != 3 {
		utils.PrintTestError(t, err, "3")
	}
}

func TestShouldReturnErrorIfGroupDoesNotExist(t *testing.T) {
	defer teardownGroupTest()
	groupRepository := setupGroupRepository()
	testGroup, err := groupRepository.GetGroupById("2332", false, true, true)

	if err == nil {
		utils.PrintTestError(t, err, "error")
	}
	if testGroup.ID != 0 {
		utils.PrintTestError(t, testGroup.ID, "0")
	}
}

func TestShouldUpdateGroup(t *testing.T) {
	defer teardownGroupTest()
	CreateTestGroup()
	updateGroup := commands.UpsertGroupCommand{Name: "new name", Status: models.GROUP_ARCHIVED}
	groupRepository := setupGroupRepository()
	updatedGroup, err := groupRepository.UpdateGroup(updateGroup, "1")

	if err != nil {
		utils.PrintTestError(t, err, "error")
	}
	if updatedGroup.Name != "new name" {
		utils.PrintTestError(t, updatedGroup.Name, "new name")
	}
	if updatedGroup.Status != models.GROUP_ARCHIVED {
		utils.PrintTestError(t, updatedGroup.Status, models.GROUP_ARCHIVED)
	}
}
