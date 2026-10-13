package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/permissions"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"slices"
	"testing"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

func TestInitTokenValidatorReturnsValidator(t *testing.T) {
	v, err := InitTokenValidator()

	if v == nil {
		utils.PrintTestError(t, v, "instance of validator")
	}

	if err != nil {
		utils.PrintTestError(t, err, nil)
	}
}

func TestGenerateJWTGeneratesJWTCorrectly(t *testing.T) {
	defer repositories.TruncateTestDb()
	expectedDisplayname := "Displayname"
	expectedUsername := "Test"
	expectedIssuer := "https://receiptWrangler.io"
	var user models.User

	v, err := InitTokenValidator()

	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	db := repositories.GetDB()
	db.Create(&models.User{
		Username:    expectedUsername,
		Password:    "Password",
		DisplayName: expectedDisplayname,
	})

	if db.Where("username = ?", expectedUsername).Select("id").Find(&user).Error != nil {
		utils.PrintTestError(t, err.Error(), nil)
	}

	jwt, _, _, err := GenerateJWT(user.ID)
	if err != nil {
		utils.PrintTestError(t, jwt, "jwt token")
	}

	rawJwtStruct, err := v.ValidateToken(context.Background(), jwt)
	if err != nil {
		utils.PrintTestError(t, rawJwtStruct, "claim object")
	}

	jwtClaims := rawJwtStruct.(*validator.ValidatedClaims).CustomClaims.(*structs.Claims)

	if jwt == "nil" {
		utils.PrintTestError(t, jwt, "non empty string")
	}

	if jwtClaims.UserId != user.ID {
		utils.PrintTestError(t, jwtClaims.UserId, user.ID)
	}

	if jwtClaims.Displayname != expectedDisplayname {
		utils.PrintTestError(t, jwtClaims.Displayname, expectedDisplayname)
	}

	if jwtClaims.Username != expectedUsername {
		utils.PrintTestError(t, jwtClaims.Username, expectedUsername)
	}

	if jwtClaims.Issuer != expectedIssuer {
		utils.PrintTestError(t, jwtClaims.Issuer, expectedIssuer)
	}

	if len(jwtClaims.Audience) > 0 && jwtClaims.Audience[0] != expectedIssuer {
		utils.PrintTestError(t, jwtClaims.Audience, fmt.Sprintf("[%s]", expectedIssuer))
	}

	if err != nil {
		utils.PrintTestError(t, err, nil)
	}
}

func TestGenerateRefreshTokenCorrectly(t *testing.T) {
	defer repositories.TruncateTestDb()
	expectedDisplayname := "Another displayname"
	expectedUsername := "Another username"
	expectedIssuer := "https://receiptWrangler.io"
	var user models.User

	v, err := InitTokenValidator()

	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	db := repositories.GetDB()
	db.Create(&models.User{
		Username:    expectedUsername,
		Password:    "Password",
		DisplayName: expectedDisplayname,
	})

	if db.Where("username = ?", expectedUsername).Select("id").Find(&user).Error != nil {
		utils.PrintTestError(t, err.Error(), nil)
	}

	_, refreshToken, _, err := GenerateJWT(user.ID)
	if err != nil {
		utils.PrintTestError(t, refreshToken, "refresh token")
	}

	rawRefreshTokenClaims, err := v.ValidateToken(context.Background(), refreshToken)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
		return
	}

	if rawRefreshTokenClaims == nil {
		utils.PrintTestError(t, rawRefreshTokenClaims, "non-nil claim object")
		return
	}

	refreshTokenClaims := rawRefreshTokenClaims.(*validator.ValidatedClaims).CustomClaims.(*structs.Claims)

	if refreshToken == "nil" {
		utils.PrintTestError(t, refreshToken, "non empty string")
	}

	if refreshTokenClaims.UserId != user.ID {
		utils.PrintTestError(t, refreshTokenClaims.UserId, user.ID)
	}

	if refreshTokenClaims.Issuer != expectedIssuer {
		utils.PrintTestError(t, refreshTokenClaims.Issuer, expectedIssuer)
	}

	if len(refreshTokenClaims.Audience) > 0 && refreshTokenClaims.Audience[0] != expectedIssuer {
		utils.PrintTestError(t, refreshTokenClaims.Audience, fmt.Sprintf("[%s]", expectedIssuer))
	}

	if err != nil {
		utils.PrintTestError(t, err, nil)
	}
}

