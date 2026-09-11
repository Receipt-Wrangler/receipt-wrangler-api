package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/permissions"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/shopspring/decimal"
	"gorm.io/gorm/clause"
)

func TestNewHandlerRejectsMissingToken(t *testing.T) {
	handler := NewHandler()

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/mcp", nil))

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without a bearer token, got %d", recorder.Code)
	}

	wwwAuth := recorder.Header().Get("WWW-Authenticate")
	if !strings.Contains(wwwAuth, "resource_metadata") {
		t.Errorf("expected WWW-Authenticate to advertise resource_metadata, got %q", wwwAuth)
	}
}

// testMcpAudience is a fixed MCP resource audience for direct verifyToken
// tests, decoupled from whatever public URL System Settings resolves to.
const testMcpAudience = "https://receipts.example.com/mcp"

func TestVerifyTokenAcceptsValidJwt(t *testing.T) {
	defer repositories.TruncateTestDb()

	user := createUser(t, "tokenuser")
	accessToken, _, _, err := services.GenerateMcpJWT(user.ID, testMcpAudience)
	if err != nil {
		t.Fatalf("failed to generate jwt: %v", err)
	}

	info, err := verifyToken(context.Background(), testMcpAudience, accessToken)
	if err != nil {
		t.Fatalf("verifyToken rejected a valid token: %v", err)
	}
	if info.UserID != utils.UintToString(user.ID) {
		t.Errorf("token UserID = %q, want %q", info.UserID, utils.UintToString(user.ID))
	}
	if _, ok := info.Extra[claimsKey]; !ok {
		t.Errorf("expected verified claims to be stored on the token info")
	}
}

// TestVerifyTokenRejectsNonMcpAudience proves an MCP token is bound to the MCP
// audience: a normal REST token (and any other audience) is rejected.
func TestVerifyTokenRejectsNonMcpAudience(t *testing.T) {
	defer repositories.TruncateTestDb()

	user := createUser(t, "restuser")
	restToken, _, _, err := services.GenerateJWT(user.ID)
	if err != nil {
		t.Fatalf("failed to generate jwt: %v", err)
	}

	if _, err := verifyToken(context.Background(), testMcpAudience, restToken); err == nil {
		t.Fatalf("expected a normal REST token to be rejected by the MCP validator")
	}
}

func TestHandlerAcceptsIssuedToken(t *testing.T) {
	defer repositories.TruncateTestDb()

	user := createUser(t, "bearer")
	// The handler reads the audience live from System Settings, so mint with
	// the same resolved resource URL it will validate against.
	accessToken, _, _, err := services.GenerateMcpJWT(user.ID, services.GetMcpResourceUrl())
	if err != nil {
		t.Fatalf("failed to generate jwt: %v", err)
	}

	handler := NewHandler()
	request := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	request.Header.Set("Authorization", "Bearer "+accessToken)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	// A valid token must clear the auth layer. The transport may then reject a
	// bare GET for protocol reasons, but it must not be a 401.
	if recorder.Code == http.StatusUnauthorized {
		t.Fatalf("a valid issued token was rejected by the mcp auth layer")
	}
}

func TestVerifyTokenRejectsInvalidJwt(t *testing.T) {
	_, err := verifyToken(context.Background(), testMcpAudience, "not-a-real-token")
	if err == nil {
		t.Fatalf("expected an error for an invalid token")
	}
	if !strings.Contains(err.Error(), auth.ErrInvalidToken.Error()) {
		t.Errorf("expected error to wrap ErrInvalidToken, got %v", err)
	}
}

func createReceiptInGroup(t *testing.T, name string, groupId uint, paidByUserId uint) models.Receipt {
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
		t.Fatalf("failed to create receipt: %v", err)
	}
	return receipt
}

func addGroupMember(t *testing.T, userId uint, groupId uint, perms []string) {
	t.Helper()
	addGroupMemberWithGrants(t, userId, groupId, "mcp-test-role", perms, nil, nil, nil, false)
}

