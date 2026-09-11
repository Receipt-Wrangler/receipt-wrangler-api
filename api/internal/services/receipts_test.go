package services

import (
	"errors"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/permissions"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm/clause"
)

// seedReceipt persists a receipt (no associations) paid by the given user.
func seedReceipt(t *testing.T, name string, groupId uint, paidByUserId uint) models.Receipt {
	t.Helper()
	receipt := models.Receipt{
		Name:         name,
		Amount:       decimal.NewFromInt(10),
		Date:         time.Now(),
		Status:       models.OPEN,
		GroupId:      groupId,
		PaidByUserID: paidByUserId,
	}
	if err := repositories.GetDB().Omit(clause.Associations).Create(&receipt).Error; err != nil {
		t.Fatalf("seed receipt: %v", err)
	}
	return receipt
}

// seedReceiptMember creates a group, a group role with the given permissions and
// grants, and a member assigned to it. Returns the member's user id and group id.
func seedReceiptMember(
	t *testing.T,
	username string,
	roleName string,
	perms []string,
	categoryGrants []uint,
	tagGrants []uint,
	paidByGrants []uint,
	includeOwn bool,
) (uint, uint) {
	t.Helper()
	db := repositories.GetDB()

	group := models.Group{Name: "grp-" + username}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("seed group: %v", err)
	}

	role, err := repositories.NewRoleRepository(nil).CreateGroupRole(roleName, "", perms, categoryGrants, tagGrants, paidByGrants, includeOwn, false)
	if err != nil {
		t.Fatalf("seed group role: %v", err)
	}

	user := models.User{Username: username, Password: "password"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed member user: %v", err)
	}
	member := models.GroupMember{GroupID: group.ID, UserID: user.ID, GroupRoleID: &role.ID}
	if err := db.Create(&member).Error; err != nil {
		t.Fatalf("seed group member: %v", err)
	}

	ClearRolePermissionCacheForTests()
	ClearGroupRoleGrantCacheForTests()
	return user.ID, group.ID
}

// giveAppRole assigns an app role with the given permissions to a user.
func giveAppRole(t *testing.T, userId uint, roleName string, perms []string) {
	t.Helper()
	role, err := repositories.NewRoleRepository(nil).CreateAppRole(roleName, "", perms, false)
	if err != nil {
		t.Fatalf("create app role: %v", err)
	}
	if err := repositories.GetDB().Model(&models.User{}).Where("id = ?", userId).Update("app_role_id", role.ID).Error; err != nil {
		t.Fatalf("assign app role: %v", err)
	}
	ClearRolePermissionCacheForTests()
}

func TestGetReceiptForUserReturnsReceiptForMember(t *testing.T) {
	defer repositories.TruncateTestDb()

	userId, groupId := seedReceiptMember(t, "getok", "getok-role", []string{permissions.GroupReceiptsRead}, nil, nil, nil, false)
	receipt := seedReceipt(t, "Lunch", groupId, userId)

	got, err := NewReceiptService(nil).GetReceiptForUser(userId, utils.UintToString(receipt.ID))
	if err != nil {
		t.Fatalf("GetReceiptForUser returned error: %v", err)
	}
	if got.ID != receipt.ID {
		t.Errorf("expected receipt %d, got %d", receipt.ID, got.ID)
	}
}

func TestGetReceiptForUserDeniesNonMember(t *testing.T) {
	defer repositories.TruncateTestDb()

	ownerId, groupId := seedReceiptMember(t, "owner", "owner-role", []string{permissions.GroupReceiptsRead}, nil, nil, nil, false)
	receipt := seedReceipt(t, "Lunch", groupId, ownerId)
	outsider := makeUser(t, "outsider")

	_, err := NewReceiptService(nil).GetReceiptForUser(outsider, utils.UintToString(receipt.ID))
	if !errors.Is(err, ErrReceiptAccessDenied) {
		t.Errorf("expected ErrReceiptAccessDenied for a non-member, got %v", err)
	}
}

func TestGetReceiptForUserDeniesMemberWithoutReadPermission(t *testing.T) {
	defer repositories.TruncateTestDb()

	// A member whose group role lacks group.receipts.read is denied — proving the
	// permission gate is enforced independently of group membership.
	userId, groupId := seedReceiptMember(t, "noread", "noread-role", []string{}, nil, nil, nil, false)
	receipt := seedReceipt(t, "Lunch", groupId, userId)

	_, err := NewReceiptService(nil).GetReceiptForUser(userId, utils.UintToString(receipt.ID))
	if !errors.Is(err, ErrReceiptAccessDenied) {
		t.Errorf("expected ErrReceiptAccessDenied for a member without group.receipts.read, got %v", err)
	}
}