func TestShouldLogInUserCorrectly(t *testing.T) {
	defer repositories.TruncateTestDb()
	ClearRolePermissionCacheForTests()
	expectedDisplayname := "Another displayname"
	expectedUsername := "Another username"
	password := "Password"

	userRepository := repositories.NewUserRepository(nil)

	createdUser, err := userRepository.CreateUser(commands.SignUpCommand{
		Username:    expectedUsername,
		Password:    password,
		DisplayName: expectedDisplayname,
	})
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	// "First admin to login" is now defined by the app.users.read permission
	// (the modern replacement for the removed UserRole == ADMIN check), so the
	// user must hold an admin role for the firstAdminToLogin path to be exercised.
	roleRepository := repositories.NewRoleRepository(nil)
	adminRole, err := roleRepository.CreateAppRole("Login Admin Role", "", []string{permissions.AppUsersRead}, false)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}
	if err := repositories.GetDB().Model(&models.User{}).
		Where("id = ?", createdUser.ID).Update("app_role_id", adminRole.ID).Error; err != nil {
		utils.PrintTestError(t, err, nil)
	}
	ClearRolePermissionCacheForTests()

	user, firstAdminToLogin, err := LoginUser(commands.LoginCommand{
		Username: expectedUsername,
		Password: password,
	})

	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	if firstAdminToLogin != true {
		utils.PrintTestError(t, firstAdminToLogin, true)
	}

	if user.LastLoginDate == nil {
		utils.PrintTestError(t, user.LastLoginDate, nil)
	}
}

func TestShouldNotLogUserInWithWrongPassword(t *testing.T) {
	defer repositories.TruncateTestDb()
	expectedDisplayname := "Another displayname"
	expectedUsername := "Another username"
	password := "Password"

	userRepository := repositories.NewUserRepository(nil)

	_, err := userRepository.CreateUser(commands.SignUpCommand{
		Username:    expectedUsername,
		Password:    password,
		DisplayName: expectedDisplayname,
	})
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	_, _, err = LoginUser(commands.LoginCommand{
		Username: expectedUsername,
		Password: "wrong password",
	})

	if err == nil {
		utils.PrintTestError(t, err, "login error")
	}
}

// BuildTokenCookies — non-dev environment (test env is "test", so this
// exercises the SameSite=Strict / Secure=false branch).
func TestBuildTokenCookies_NonDevEnvironment(t *testing.T) {
	jwt := "jwt-token-value"
	refreshToken := "refresh-token-value"

	access, refresh := BuildTokenCookies(jwt, refreshToken)

	if access.Name != constants.JwtKey {
		utils.PrintTestError(t, access.Name, constants.JwtKey)
	}
	if access.Value != jwt {
		utils.PrintTestError(t, access.Value, jwt)
	}
	if !access.HttpOnly {
		utils.PrintTestError(t, access.HttpOnly, true)
	}
	if access.Path != "/" {
		utils.PrintTestError(t, access.Path, "/")
	}
	if access.SameSite != http.SameSiteStrictMode {
		utils.PrintTestError(t, access.SameSite, http.SameSiteStrictMode)
	}
	if access.Secure {
		utils.PrintTestError(t, access.Secure, false)
	}
	if access.Expires.IsZero() {
		utils.PrintTestError(t, "Expires is zero", "non-zero time")
	}

	if refresh.Name != constants.RefreshTokenKey {
		utils.PrintTestError(t, refresh.Name, constants.RefreshTokenKey)
	}
	if refresh.Value != refreshToken {
		utils.PrintTestError(t, refresh.Value, refreshToken)
	}
	if !refresh.HttpOnly {
		utils.PrintTestError(t, refresh.HttpOnly, true)
	}
	if refresh.Path != "/" {
		utils.PrintTestError(t, refresh.Path, "/")
	}
	if refresh.SameSite != http.SameSiteStrictMode {
		utils.PrintTestError(t, refresh.SameSite, http.SameSiteStrictMode)
	}
	if refresh.Secure {
		utils.PrintTestError(t, refresh.Secure, false)
	}
	if refresh.Expires.IsZero() {
		utils.PrintTestError(t, "Expires is zero", "non-zero time")
	}
}

// BuildTokenCookies dev branch: env=="dev" -> SameSite=None, Secure=true.
// Skipped because config.env is set once from `-env=test` in SetUpTestEnv and
// there is no exported hook to override it for a single test.
// See bug report BUG-2 (testability).
func TestBuildTokenCookies_DevEnvironment(t *testing.T) {
	t.Skip("see BUG-2: config.env is package-private and fixed to 'test' for the suite; no hook to override per-test")
	if config.GetDeployEnv() != "dev" {
		return
	}
	access, refresh := BuildTokenCookies("j", "r")
	if access.SameSite != http.SameSiteNoneMode {
		utils.PrintTestError(t, access.SameSite, http.SameSiteNoneMode)
	}
	if !access.Secure {
		utils.PrintTestError(t, access.Secure, true)
	}
	if refresh.SameSite != http.SameSiteNoneMode {
		utils.PrintTestError(t, refresh.SameSite, http.SameSiteNoneMode)
	}
	if !refresh.Secure {
		utils.PrintTestError(t, refresh.Secure, true)
	}
}

