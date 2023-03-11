package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	db "receipt-wrangler/api/internal/database"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"strings"
	"testing"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

var r *http.Request
var w *httptest.ResponseRecorder
var fakeHandler http.Handler

func TestMain(m *testing.M) {
	code, err := run(m)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

func run(m *testing.M) (code int, err error) {
	os.Args = append(os.Args, "-env=test")
	config.SetConfigs()
	logging.InitLog()
	InitMiddlewareLogger()
	db.InitTestDb()
	db.MakeMigrations()
	setup()

	defer func() {
	}()

	return m.Run(), nil
}

func setup() {
	createUserAndGroup()
	createFakeHandler()
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

func createFakeHandler() {
	reader := strings.NewReader("")
	r = httptest.NewRequest(http.MethodGet, "/api/1", reader)
	w = httptest.NewRecorder()

	var vClaims validator.ValidatedClaims
	vClaims.CustomClaims = &utils.Claims{UserId: 1}

	ctx := context.WithValue(r.Context(), "groupId", "1")
	ctx = context.WithValue(ctx, jwtmiddleware.ContextKey{}, &vClaims)
	r = r.WithContext(ctx)

	fakeHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}

func TestValidateGroupeRoleShouldAuthorize1(t *testing.T) {
	createFakeHandler()
	db.GetDB().Model(models.GroupMember{}).Where("group_id = ? AND user_id = ?", "1", "1").Update("group_role", models.VIEWER)

	mw := ValidateGroupRole(models.VIEWER)
	handler := mw(fakeHandler)
	handler.ServeHTTP(w, r)

	if w.Result().StatusCode != 200 {
		utils.PrintTestError(t, w.Result().StatusCode, 200)
	}
}

func TestValidateGroupeRoleShouldAuthorize2(t *testing.T) {
	createFakeHandler()
	db.GetDB().Model(models.GroupMember{}).Where("group_id = ? AND user_id = ?", "1", "1").Update("group_role", models.EDITOR)

	mw := ValidateGroupRole(models.VIEWER)
	handler := mw(fakeHandler)
	handler.ServeHTTP(w, r)

	if w.Result().StatusCode != 200 {
		utils.PrintTestError(t, w.Result().StatusCode, 200)
	}
}

func TestValidateGroupeRoleShouldAuthorize3(t *testing.T) {
	createFakeHandler()
	db.GetDB().Model(models.GroupMember{}).Where("group_id = ? AND user_id = ?", "1", "1").Update("group_role", models.OWNER)

	mw := ValidateGroupRole(models.VIEWER)
	handler := mw(fakeHandler)
	handler.ServeHTTP(w, r)

	if w.Result().StatusCode != 200 {
		utils.PrintTestError(t, w.Result().StatusCode, 200)
	}
}

func TestValidateGroupeRoleShouldAuthorize4(t *testing.T) {
	createFakeHandler()
	db.GetDB().Model(models.GroupMember{}).Where("group_id = ? AND user_id = ?", "1", "1").Update("group_role", models.OWNER)

	mw := ValidateGroupRole(models.EDITOR)
	handler := mw(fakeHandler)
	handler.ServeHTTP(w, r)

	if w.Result().StatusCode != 200 {
		utils.PrintTestError(t, w.Result().StatusCode, 200)
	}
}

func TestValidateGroupeRoleShouldAuthorize5(t *testing.T) {
	createFakeHandler()
	db.GetDB().Model(models.GroupMember{}).Where("group_id = ? AND user_id = ?", "1", "1").Update("group_role", models.EDITOR)

	mw := ValidateGroupRole(models.EDITOR)
	handler := mw(fakeHandler)
	handler.ServeHTTP(w, r)

	if w.Result().StatusCode != 200 {
		utils.PrintTestError(t, w.Result().StatusCode, 200)
	}
}

func TestValidateGroupeRoleShouldAuthorize6(t *testing.T) {
	createFakeHandler()
	db.GetDB().Model(models.GroupMember{}).Where("group_id = ? AND user_id = ?", "1", "1").Update("group_role", models.OWNER)

	mw := ValidateGroupRole(models.OWNER)
	handler := mw(fakeHandler)
	handler.ServeHTTP(w, r)

	if w.Result().StatusCode != 200 {
		utils.PrintTestError(t, w.Result().StatusCode, 200)
	}
}

func TestValidateGroupeRoleShouldDeny1(t *testing.T) {
	createFakeHandler()
	db.GetDB().Model(models.GroupMember{}).Where("group_id = ? AND user_id = ?", "1", "1").Update("group_role", models.VIEWER)

	mw := ValidateGroupRole(models.OWNER)
	handler := mw(fakeHandler)
	handler.ServeHTTP(w, r)

	if w.Result().StatusCode != 403 {
		utils.PrintTestError(t, w.Result().StatusCode, 403)
	}
}

func TestValidateGroupeRoleShouldDeny2(t *testing.T) {
	createFakeHandler()
	db.GetDB().Model(models.GroupMember{}).Where("group_id = ? AND user_id = ?", "1", "1").Update("group_role", models.EDITOR)

	mw := ValidateGroupRole(models.OWNER)
	handler := mw(fakeHandler)
	handler.ServeHTTP(w, r)

	if w.Result().StatusCode != 403 {
		utils.PrintTestError(t, w.Result().StatusCode, 403)
	}
}

func TestValidateGroupeRoleShouldDeny3(t *testing.T) {
	createFakeHandler()
	db.GetDB().Model(models.GroupMember{}).Where("group_id = ? AND user_id = ?", "1", "1").Update("group_role", models.VIEWER)

	mw := ValidateGroupRole(models.EDITOR)
	handler := mw(fakeHandler)
	handler.ServeHTTP(w, r)

	if w.Result().StatusCode != 403 {
		utils.PrintTestError(t, w.Result().StatusCode, 403)
	}
}