func TestGetReceiptForUserDeniesPaidByHidden(t *testing.T) {
	defer repositories.TruncateTestDb()

	allowedPayer := makeUser(t, "allowedpayer")
	hiddenPayer := makeUser(t, "hiddenpayer")
	// Restrict the member to receipts paid by allowedPayer only.
	userId, groupId := seedReceiptMember(t, "pbviewer", "pb-role", []string{permissions.GroupReceiptsRead}, nil, nil, []uint{allowedPayer}, false)

	visibleReceipt := seedReceipt(t, "allowed", groupId, allowedPayer)
	hiddenReceipt := seedReceipt(t, "hidden", groupId, hiddenPayer)

	if _, err := NewReceiptService(nil).GetReceiptForUser(userId, utils.UintToString(visibleReceipt.ID)); err != nil {
		t.Fatalf("expected the allowed receipt to be visible: %v", err)
	}
	if _, err := NewReceiptService(nil).GetReceiptForUser(userId, utils.UintToString(hiddenReceipt.ID)); !errors.Is(err, ErrReceiptAccessDenied) {
		t.Errorf("expected ErrReceiptAccessDenied for a paid-by-hidden receipt, got %v", err)
	}
}

func TestGetReceiptForUserStripsCategoriesAndTagsToGrants(t *testing.T) {
	defer repositories.TruncateTestDb()
	db := repositories.GetDB()

	allowedCategory := models.Category{Name: "allowed-cat"}
	hiddenCategory := models.Category{Name: "hidden-cat"}
	allowedTag := models.Tag{Name: "allowed-tag"}
	hiddenTag := models.Tag{Name: "hidden-tag"}
	for _, m := range []interface{}{&allowedCategory, &hiddenCategory, &allowedTag, &hiddenTag} {
		if err := db.Create(m).Error; err != nil {
			t.Fatalf("create fixture: %v", err)
		}
	}

	// The role grants read plus exactly one category and one tag.
	userId, groupId := seedReceiptMember(t, "stripper", "strip-role",
		[]string{permissions.GroupReceiptsRead}, []uint{allowedCategory.ID}, []uint{allowedTag.ID}, nil, false)
	receipt := seedReceipt(t, "with-cats-and-tags", groupId, userId)
	if err := db.Model(&receipt).Association("Categories").Append([]models.Category{allowedCategory, hiddenCategory}); err != nil {
		t.Fatalf("attach categories: %v", err)
	}
	if err := db.Model(&receipt).Association("Tags").Append([]models.Tag{allowedTag, hiddenTag}); err != nil {
		t.Fatalf("attach tags: %v", err)
	}

	got, err := NewReceiptService(nil).GetReceiptForUser(userId, utils.UintToString(receipt.ID))
	if err != nil {
		t.Fatalf("GetReceiptForUser returned error: %v", err)
	}
	if len(got.Categories) != 1 || got.Categories[0].ID != allowedCategory.ID {
		t.Errorf("expected only the granted category, got %+v", got.Categories)
	}
	if len(got.Tags) != 1 || got.Tags[0].ID != allowedTag.ID {
		t.Errorf("expected only the granted tag, got %+v", got.Tags)
	}
}

func TestGetReceiptForUserDeniesMissingReceipt(t *testing.T) {
	defer repositories.TruncateTestDb()

	userId := makeUser(t, "any")
	if _, err := NewReceiptService(nil).GetReceiptForUser(userId, "999999"); !errors.Is(err, ErrReceiptAccessDenied) {
		t.Errorf("expected ErrReceiptAccessDenied for a missing receipt, got %v", err)
	}
}

func TestSearchReceiptsForUserRequiresSearchPermission(t *testing.T) {
	defer repositories.TruncateTestDb()

	userId := makeUser(t, "nosearch")
	if _, err := NewReceiptService(nil).SearchReceiptsForUser(userId, "coffee", 100); !errors.Is(err, ErrSearchForbidden) {
		t.Errorf("expected ErrSearchForbidden without app.receipts.search, got %v", err)
	}
}

func TestSearchReceiptsForUserScopesAndAppliesPaidBy(t *testing.T) {
	defer repositories.TruncateTestDb()

	allowedPayer := makeUser(t, "spallowed")
	hiddenPayer := makeUser(t, "sphidden")
	userId, groupId := seedReceiptMember(t, "spsearcher", "sp-role", []string{permissions.GroupReceiptsRead}, nil, nil, []uint{allowedPayer}, false)
	giveAppRole(t, userId, "sp-app-role", []string{permissions.AppReceiptsSearch})

	// A receipt in another group the user does not belong to must be excluded by scope.
	otherGroup := models.Group{Name: "other"}
	if err := repositories.GetDB().Create(&otherGroup).Error; err != nil {
		t.Fatalf("create other group: %v", err)
	}

	seedReceipt(t, "Coffee allowed", groupId, allowedPayer)
	seedReceipt(t, "Coffee hidden", groupId, hiddenPayer)
	seedReceipt(t, "Coffee elsewhere", otherGroup.ID, allowedPayer)

	results, err := NewReceiptService(nil).SearchReceiptsForUser(userId, "Coffee", 100)
	if err != nil {
		t.Fatalf("SearchReceiptsForUser returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected exactly 1 visible result, got %d (%+v)", len(results), results)
	}
	if results[0].GroupID != groupId || results[0].PaidByUserId != allowedPayer {
		t.Errorf("expected the allowed-payer receipt in the member's group, got %+v", results[0])
	}
}