// addGroupMemberWithGrants creates a group role with the given permissions and
// category/tag/paid-by grants, then assigns the user to the group with it. Use a
// distinct roleName when a single test needs more than one role (names are unique).
func addGroupMemberWithGrants(
	t *testing.T,
	userId uint,
	groupId uint,
	roleName string,
	perms []string,
	categoryGrants []uint,
	tagGrants []uint,
	paidByGrants []uint,
	includeOwn bool,
) {
	t.Helper()
	role, err := repositories.NewRoleRepository(nil).CreateGroupRole(
		roleName, "", perms, categoryGrants, tagGrants, paidByGrants, includeOwn, false)
	if err != nil {
		t.Fatalf("failed to create group role: %v", err)
	}
	member := models.GroupMember{UserID: userId, GroupID: groupId, GroupRoleID: &role.ID}
	if err := repositories.GetDB().Create(&member).Error; err != nil {
		t.Fatalf("failed to create group member: %v", err)
	}
	// The permission/grant resolution caches are keyed by role id, which the
	// test DB can reuse across truncations — clear them so the freshly created
	// role's permissions are resolved from the DB rather than a stale entry.
	services.ClearRolePermissionCacheForTests()
	services.ClearGroupRoleGrantCacheForTests()
}

// setAppRole creates an app role with the given permissions and assigns it to the
// user, so app-scoped checks (e.g. app.receipts.search, app.categories.read)
// resolve to it.
func setAppRole(t *testing.T, userId uint, roleName string, perms []string) {
	t.Helper()
	role, err := repositories.NewRoleRepository(nil).CreateAppRole(roleName, "", perms, false)
	if err != nil {
		t.Fatalf("failed to create app role: %v", err)
	}
	if err := repositories.GetDB().Model(&models.User{}).Where("id = ?", userId).
		Update("app_role_id", role.ID).Error; err != nil {
		t.Fatalf("failed to assign app role: %v", err)
	}
	services.ClearRolePermissionCacheForTests()
}

func listCategories(t *testing.T, userId uint) []models.Category {
	t.Helper()
	_, out, err := handleListCategories(context.Background(), requestForUser(userId), emptyInput{})
	if err != nil {
		t.Fatalf("handleListCategories returned error: %v", err)
	}
	categories, ok := out.([]models.Category)
	if !ok {
		t.Fatalf("expected []models.Category, got %T", out)
	}
	return categories
}

// An app.categories.read holder (admin / category manager) sees the whole pool,
// matching the admin-only REST global list.
func TestListCategoriesAppReaderSeesAll(t *testing.T) {
	defer repositories.TruncateTestDb()

	user := createUser(t, "catadmin")
	repositories.CreateTestCategories()
	setAppRole(t, user.ID, "cat-reader", []string{permissions.AppCategoriesRead})

	if got := len(listCategories(t, user.ID)); got != 3 {
		t.Errorf("expected the app reader to see all 3 categories, got %d", got)
	}
}

// A member whose group role restricts categories sees only the granted subset.
func TestListCategoriesRestrictedToGroupGrants(t *testing.T) {
	defer repositories.TruncateTestDb()

	user := createUser(t, "catmember")
	group := models.Group{Name: "g"}
	if err := repositories.GetDB().Create(&group).Error; err != nil {
		t.Fatalf("failed to create group: %v", err)
	}
	repositories.CreateTestCategories()
	var categories []models.Category
	if err := repositories.GetDB().Order("id asc").Find(&categories).Error; err != nil {
		t.Fatalf("failed to load categories: %v", err)
	}

	// Grant only the first category.
	addGroupMemberWithGrants(t, user.ID, group.ID, "cat-grant-role",
		[]string{permissions.GroupReceiptsRead}, []uint{categories[0].ID}, nil, nil, false)

	visible := listCategories(t, user.ID)
	if len(visible) != 1 || visible[0].ID != categories[0].ID {
		t.Errorf("expected only the granted category, got %+v", visible)
	}
}

// A member whose group role has no category grants is unrestricted (sees all).
func TestListCategoriesUnrestrictedGroupMemberSeesAll(t *testing.T) {
	defer repositories.TruncateTestDb()

	user := createUser(t, "catunrestricted")
	group := models.Group{Name: "g"}
	if err := repositories.GetDB().Create(&group).Error; err != nil {
		t.Fatalf("failed to create group: %v", err)
	}
	repositories.CreateTestCategories()
	addGroupMember(t, user.ID, group.ID, []string{permissions.GroupReceiptsRead})

	if got := len(listCategories(t, user.ID)); got != 3 {
		t.Errorf("expected an unrestricted member to see all 3 categories, got %d", got)
	}
}

