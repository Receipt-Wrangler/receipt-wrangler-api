package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"strings"
	"testing"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

func groupSetup() (http.HandlerFunc, *http.Request, *httptest.ResponseRecorder) {
	createUserAndGroup()
	fakeHandler, r, w := createFakeGroupHandler()

	return fakeHandler, r, w
}

func createUserAndGroup() {
	user := models.User{
		Username:    "test",
		Password:    "Password",
		DisplayName: "test",
	}
	db := db.GetDB()
	db.Create(&user)

	groupMembers := make([]models.GroupMember, 1)
	groupMembers = append(groupMembers, models.GroupMember{UserID: user.ID, GroupRole: models.OWNER})

	group := models.Group{
		Name:         "Test",
		GroupMembers: groupMembers,
	}

	db.Create(&group)
}

func createFakeGroupHandler() (http.HandlerFunc, *http.Request, *httptest.ResponseRecorder) {
	reader := strings.NewReader("")
	r := httptest.NewRequest(http.MethodGet, "/api/1", reader)
	w := httptest.NewRecorder()

	var vClaims validator.ValidatedClaims
	vClaims.CustomClaims = &utils.Claims{UserId: 1}

	ctx := context.WithValue(r.Context(), "groupId", "1")
	ctx = context.WithValue(ctx, jwtmiddleware.ContextKey{}, &vClaims)
	r = r.WithContext(ctx)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), r, w
}

func teardownGroupTest() {
	db := db.GetDB()
	utils.TruncateTable(db, "group_members")
	utils.TruncateTable(db, "groups")
	utils.TruncateTable(db, "users")
}

func TestValidateGroupeRoleShouldAuthorize1(t *testing.T) {
	fakeHandler, r, w := groupSetup()
	db.GetDB().Model(models.GroupMember{}).Where("group_id = ? AND user_id = ?", "1", "1").Update("group_role", models.VIEWER)

	mw := ValidateGroupRole(models.VIEWER)
	handler := mw(fakeHandler)
	handler.ServeHTTP(w, r)

	teardownGroupTest()

	if w.Result().StatusCode != 200 {
		utils.PrintTestError(t, w.Result().StatusCode, 200)
	}
}

func TestValidateGroupeRoleShouldAuthorize2(t *testing.T) {
	fakeHandler, r, w := groupSetup()
	db.GetDB().Model(models.GroupMember{}).Where("group_id = ? AND user_id = ?", "1", "1").Update("group_role", models.EDITOR)

	mw := ValidateGroupRole(models.VIEWER)
	handler := mw(fakeHandler)
	handler.ServeHTTP(w, r)

	teardownGroupTest()

	if w.Result().StatusCode != 200 {
		utils.PrintTestError(t, w.Result().StatusCode, 200)
	}
}

func TestValidateGroupeRoleShouldAuthorize3(t *testing.T) {
	fakeHandler, r, w := groupSetup()
	db.GetDB().Model(models.GroupMember{}).Where("group_id = ? AND user_id = ?", "1", "1").Update("group_role", models.OWNER)

	mw := ValidateGroupRole(models.VIEWER)
	handler := mw(fakeHandler)
	handler.ServeHTTP(w, r)

	teardownGroupTest()

	if w.Result().StatusCode != 200 {
		utils.PrintTestError(t, w.Result().StatusCode, 200)
	}
}

func TestValidateGroupeRoleShouldAuthorize4(t *testing.T) {
	fakeHandler, r, w := groupSetup()
	db.GetDB().Model(models.GroupMember{}).Where("group_id = ? AND user_id = ?", "1", "1").Update("group_role", models.OWNER)

	mw := ValidateGroupRole(models.EDITOR)
	handler := mw(fakeHandler)
	handler.ServeHTTP(w, r)

	teardownGroupTest()

	if w.Result().StatusCode != 200 {
		utils.PrintTestError(t, w.Result().StatusCode, 200)
	}
}