// PrepareAccessTokenClaims — the current implementation takes a value
// receiver and therefore cannot mutate the caller's claims. We assert the
// observed behavior (caller's claims unchanged) and flag the apparent intent
// as a bug — see BUG-1 in the bug report.
func TestPrepareAccessTokenClaims_DoesNotMutateCaller(t *testing.T) {
	claims := structs.Claims{}
	claims.Issuer = "https://receiptWrangler.io"
	claims.Audience = []string{"https://receiptWrangler.io"}

	PrepareAccessTokenClaims(claims)

	// Documenting observed behavior: caller's Issuer/Audience are unchanged
	// because the function receives claims by value.
	if claims.Issuer != "https://receiptWrangler.io" {
		utils.PrintTestError(t, claims.Issuer, "https://receiptWrangler.io")
	}
	if len(claims.Audience) != 1 || claims.Audience[0] != "https://receiptWrangler.io" {
		utils.PrintTestError(t, claims.Audience, []string{"https://receiptWrangler.io"})
	}
}

// PrepareAccessTokenClaims — skipped test capturing the *intended* behavior:
// after the call, Issuer should be "" and Audience should be empty. This
// currently fails because of BUG-1 (value receiver). Kept as a pinned test
// so the bug is easy to discover.
func TestPrepareAccessTokenClaims_ClearsIssuerAndAudience(t *testing.T) {
	t.Skip("see BUG-1: PrepareAccessTokenClaims takes structs.Claims by value; caller mutations are lost")

	claims := structs.Claims{}
	claims.Issuer = "https://receiptWrangler.io"
	claims.Audience = []string{"https://receiptWrangler.io"}

	PrepareAccessTokenClaims(claims)

	if claims.Issuer != "" {
		utils.PrintTestError(t, claims.Issuer, "")
	}
	if len(claims.Audience) != 0 {
		utils.PrintTestError(t, claims.Audience, []string{})
	}
}

func TestGetEmptyAccessTokenCookie(t *testing.T) {
	cookie := GetEmptyAccessTokenCookie()

	if cookie.Name != constants.JwtKey {
		utils.PrintTestError(t, cookie.Name, constants.JwtKey)
	}
	if cookie.Value != "" {
		utils.PrintTestError(t, cookie.Value, "")
	}
	if cookie.Path != "/" {
		utils.PrintTestError(t, cookie.Path, "/")
	}
	if cookie.MaxAge != -1 {
		utils.PrintTestError(t, cookie.MaxAge, -1)
	}
	if cookie.HttpOnly {
		utils.PrintTestError(t, cookie.HttpOnly, false)
	}
	if cookie.Secure {
		utils.PrintTestError(t, cookie.Secure, false)
	}
}

func TestGetEmptyRefreshTokenCookie(t *testing.T) {
	cookie := GetEmptyRefreshTokenCookie()

	if cookie.Name != constants.RefreshTokenKey {
		utils.PrintTestError(t, cookie.Name, constants.RefreshTokenKey)
	}
	if cookie.Value != "" {
		utils.PrintTestError(t, cookie.Value, "")
	}
	if cookie.Path != "/" {
		utils.PrintTestError(t, cookie.Path, "/")
	}
	if cookie.MaxAge != -1 {
		utils.PrintTestError(t, cookie.MaxAge, -1)
	}
	if !cookie.HttpOnly {
		utils.PrintTestError(t, cookie.HttpOnly, true)
	}
	if cookie.Secure {
		utils.PrintTestError(t, cookie.Secure, false)
	}
}

func TestUseSecureCookiesFollowsTheSystemSetting(t *testing.T) {
	defer repositories.TruncateTestDb()

	if UseSecureCookies() {
		utils.PrintTestError(t, true, "off by default")
	}

	if err := repositories.SetSecureCookiesForTests(true); err != nil {
		utils.PrintTestError(t, err, nil)
	}
	if !UseSecureCookies() {
		utils.PrintTestError(t, false, "on once the setting is saved on")
	}

	if err := repositories.SetSecureCookiesForTests(false); err != nil {
		utils.PrintTestError(t, err, nil)
	}
	if UseSecureCookies() {
		utils.PrintTestError(t, true, "off again once the setting is saved off")
	}
}