// A user with no app read permission and no groups sees nothing — the global pool
// is not leaked, matching the REST behavior where non-admins only get per-group
// filtered catalogs.
func TestListCategoriesNoAccessSeesNone(t *testing.T) {
	defer repositories.TruncateTestDb()

	user := createUser(t, "catnoaccess")
	repositories.CreateTestCategories()

	if got := len(listCategories(t, user.ID)); got != 0 {
		t.Errorf("expected a user with no grants to see 0 categories, got %d", got)
	}
}

// list_tags is grant-filtered the same way (spot-check the restricted path).
func TestListTagsRestrictedToGroupGrants(t *testing.T) {
	defer repositories.TruncateTestDb()

	user := createUser(t, "tagmember")
	group := models.Group{Name: "g"}
	if err := repositories.GetDB().Create(&group).Error; err != nil {
		t.Fatalf("failed to create group: %v", err)
	}
	allowedTag := models.Tag{Name: "allowed"}
	hiddenTag := models.Tag{Name: "hidden"}
	if err := repositories.GetDB().Create(&allowedTag).Error; err != nil {
		t.Fatalf("failed to create tag: %v", err)
	}
	if err := repositories.GetDB().Create(&hiddenTag).Error; err != nil {
		t.Fatalf("failed to create tag: %v", err)
	}

	addGroupMemberWithGrants(t, user.ID, group.ID, "tag-grant-role",
		[]string{permissions.GroupReceiptsRead}, nil, []uint{allowedTag.ID}, nil, false)

	_, out, err := handleListTags(context.Background(), requestForUser(user.ID), emptyInput{})
	if err != nil {
		t.Fatalf("handleListTags returned error: %v", err)
	}
	tags, ok := out.([]models.Tag)
	if !ok {
		t.Fatalf("expected []models.Tag, got %T", out)
	}
	if len(tags) != 1 || tags[0].ID != allowedTag.ID {
		t.Errorf("expected only the granted tag, got %+v", tags)
	}
}

func TestGetReceiptEnforcesGroupAccess(t *testing.T) {
	defer repositories.TruncateTestDb()

	member := createUser(t, "member")
	outsider := createUser(t, "outsider")

	group := models.Group{Name: "g1"}
	if err := repositories.GetDB().Create(&group).Error; err != nil {
		t.Fatalf("failed to create group: %v", err)
	}
	addGroupMember(t, member.ID, group.ID, []string{permissions.GroupReceiptsRead})

	receipt := createReceiptInGroup(t, "Lunch", group.ID, member.ID)
	receiptId := utils.UintToString(receipt.ID)

	// A member of the group can read the receipt.
	_, out, err := handleGetReceipt(context.Background(), requestForUser(member.ID), getReceiptInput{Id: receiptId})
	if err != nil {
		t.Fatalf("group member could not read receipt: %v", err)
	}
	if got, ok := out.(models.Receipt); !ok || got.ID != receipt.ID {
		t.Errorf("expected the requested receipt to be returned")
	}

	// A non-member is denied, with an error that does not confirm existence.
	_, _, err = handleGetReceipt(context.Background(), requestForUser(outsider.ID), getReceiptInput{Id: receiptId})
	if err == nil {
		t.Fatalf("expected non-member to be denied access to the receipt")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected a non-leaking 'not found' error, got %v", err)
	}
}