func TestSearchReceiptsForUserBlankQueryReturnsEmpty(t *testing.T) {
	defer repositories.TruncateTestDb()

	userId, groupId := seedReceiptMember(t, "blank", "blank-role", []string{permissions.GroupReceiptsRead}, nil, nil, nil, false)
	giveAppRole(t, userId, "blank-app-role", []string{permissions.AppReceiptsSearch})
	seedReceipt(t, "Coffee", groupId, userId)

	results, err := NewReceiptService(nil).SearchReceiptsForUser(userId, "   ", 100)
	if err != nil {
		t.Fatalf("SearchReceiptsForUser returned error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected a blank query to return no results, got %d", len(results))
	}
}

func TestSearchReceiptsForUserRequiresGroupReadPermission(t *testing.T) {
	defer repositories.TruncateTestDb()

	// A member whose group role lacks group.receipts.read is filtered out, proving
	// membership alone no longer grants search access. Mirrors the single-receipt
	// TestGetReceiptForUserDeniesMemberWithoutReadPermission.
	userId, groupId := seedReceiptMember(t, "srnoread", "srnoread-role", []string{}, nil, nil, nil, false)
	giveAppRole(t, userId, "srnoread-app-role", []string{permissions.AppReceiptsSearch})
	seedReceipt(t, "Coffee", groupId, userId)

	results, err := NewReceiptService(nil).SearchReceiptsForUser(userId, "Coffee", 100)
	if errors.Is(err, ErrSearchForbidden) {
		t.Fatalf("expected no results, not a denial — the caller does hold app.receipts.search")
	}
	if err != nil {
		t.Fatalf("SearchReceiptsForUser returned error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results without group.receipts.read, got %d (%+v)", len(results), results)
	}
}

func TestSearchReceiptsForUserExcludesGroupsWithoutReadPermission(t *testing.T) {
	defer repositories.TruncateTestDb()

	// The permission is applied per group, not as a blanket "has read somewhere":
	// the readable group's receipt comes back, the role-less group's does not.
	userId, readableGroupId := seedReceiptMember(t, "srmixed", "srmixed-role",
		[]string{permissions.GroupReceiptsRead}, nil, nil, nil, false)
	giveAppRole(t, userId, "srmixed-app-role", []string{permissions.AppReceiptsSearch})
	unreadableGroupId := addMemberToNewGroup(t, userId, "srmixed-noread")

	seedReceipt(t, "Coffee readable", readableGroupId, userId)
	seedReceipt(t, "Coffee unreadable", unreadableGroupId, userId)

	results, err := NewReceiptService(nil).SearchReceiptsForUser(userId, "Coffee", 100)
	if err != nil {
		t.Fatalf("SearchReceiptsForUser returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected exactly 1 result from the readable group, got %d (%+v)", len(results), results)
	}
	if results[0].GroupID != readableGroupId {
		t.Errorf("expected the receipt from group %d, got group %d", readableGroupId, results[0].GroupID)
	}
}

func TestSearchReceiptsForUserWithNoReadableGroupsReturnsEmptyNotForbidden(t *testing.T) {
	defer repositories.TruncateTestDb()

	// Read on none of the caller's groups is "no results", never a 403: they do hold
	// app.receipts.search. The non-nil assertion pins the wire shape — an empty set
	// must serialize as [], never null.
	userId, groupId := seedReceiptMember(t, "srnone", "srnone-role", []string{}, nil, nil, nil, false)
	giveAppRole(t, userId, "srnone-app-role", []string{permissions.AppReceiptsSearch})
	secondGroupId := addMemberToNewGroup(t, userId, "srnone-second")

	seedReceipt(t, "Coffee first", groupId, userId)
	seedReceipt(t, "Coffee second", secondGroupId, userId)

	results, err := NewReceiptService(nil).SearchReceiptsForUser(userId, "Coffee", 100)
	if errors.Is(err, ErrSearchForbidden) {
		t.Fatalf("expected an empty result, not ErrSearchForbidden")
	}
	if err != nil {
		t.Fatalf("SearchReceiptsForUser returned error: %v", err)
	}
	if results == nil {
		t.Fatalf("expected a non-nil empty slice so the response marshals as [], not null")
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d (%+v)", len(results), results)
	}
}