func TestValidateGroupeRoleShouldAuthorize5(t *testing.T) {
	fakeHandler, r, w := groupSetup()
	db.GetDB().Model(models.GroupMember{}).Where("group_id = ? AND user_id = ?", "1", "1").Update("group_role", models.EDITOR)

	mw := ValidateGroupRole(models.EDITOR)
	handler := mw(fakeHandler)
	handler.ServeHTTP(w, r)

	teardownGroupTest()

	if w.Result().StatusCode != 200 {
		utils.PrintTestError(t, w.Result().StatusCode, 200)
	}
}

func TestValidateGroupeRoleShouldAuthorize6(t *testing.T) {
	fakeHandler, r, w := groupSetup()
	db.GetDB().Model(models.GroupMember{}).Where("group_id = ? AND user_id = ?", "1", "1").Update("group_role", models.OWNER)

	mw := ValidateGroupRole(models.OWNER)
	handler := mw(fakeHandler)
	handler.ServeHTTP(w, r)

	teardownGroupTest()

	if w.Result().StatusCode != 200 {
		utils.PrintTestError(t, w.Result().StatusCode, 200)
	}
}

func TestValidateGroupeRoleShouldDeny1(t *testing.T) {
	fakeHandler, r, w := groupSetup()
	db.GetDB().Model(models.GroupMember{}).Where("group_id = ? AND user_id = ?", "1", "1").Update("group_role", models.VIEWER)

	mw := ValidateGroupRole(models.OWNER)
	handler := mw(fakeHandler)
	handler.ServeHTTP(w, r)

	teardownGroupTest()

	if w.Result().StatusCode != 403 {
		utils.PrintTestError(t, w.Result().StatusCode, 403)
	}
}

func TestValidateGroupeRoleShouldDeny2(t *testing.T) {
	fakeHandler, r, w := groupSetup()
	db.GetDB().Model(models.GroupMember{}).Where("group_id = ? AND user_id = ?", "1", "1").Update("group_role", models.EDITOR)

	mw := ValidateGroupRole(models.OWNER)
	handler := mw(fakeHandler)
	handler.ServeHTTP(w, r)

	teardownGroupTest()

	if w.Result().StatusCode != 403 {
		utils.PrintTestError(t, w.Result().StatusCode, 403)
	}
}

func TestValidateGroupeRoleShouldDeny3(t *testing.T) {
	fakeHandler, r, w := groupSetup()
	db.GetDB().Model(models.GroupMember{}).Where("group_id = ? AND user_id = ?", "1", "1").Update("group_role", models.VIEWER)

	mw := ValidateGroupRole(models.EDITOR)
	handler := mw(fakeHandler)
	handler.ServeHTTP(w, r)

	teardownGroupTest()

	if w.Result().StatusCode != 403 {
		utils.PrintTestError(t, w.Result().StatusCode, 403)
	}
}

func TestCanDeleteGroupShouldReject1(t *testing.T) {
	fakeHandler, r, w := groupSetup()
	handler := CanDeleteGroup(fakeHandler)
	handler.ServeHTTP(w, r)

	teardownGroupTest()

	if w.Result().StatusCode != 500 {
		utils.PrintTestError(t, w.Result().StatusCode, 500)
	}
}

func TestCanDeleteGroupShouldReject2(t *testing.T) {
	fakeHandler, _, w := groupSetup()
	db.GetDB().Model(models.GroupMember{}).Where("group_id = ? AND user_id = ?", "1", "1").Update("group_role", models.OWNER)
	groupMembers := make([]models.GroupMember, 1)
	groupMembers = append(groupMembers, models.GroupMember{UserID: 1, GroupRole: models.OWNER})

	group := models.Group{
		Name:         "Another group",
		GroupMembers: groupMembers,
	}

	db.GetDB().Create(&group)

	CanDeleteGroup(fakeHandler)

	teardownGroupTest()

	if w.Result().StatusCode != 200 {
		utils.PrintTestError(t, w.Result().StatusCode, 200)
	}
}