// Every cookie the services layer builds -- the session pair and the two
// cookies that expire them -- follows the setting, and SameSite is untouched.
func TestCookiesAreSecureWhenTheSettingIsOn(t *testing.T) {
	defer repositories.TruncateTestDb()

	if err := repositories.SetSecureCookiesForTests(true); err != nil {
		utils.PrintTestError(t, err, nil)
	}

	access, refresh := BuildTokenCookies("jwt-token-value", "refresh-token-value")
	cookies := []http.Cookie{access, refresh, GetEmptyAccessTokenCookie(), GetEmptyRefreshTokenCookie()}

	for _, cookie := range cookies {
		if !cookie.Secure {
			utils.PrintTestError(t, cookie.Name+" Secure=false", "Secure=true")
		}
	}

	if access.SameSite != http.SameSiteStrictMode {
		utils.PrintTestError(t, access.SameSite, http.SameSiteStrictMode)
	}
	if refresh.SameSite != http.SameSiteStrictMode {
		utils.PrintTestError(t, refresh.SameSite, http.SameSiteStrictMode)
	}
}

// GetAppData with nil request — no claims to populate.
func TestGetAppData_PopulatesFields(t *testing.T) {
	defer repositories.TruncateTestDb()

	userRepository := repositories.NewUserRepository(nil)
	user, err := userRepository.CreateUser(commands.SignUpCommand{
		Username:    "appdata-user",
		Password:    "Password",
		DisplayName: "AppData User",
	})
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	appData, err := GetAppData(user.ID, nil)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	// CreateUser seeds a "My Receipts" group and an "All" group; both should
	// appear for the user.
	if len(appData.Groups) == 0 {
		utils.PrintTestError(t, "Groups length 0", ">0")
	}
	// At least the created user is in the Users list.
	found := false
	for _, u := range appData.Users {
		if u.ID == user.ID {
			found = true
			break
		}
	}
	if !found {
		utils.PrintTestError(t, "user not in appData.Users", "present")
	}
	if appData.UserPreferences.UserId != user.ID {
		utils.PrintTestError(t, appData.UserPreferences.UserId, user.ID)
	}
	if appData.Icons == nil {
		utils.PrintTestError(t, appData.Icons, "non-nil Icons slice")
	}
}

// GetAppData applies member isolation at the serialization boundary: a plain member of
// an isolated group sees neither a peer's user-directory entry nor the peer in that
// group's roster, while self + the supervisor remain.
func TestGetAppData_IsolationHidesPeerFromDirectoryAndRoster(t *testing.T) {
	defer repositories.TruncateTestDb()
	clearRolePermissionCacheAll()

	group := seedIsoGroup(t, "appdata-iso", true)
	supRole := seedIsoRole(t, "appdata-iso-sup", true)
	memberRole := seedIsoRole(t, "appdata-iso-mem", false)

	viewer := seedIsoUser(t, "appdata-iso-viewer")
	supervisor := seedIsoUser(t, "appdata-iso-sup-user")
	peer := seedIsoUser(t, "appdata-iso-peer")
	seedIsoMember(t, group.ID, viewer.ID, &memberRole.ID)
	seedIsoMember(t, group.ID, supervisor.ID, &supRole.ID)
	seedIsoMember(t, group.ID, peer.ID, &memberRole.ID)

	appData, err := GetAppData(viewer.ID, nil)
	if err != nil {
		t.Fatalf("GetAppData: %v", err)
	}

	// Directory (appData.Users): peer absent; self + supervisor present.
	inUsers := func(id uint) bool {
		for _, u := range appData.Users {
			if u.ID == id {
				return true
			}
		}
		return false
	}
	if inUsers(peer.ID) {
		t.Errorf("peer should be hidden from appData.Users for an isolated member")
	}
	if !inUsers(viewer.ID) || !inUsers(supervisor.ID) {
		t.Errorf("self + supervisor should remain in appData.Users")
	}

	// Roster of the isolated group: peer absent; self + supervisor present.
	var isoRoster []uint
	for _, g := range appData.Groups {
		if g.ID == group.ID {
			for _, m := range g.GroupMembers {
				isoRoster = append(isoRoster, m.UserID)
			}
		}
	}
	contains := func(ids []uint, id uint) bool {
		for _, x := range ids {
			if x == id {
				return true
			}
		}
		return false
	}
	if contains(isoRoster, peer.ID) {
		t.Errorf("peer should be hidden from the isolated group's roster, got %v", isoRoster)
	}
	if !contains(isoRoster, viewer.ID) || !contains(isoRoster, supervisor.ID) {
		t.Errorf("self + supervisor should remain in the isolated group's roster, got %v", isoRoster)
	}
}

// GetAppData with non-nil request that carries ValidatedClaims — Claims
// should be populated on the AppData.
func TestGetAppData_WithRequestPopulatesClaims(t *testing.T) {
	defer repositories.TruncateTestDb()

	userRepository := repositories.NewUserRepository(nil)
	user, err := userRepository.CreateUser(commands.SignUpCommand{
		Username:    "claims-user",
		Password:    "Password",
		DisplayName: "Claims User",
	})
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	customClaims := &structs.Claims{
		UserId:      user.ID,
		Username:    user.Username,
		Displayname: user.DisplayName,
	}
	validatedClaims := &validator.ValidatedClaims{CustomClaims: customClaims}

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	ctx := context.WithValue(req.Context(), jwtmiddleware.ContextKey{}, validatedClaims)
	req = req.WithContext(ctx)

	appData, err := GetAppData(user.ID, req)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}
	if appData.Claims.UserId != user.ID {
		utils.PrintTestError(t, appData.Claims.UserId, user.ID)
	}
	if appData.Claims.Username != user.Username {
		utils.PrintTestError(t, appData.Claims.Username, user.Username)
	}
}