func TestSearchReceiptsScopesToUserGroups(t *testing.T) {
	defer repositories.TruncateTestDb()

	user := createUser(t, "searcher")

	memberGroup := models.Group{Name: "mine"}
	otherGroup := models.Group{Name: "theirs"}
	if err := repositories.GetDB().Create(&memberGroup).Error; err != nil {
		t.Fatalf("failed to create member group: %v", err)
	}
	if err := repositories.GetDB().Create(&otherGroup).Error; err != nil {
		t.Fatalf("failed to create other group: %v", err)
	}
	setAppRole(t, user.ID, "searcher-role", []string{permissions.AppReceiptsSearch})
	addGroupMember(t, user.ID, memberGroup.ID, []string{permissions.GroupReceiptsRead})

	createReceiptInGroup(t, "Coffee shop", memberGroup.ID, user.ID)
	createReceiptInGroup(t, "Coffee beans", memberGroup.ID, user.ID)
	createReceiptInGroup(t, "Coffee elsewhere", otherGroup.ID, user.ID)

	_, out, err := handleSearchReceipts(context.Background(), requestForUser(user.ID), searchReceiptsInput{Query: "Coffee"})
	if err != nil {
		t.Fatalf("handleSearchReceipts returned error: %v", err)
	}

	results, ok := out.([]structs.SearchResult)
	if !ok {
		t.Fatalf("expected []structs.SearchResult, got %T", out)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 receipts from the user's group, got %d", len(results))
	}
	for _, result := range results {
		if result.GroupID != memberGroup.ID {
			t.Errorf("search returned a receipt from group %d outside the user's groups", result.GroupID)
		}
	}
}

func TestGetReceiptStripsCategoriesAndTags(t *testing.T) {
	defer repositories.TruncateTestDb()
	db := repositories.GetDB()

	user := createUser(t, "stripuser")
	group := models.Group{Name: "g"}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("failed to create group: %v", err)
	}

	allowedCategory := models.Category{Name: "allowed-cat"}
	hiddenCategory := models.Category{Name: "hidden-cat"}
	allowedTag := models.Tag{Name: "allowed-tag"}
	hiddenTag := models.Tag{Name: "hidden-tag"}
	for _, m := range []interface{}{&allowedCategory, &hiddenCategory, &allowedTag, &hiddenTag} {
		if err := db.Create(m).Error; err != nil {
			t.Fatalf("failed to create fixture: %v", err)
		}
	}

	// The role grants read plus exactly one category and one tag.
	addGroupMemberWithGrants(t, user.ID, group.ID, "strip-role",
		[]string{permissions.GroupReceiptsRead},
		[]uint{allowedCategory.ID}, []uint{allowedTag.ID}, nil, false)

	receipt := createReceiptInGroup(t, "lunch", group.ID, user.ID)
	if err := db.Model(&receipt).Association("Categories").Append([]models.Category{allowedCategory, hiddenCategory}); err != nil {
		t.Fatalf("failed to attach categories: %v", err)
	}
	if err := db.Model(&receipt).Association("Tags").Append([]models.Tag{allowedTag, hiddenTag}); err != nil {
		t.Fatalf("failed to attach tags: %v", err)
	}

	_, out, err := handleGetReceipt(context.Background(), requestForUser(user.ID), getReceiptInput{Id: utils.UintToString(receipt.ID)})
	if err != nil {
		t.Fatalf("handleGetReceipt returned error: %v", err)
	}
	got, ok := out.(models.Receipt)
	if !ok {
		t.Fatalf("expected models.Receipt, got %T", out)
	}
	if len(got.Categories) != 1 || got.Categories[0].ID != allowedCategory.ID {
		t.Errorf("expected only the granted category, got %+v", got.Categories)
	}
	if len(got.Tags) != 1 || got.Tags[0].ID != allowedTag.ID {
		t.Errorf("expected only the granted tag, got %+v", got.Tags)
	}
}

func TestGetReceiptEnforcesPaidByVisibility(t *testing.T) {
	defer repositories.TruncateTestDb()
	db := repositories.GetDB()

	user := createUser(t, "pbviewer")
	allowedPayer := createUser(t, "allowedpayer")
	hiddenPayer := createUser(t, "hiddenpayer")

	group := models.Group{Name: "g"}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("failed to create group: %v", err)
	}
	// Restrict the member to receipts paid by allowedPayer only.
	addGroupMemberWithGrants(t, user.ID, group.ID, "pb-role",
		[]string{permissions.GroupReceiptsRead}, nil, nil, []uint{allowedPayer.ID}, false)

	allowedReceipt := createReceiptInGroup(t, "allowed", group.ID, allowedPayer.ID)
	hiddenReceipt := createReceiptInGroup(t, "hidden", group.ID, hiddenPayer.ID)

	// A receipt paid by an allowed payer is visible.
	if _, _, err := handleGetReceipt(context.Background(), requestForUser(user.ID), getReceiptInput{Id: utils.UintToString(allowedReceipt.ID)}); err != nil {
		t.Fatalf("expected the allowed receipt to be visible: %v", err)
	}

	// A receipt paid by a hidden payer reports the non-leaking "not found".
	_, _, err := handleGetReceipt(context.Background(), requestForUser(user.ID), getReceiptInput{Id: utils.UintToString(hiddenReceipt.ID)})
	if err == nil {
		t.Fatalf("expected the paid-by-hidden receipt to be denied")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected a non-leaking 'not found' error, got %v", err)
	}
}

func TestSearchReceiptsEnforcesPaidByVisibility(t *testing.T) {
	defer repositories.TruncateTestDb()
	db := repositories.GetDB()

	user := createUser(t, "pbsearcher")
	allowedPayer := createUser(t, "spallowed")
	hiddenPayer := createUser(t, "sphidden")

	group := models.Group{Name: "g"}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("failed to create group: %v", err)
	}
	setAppRole(t, user.ID, "pb-search-app-role", []string{permissions.AppReceiptsSearch})
	addGroupMemberWithGrants(t, user.ID, group.ID, "pb-search-role",
		[]string{permissions.GroupReceiptsRead}, nil, nil, []uint{allowedPayer.ID}, false)

	createReceiptInGroup(t, "Coffee allowed", group.ID, allowedPayer.ID)
	createReceiptInGroup(t, "Coffee hidden", group.ID, hiddenPayer.ID)

	_, out, err := handleSearchReceipts(context.Background(), requestForUser(user.ID), searchReceiptsInput{Query: "Coffee"})
	if err != nil {
		t.Fatalf("handleSearchReceipts returned error: %v", err)
	}
	results, ok := out.([]structs.SearchResult)
	if !ok {
		t.Fatalf("expected []structs.SearchResult, got %T", out)
	}
	if len(results) != 1 || results[0].PaidByUserId != allowedPayer.ID {
		t.Errorf("expected only the receipt paid by the allowed payer, got %+v", results)
	}
}

func TestSearchReceiptsRequiresGroupReadPermission(t *testing.T) {
	defer repositories.TruncateTestDb()

	// The mirror of the app-permission case below: the caller HAS app.receipts.search
	// and is a member of the group, but their group role lacks group.receipts.read.
	// Membership alone must not surface the receipt. Unlike a missing search
	// permission this is not a denial — an empty result is the correct answer.
	user := createUser(t, "nogroupread")
	group := models.Group{Name: "g"}
	if err := repositories.GetDB().Create(&group).Error; err != nil {
		t.Fatalf("failed to create group: %v", err)
	}
	setAppRole(t, user.ID, "nogroupread-app-role", []string{permissions.AppReceiptsSearch})
	addGroupMember(t, user.ID, group.ID, []string{})
	createReceiptInGroup(t, "Coffee", group.ID, user.ID)

	_, out, err := handleSearchReceipts(context.Background(), requestForUser(user.ID), searchReceiptsInput{Query: "Coffee"})
	if err != nil {
		t.Fatalf("handleSearchReceipts returned error: %v", err)
	}

	results, ok := out.([]structs.SearchResult)
	if !ok {
		t.Fatalf("expected []structs.SearchResult, got %T", out)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results without group.receipts.read, got %d (%+v)", len(results), results)
	}
}

func TestSearchReceiptsRequiresSearchPermission(t *testing.T) {
	defer repositories.TruncateTestDb()

	// Group read access and a matching receipt exist, but no app role -> no
	// app.receipts.search, so the denial is specifically tied to the missing
	// search permission rather than a lack of data or membership.
	user := createUser(t, "nosearch")
	group := models.Group{Name: "g"}
	if err := repositories.GetDB().Create(&group).Error; err != nil {
		t.Fatalf("failed to create group: %v", err)
	}
	addGroupMember(t, user.ID, group.ID, []string{permissions.GroupReceiptsRead})
	createReceiptInGroup(t, "anything", group.ID, user.ID)

	_, _, err := handleSearchReceipts(context.Background(), requestForUser(user.ID), searchReceiptsInput{Query: "anything"})
	if err == nil {
		t.Fatalf("expected a user without app.receipts.search to be denied")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Errorf("expected an 'unauthorized' error, got %v", err)
	}
}