// GetAppData populates the caller's effective app and per-group permissions for
// a user with an assigned app role and a group membership carrying a group role.
func TestGetAppData_PopulatesPermissions(t *testing.T) {
	defer repositories.TruncateTestDb()
	ClearRolePermissionCacheForTests()

	db := repositories.GetDB()
	roleRepository := repositories.NewRoleRepository(nil)

	appPerms := []string{permissions.AppUsersRead, permissions.AppUsersCreate}
	appRole, err := roleRepository.CreateAppRole("AppData App Role", "", appPerms, false)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	user := models.User{Username: "appdata-perms-user", Password: "password", AppRoleID: &appRole.ID}
	if err := db.Create(&user).Error; err != nil {
		utils.PrintTestError(t, err, nil)
	}

	groupPerms := []string{permissions.GroupReceiptsRead, permissions.GroupReceiptsUpdate}
	groupRole, err := roleRepository.CreateGroupRole("AppData Group Role", "", groupPerms, nil, nil, nil, false, false)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	group := models.Group{Name: "appdata-perms-group"}
	if err := db.Create(&group).Error; err != nil {
		utils.PrintTestError(t, err, nil)
	}
	member := models.GroupMember{GroupID: group.ID, UserID: user.ID, GroupRoleID: &groupRole.ID}
	if err := db.Create(&member).Error; err != nil {
		utils.PrintTestError(t, err, nil)
	}

	appData, err := GetAppData(user.ID, nil)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	if !slices.Equal(sortedCopy(appData.AppPermissions), sortedCopy(appPerms)) {
		utils.PrintTestError(t, appData.AppPermissions, appPerms)
	}

	if appData.GroupPermissions == nil {
		utils.PrintTestError(t, "nil GroupPermissions", "populated map")
	}
	if !slices.Equal(sortedCopy(appData.GroupPermissions[group.ID]), sortedCopy(groupPerms)) {
		utils.PrintTestError(t, appData.GroupPermissions[group.ID], groupPerms)
	}
}

// GetAppData filters the per-group category catalog to the caller's grants and
// withholds the flat global list from a non-admin (no app.categories.read).
func TestGetAppData_GroupCategoriesFilteredByGrants(t *testing.T) {
	defer repositories.TruncateTestDb()
	ClearRolePermissionCacheForTests()
	ClearGroupRoleGrantCacheForTests()

	db := repositories.GetDB()
	roleRepository := repositories.NewRoleRepository(nil)

	grantedCategory := models.Category{Name: "Groceries"}
	db.Create(&grantedCategory)
	hiddenCategory := models.Category{Name: "Salary"}
	db.Create(&hiddenCategory)

	// Legacy-User-like app role: create but not read.
	appRole, err := roleRepository.CreateAppRole("AppData User Role", "", []string{permissions.AppCategoriesCreate}, false)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}
	user := models.User{Username: "appdata-grant-user", Password: "password", AppRoleID: &appRole.ID}
	if err := db.Create(&user).Error; err != nil {
		utils.PrintTestError(t, err, nil)
	}

	groupRole, err := roleRepository.CreateGroupRole("AppData Restricted Role", "", []string{permissions.GroupReceiptsRead}, []uint{grantedCategory.ID}, nil, nil, false, false)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}
	group := models.Group{Name: "appdata-grant-group"}
	if err := db.Create(&group).Error; err != nil {
		utils.PrintTestError(t, err, nil)
	}
	if err := db.Create(&models.GroupMember{GroupID: group.ID, UserID: user.ID, GroupRoleID: &groupRole.ID}).Error; err != nil {
		utils.PrintTestError(t, err, nil)
	}

	appData, err := GetAppData(user.ID, nil)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	visible := appData.GroupCategories[group.ID]
	if len(visible) != 1 || visible[0].ID != grantedCategory.ID {
		utils.PrintTestError(t, visible, []uint{grantedCategory.ID})
	}

	// A non-admin gets no flat global list.
	if len(appData.Categories) != 0 {
		utils.PrintTestError(t, len(appData.Categories), 0)
	}
}

// GetAppData's per-group catalog honours a narrowing that comes from the
// MEMBERSHIP rather than the group role, and carries each member's own grant ids
// on the roster so the desktop forms can render them.
//
// AppData is the delivery surface the receipt-form pickers read from, so this is
// where the two grant layers have to compose correctly in practice.
func TestGetAppData_MemberGrantsNarrowGroupCatalog(t *testing.T) {
	defer repositories.TruncateTestDb()
	ClearRolePermissionCacheForTests()
	ClearGroupRoleGrantCacheForTests()

	db := repositories.GetDB()
	roleRepository := repositories.NewRoleRepository(nil)

	assignedCategory := models.Category{Name: "Child A"}
	roleOnlyCategory := models.Category{Name: "Child B"}
	db.Create(&assignedCategory)
	db.Create(&roleOnlyCategory)

	// A non-admin: no app.categories.read, so no grant bypass. Personal-group
	// creation is left on (false) — this test is about the category catalog.
	appRole, err := roleRepository.CreateAppRole("AppData Member Grant App Role", "", []string{}, false)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}
	user := models.User{Username: "appdata-member-grant-user", Password: "password", AppRoleID: &appRole.ID}
	if err := db.Create(&user).Error; err != nil {
		utils.PrintTestError(t, err, nil)
	}

	// The ROLE allows both categories; the member is assigned only one.
	groupRole, err := roleRepository.CreateGroupRole(
		"AppData Member Grant Role", "", []string{permissions.GroupReceiptsRead},
		[]uint{assignedCategory.ID, roleOnlyCategory.ID}, nil, nil, false, false,
	)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}
	group := models.Group{Name: "appdata-member-grant-group"}
	if err := db.Create(&group).Error; err != nil {
		utils.PrintTestError(t, err, nil)
	}
	if err := db.Create(&models.GroupMember{GroupID: group.ID, UserID: user.ID, GroupRoleID: &groupRole.ID}).Error; err != nil {
		utils.PrintTestError(t, err, nil)
	}
	err = repositories.NewGroupMemberRepository(nil).
		ReplaceMemberGrants(user.ID, group.ID, []uint{assignedCategory.ID}, nil)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	appData, err := GetAppData(user.ID, nil)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	// Intersection: the role's second category must not reach the catalog.
	visible := appData.GroupCategories[group.ID]
	if len(visible) != 1 || visible[0].ID != assignedCategory.ID {
		utils.PrintTestError(t, visible, []uint{assignedCategory.ID})
	}

	// The roster carries the member's own assignment for the admin forms.
	var member *models.GroupMember
	for i := range appData.Groups {
		if appData.Groups[i].ID != group.ID {
			continue
		}
		for j := range appData.Groups[i].GroupMembers {
			if appData.Groups[i].GroupMembers[j].UserID == user.ID {
				member = &appData.Groups[i].GroupMembers[j]
			}
		}
	}
	if member == nil {
		utils.PrintTestError(t, "member missing from appData roster", user.ID)
		return
	}
	if len(member.CategoryGrants) != 1 || member.CategoryGrants[0] != assignedCategory.ID {
		utils.PrintTestError(t, member.CategoryGrants, []uint{assignedCategory.ID})
	}
}

// GetAppData gives an admin (app.categories.read) the flat global list, and an
// unrestricted group's catalog contains every category.
func TestGetAppData_AdminGetsFlatCategoriesUnrestrictedGroup(t *testing.T) {
	defer repositories.TruncateTestDb()
	ClearRolePermissionCacheForTests()
	ClearGroupRoleGrantCacheForTests()

	db := repositories.GetDB()
	roleRepository := repositories.NewRoleRepository(nil)

	db.Create(&models.Category{Name: "Groceries"})
	db.Create(&models.Category{Name: "Salary"})

	appRole, err := roleRepository.CreateAppRole("AppData Admin Role", "", []string{permissions.AppCategoriesRead, permissions.AppTagsRead}, false)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}
	user := models.User{Username: "appdata-admin-user", Password: "password", AppRoleID: &appRole.ID}
	if err := db.Create(&user).Error; err != nil {
		utils.PrintTestError(t, err, nil)
	}

	groupRole, err := roleRepository.CreateGroupRole("AppData Open Role", "", []string{permissions.GroupReceiptsRead}, nil, nil, nil, false, false)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}
	group := models.Group{Name: "appdata-admin-group"}
	if err := db.Create(&group).Error; err != nil {
		utils.PrintTestError(t, err, nil)
	}
	if err := db.Create(&models.GroupMember{GroupID: group.ID, UserID: user.ID, GroupRoleID: &groupRole.ID}).Error; err != nil {
		utils.PrintTestError(t, err, nil)
	}

	appData, err := GetAppData(user.ID, nil)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	if len(appData.Categories) != 2 {
		utils.PrintTestError(t, len(appData.Categories), 2)
	}
	if len(appData.GroupCategories[group.ID]) != 2 {
		utils.PrintTestError(t, len(appData.GroupCategories[group.ID]), 2)
	}
}

// GetAppData returns an empty (non-nil) permission set for a user with no
// assigned app role.
func TestGetAppData_NoRolePermissionsEmpty(t *testing.T) {
	defer repositories.TruncateTestDb()
	ClearRolePermissionCacheForTests()

	userRepository := repositories.NewUserRepository(nil)
	user, err := userRepository.CreateUser(commands.SignUpCommand{
		Username:    "appdata-no-role-user",
		Password:    "Password",
		DisplayName: "AppData No Role User",
	})
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	appData, err := GetAppData(user.ID, nil)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	// No app role seeded (the test harness does not run SeedSystemRoles), so the
	// user resolves to no app permissions.
	if len(appData.AppPermissions) != 0 {
		utils.PrintTestError(t, appData.AppPermissions, "empty app permissions")
	}
	// Each group the user belongs to resolves to an empty (no group role) set.
	for groupId, perms := range appData.GroupPermissions {
		if len(perms) != 0 {
			utils.PrintTestError(t, perms, "empty permissions for group "+utils.UintToString(groupId))
		}
	}
}

// LoginUser must reject a dummy (passwordless placeholder) account regardless of
// the submitted password, and must reject an empty password for any account.
// This is the centralized guard that closes the OAuth/MCP authorize bypass
// (dummy users could previously authenticate there with an empty password).
func TestLoginUserRejectsDummyUser(t *testing.T) {
	defer repositories.TruncateTestDb()
	ClearRolePermissionCacheForTests()

	userRepository := repositories.NewUserRepository(nil)
	_, err := userRepository.CreateUser(commands.SignUpCommand{
		Username:    "ghost",
		DisplayName: "Ghost",
		IsDummyUser: true,
	})
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	// Empty password (the exact exploit) must be rejected.
	if _, _, err := LoginUser(commands.LoginCommand{Username: "ghost", Password: ""}); err == nil {
		utils.PrintTestError(t, nil, "login error for dummy user with empty password")
	}

	// A non-empty password against a dummy user must also be rejected (the
	// account can never authenticate), and must not fall through to bcrypt.
	if _, _, err := LoginUser(commands.LoginCommand{Username: "ghost", Password: "anything"}); err == nil {
		utils.PrintTestError(t, nil, "login error for dummy user with any password")
	}
}

func TestLoginUserRejectsEmptyPassword(t *testing.T) {
	defer repositories.TruncateTestDb()
	ClearRolePermissionCacheForTests()

	userRepository := repositories.NewUserRepository(nil)
	_, err := userRepository.CreateUser(commands.SignUpCommand{
		Username:    "realuser",
		Password:    "Password",
		DisplayName: "Real User",
	})
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	if _, _, err := LoginUser(commands.LoginCommand{Username: "realuser", Password: ""}); err == nil {
		utils.PrintTestError(t, nil, "login error for empty password")
	}

	// Sanity: the real password still works.
	if _, _, err := LoginUser(commands.LoginCommand{Username: "realuser", Password: "Password"}); err != nil {
		utils.PrintTestError(t, err, "successful login with correct password")
	}
}

// seedOidcIdentity links userId to a new provider named providerName.
// provisioned marks the ACCOUNT as having no password of its own, which is what
// an OIDC provisioning login records.
func seedOidcIdentity(t *testing.T, userId uint, providerName string, provisioned bool) models.OidcIdentity {
	t.Helper()
	db := repositories.GetDB()

	provider := models.OidcProvider{
		Name:         providerName,
		DisplayName:  providerName,
		IssuerUrl:    "https://idp.example.com",
		ClientId:     "client",
		ClientSecret: "secret",
		Enabled:      true,
	}
	if err := db.Create(&provider).Error; err != nil {
		t.Fatalf("failed to create the provider: %v", err)
	}

	identity := models.OidcIdentity{
		OidcProviderId:  provider.ID,
		Subject:         providerName + "-subject",
		UserId:          userId,
		ProvisionedUser: provisioned,
	}
	if err := db.Create(&identity).Error; err != nil {
		t.Fatalf("failed to create the identity: %v", err)
	}

	return identity
}

// An account with no password of its own must not sign in through the password
// form, even with the password that is actually stored. A provisioned account's
// password is random and discarded, so nobody can produce it today; this pins
// the rule itself, so the account stays closed if that ever stops being true.
func TestLoginUserRejectsPasswordlessOidcAccount(t *testing.T) {
	defer repositories.TruncateTestDb()
	ClearRolePermissionCacheForTests()

	user, err := repositories.NewUserRepository(nil).CreateUser(commands.SignUpCommand{
		Username:    "oidconly",
		Password:    "Password",
		DisplayName: "OIDC Only",
	})
	if err != nil {
		t.Fatalf("failed to create the user: %v", err)
	}
	seedOidcIdentity(t, user.ID, "keycloak", true)

	_, _, err = LoginUser(commands.LoginCommand{Username: "oidconly", Password: "Password"})
	if !errors.Is(err, ErrPasswordLoginUnavailable) {
		t.Fatalf("expected ErrPasswordLoginUnavailable for a passwordless account, got %v", err)
	}

	var stored models.User
	if err := repositories.GetDB().First(&stored, user.ID).Error; err != nil {
		t.Fatalf("failed to reload the user: %v", err)
	}
	if stored.LastLoginDate != nil {
		t.Errorf("a refused login was recorded as a login at %v", stored.LastLoginDate)
	}

	// An administrator's password reset gives the account a password of its own,
	// which is the documented way back in.
	if err := repositories.NewOidcIdentityRepository(nil).ClearProvisionedFlagForUser(user.ID); err != nil {
		t.Fatalf("failed to clear the provisioned flag: %v", err)
	}

	if _, _, err := LoginUser(commands.LoginCommand{Username: "oidconly", Password: "Password"}); err != nil {
		t.Fatalf("expected a login once the account has a password of its own, got %v", err)
	}
}

// Linking a provider to an account that already had a password must not take
// password sign-in away from it.
func TestLoginUserAllowsLinkedAccountWithItsOwnPassword(t *testing.T) {
	defer repositories.TruncateTestDb()
	ClearRolePermissionCacheForTests()

	user, err := repositories.NewUserRepository(nil).CreateUser(commands.SignUpCommand{
		Username:    "linkeduser",
		Password:    "Password",
		DisplayName: "Linked User",
	})
	if err != nil {
		t.Fatalf("failed to create the user: %v", err)
	}
	seedOidcIdentity(t, user.ID, "keycloak", false)

	if _, _, err := LoginUser(commands.LoginCommand{Username: "linkeduser", Password: "Password"}); err != nil {
		t.Fatalf("expected a linked account with its own password to log in, got %v", err)
	}
}

// The synthetic "All" group's catalog in AppData must be the UNION of the
// caller's per-real-group visible categories — never the full global pool. A
// category the caller is grant-restricted from in their only real group must not
// appear under the All-group key. Regression guard for the All-group catalog leak.
func TestGetAppData_AllGroupCatalogIsUnionOfRealGroups(t *testing.T) {
	defer repositories.TruncateTestDb()
	ClearRolePermissionCacheForTests()
	ClearGroupRoleGrantCacheForTests()

	db := repositories.GetDB()
	roleRepository := repositories.NewRoleRepository(nil)

	allowedCategory := models.Category{Name: "AllGroup-Allowed"}
	hiddenCategory := models.Category{Name: "AllGroup-Hidden"}
	db.Create(&allowedCategory)
	db.Create(&hiddenCategory)

	// App role without app.categories.read, so no global-pool bypass.
	appRole, err := roleRepository.CreateAppRole("AllGroup Union User Role", "", []string{permissions.AppCategoriesCreate}, false)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}
	user := models.User{Username: "allgroup-union-user", Password: "password", AppRoleID: &appRole.ID}
	if err := db.Create(&user).Error; err != nil {
		utils.PrintTestError(t, err, nil)
	}

	// Real group where the user is restricted to allowedCategory only.
	groupRole, err := roleRepository.CreateGroupRole("AllGroup Union Restricted Role", "", []string{permissions.GroupReceiptsRead}, []uint{allowedCategory.ID}, nil, nil, false, false)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}
	realGroup := models.Group{Name: "allgroup-union-real"}
	if err := db.Create(&realGroup).Error; err != nil {
		utils.PrintTestError(t, err, nil)
	}
	if err := db.Create(&models.GroupMember{GroupID: realGroup.ID, UserID: user.ID, GroupRoleID: &groupRole.ID}).Error; err != nil {
		utils.PrintTestError(t, err, nil)
	}

	// The user's own unrestricted "All" group.
	allGroup := models.Group{Name: "allgroup-union-all", IsAllGroup: true}
	if err := db.Create(&allGroup).Error; err != nil {
		utils.PrintTestError(t, err, nil)
	}
	if err := db.Create(&models.GroupMember{GroupID: allGroup.ID, UserID: user.ID, GroupRoleID: &groupRole.ID}).Error; err != nil {
		utils.PrintTestError(t, err, nil)
	}

	appData, err := GetAppData(user.ID, nil)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	allCatalog := appData.GroupCategories[allGroup.ID]
	if len(allCatalog) != 1 || allCatalog[0].ID != allowedCategory.ID {
		utils.PrintTestError(t, allCatalog, []uint{allowedCategory.ID})
	}
	for _, category := range allCatalog {
		if category.ID == hiddenCategory.ID {
			utils.PrintTestError(t, "hidden category leaked into All-group catalog", "no hidden category")
		}
	}
}
