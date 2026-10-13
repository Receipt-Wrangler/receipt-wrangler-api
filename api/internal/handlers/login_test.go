package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/permissions"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strings"
	"testing"
)

func teardownLoginTests() {
	repositories.TruncateTestDb()
}

// loginRequest builds a request with the given login command in the context
// and an optional query string appended to the URL. When userAgent is
// non-empty it's set on the request (used to simulate mobile clients).
func loginRequest(cmd commands.LoginCommand, query, userAgent string) *http.Request {
	url := "/api/login"
	if query != "" {
		url += "?" + query
	}
	r := httptest.NewRequest(http.MethodPost, url, nil)
	r = r.WithContext(context.WithValue(r.Context(), "user", cmd))
	if userAgent != "" {
		r.Header.Set("User-Agent", userAgent)
	}
	return r
}

func createAdminUser(t *testing.T, username, password string) models.User {
	t.Helper()
	userRepo := repositories.NewUserRepository(nil)
	user, err := userRepo.CreateUser(commands.SignUpCommand{
		Username:    username,
		Password:    password,
		DisplayName: username,
	})
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}
	// "Administrator" is now defined by the modern app.users.read permission
	// (see IsFirstAdminToLogin): the first-admin login branch that seeds the
	// default prompt keys off it. The handlers test DB does not seed system
	// roles, so grant the admin permission explicitly.
	grantAppPerms(t, user.ID, permissions.AppUsersRead)
	return user
}

func TestLogin_UserNotFound(t *testing.T) {
	defer teardownLoginTests()

	w := httptest.NewRecorder()
	r := loginRequest(commands.LoginCommand{Username: "nobody", Password: "pw"}, "", "")

	Login(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	defer teardownLoginTests()

	createAdminUser(t, "wrongpw-user", "correctpassword")

	w := httptest.NewRecorder()
	r := loginRequest(commands.LoginCommand{Username: "wrongpw-user", Password: "not-the-password"}, "", "")

	Login(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestLogin_DummyUserRejected(t *testing.T) {
	defer teardownLoginTests()

	userRepo := repositories.NewUserRepository(nil)
	_, err := userRepo.CreateUser(commands.SignUpCommand{
		Username:    "dummy-user",
		Password:    "Password",
		DisplayName: "Dummy",
		IsDummyUser: true,
	})
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	w := httptest.NewRecorder()
	r := loginRequest(commands.LoginCommand{Username: "dummy-user", Password: "Password"}, "", "")

	Login(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestLogin_SuccessFirstAdminSetsCookies(t *testing.T) {
	defer teardownLoginTests()

	createAdminUser(t, "first-admin", "Password")

	w := httptest.NewRecorder()
	r := loginRequest(commands.LoginCommand{Username: "first-admin", Password: "Password"}, "", "")

	Login(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}

	jwtCookieSet, refreshCookieSet := false, false
	for _, c := range w.Result().Cookies() {
		if c.Name == constants.JwtKey && c.Value != "" {
			jwtCookieSet = true
		}
		if c.Name == constants.RefreshTokenKey && c.Value != "" {
			refreshCookieSet = true
		}
	}
	if !jwtCookieSet {
		utils.PrintTestError(t, "jwt cookie not set", "jwt cookie present")
	}
	if !refreshCookieSet {
		utils.PrintTestError(t, "refresh_token cookie not set", "refresh_token cookie present")
	}

	// Body should parse as AppData. Since we're not in mobile/tokensInBody
	// mode, Jwt/RefreshToken should be empty in the body.
	var appData structs.AppData
	if err := json.NewDecoder(w.Result().Body).Decode(&appData); err != nil {
		utils.PrintTestError(t, err, nil)
	}
	if appData.Jwt != "" {
		utils.PrintTestError(t, appData.Jwt, "")
	}
	if appData.RefreshToken != "" {
		utils.PrintTestError(t, appData.RefreshToken, "")
	}

	// Verify the first-admin path also kicked off default-prompt creation.
	var promptCount int64
	repositories.GetDB().Model(&models.Prompt{}).Count(&promptCount)
	if promptCount == 0 {
		utils.PrintTestError(t, "no default prompt created", "at least 1 prompt")
	}
}

// The wire proof for the secureCookies setting: the raw Set-Cookie headers a
// login writes carry the Secure attribute exactly when the setting is on.
func TestLogin_SetCookieHeadersFollowTheSecureCookiesSetting(t *testing.T) {
	defer teardownLoginTests()

	createAdminUser(t, "secure-cookie-admin", "Password")

	loginSetCookies := func() []string {
		w := httptest.NewRecorder()
		Login(w, loginRequest(commands.LoginCommand{Username: "secure-cookie-admin", Password: "Password"}, "", ""))
		if w.Result().StatusCode != http.StatusOK {
			utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		}
		return w.Result().Header.Values("Set-Cookie")
	}

	countSecure := func(headers []string) int {
		count := 0
		for _, header := range headers {
			if strings.Contains(header, "; Secure") {
				count++
			}
		}
		return count
	}

	off := loginSetCookies()
	if len(off) != 2 || countSecure(off) != 0 {
		utils.PrintTestError(t, off, "two Set-Cookie headers, neither Secure")
	}

	if err := repositories.SetSecureCookiesForTests(true); err != nil {
		utils.PrintTestError(t, err, nil)
	}

	on := loginSetCookies()
	if len(on) != 2 || countSecure(on) != 2 {
		utils.PrintTestError(t, on, "two Set-Cookie headers, both Secure")
	}
}

func TestLogin_SuccessSubsequentAdmin(t *testing.T) {
	defer teardownLoginTests()

	createAdminUser(t, "admin-one", "Password")

	// First login marks admin-one's LastLoginDate so the second login will
	// take the firstAdminToLogin=false branch.
	w := httptest.NewRecorder()
	r := loginRequest(commands.LoginCommand{Username: "admin-one", Password: "Password"}, "", "")
	Login(w, r)
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}

	// Second login — same admin, should return 200 without re-running the
	// "first admin" branch.
	w2 := httptest.NewRecorder()
	r2 := loginRequest(commands.LoginCommand{Username: "admin-one", Password: "Password"}, "", "")
	Login(w2, r2)
	if w2.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w2.Result().StatusCode, http.StatusOK)
	}

	// Only one default prompt should exist — the branch shouldn't have run twice.
	var promptCount int64
	repositories.GetDB().Model(&models.Prompt{}).Count(&promptCount)
	if promptCount != 1 {
		utils.PrintTestError(t, promptCount, int64(1))
	}
}

func TestLogin_TokensInBodyQuery(t *testing.T) {
	defer teardownLoginTests()

	createAdminUser(t, "body-tokens-user", "Password")

	w := httptest.NewRecorder()
	r := loginRequest(commands.LoginCommand{Username: "body-tokens-user", Password: "Password"}, "tokensInBody=true", "")

	Login(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}

	// No cookies should be set when tokensInBody=true.
	for _, c := range w.Result().Cookies() {
		if c.Name == constants.JwtKey || c.Name == constants.RefreshTokenKey {
			utils.PrintTestError(t, "cookie unexpectedly set: "+c.Name, "no auth cookies")
		}
	}

	var appData structs.AppData
	if err := json.NewDecoder(w.Result().Body).Decode(&appData); err != nil {
		utils.PrintTestError(t, err, nil)
	}
	if appData.Jwt == "" {
		utils.PrintTestError(t, "empty jwt in body", "jwt set")
	}
	if appData.RefreshToken == "" {
		utils.PrintTestError(t, "empty refresh token in body", "refresh token set")
	}
}

func TestLogin_MobileUserAgent(t *testing.T) {
	defer teardownLoginTests()

	createAdminUser(t, "mobile-user", "Password")

	w := httptest.NewRecorder()
	r := loginRequest(commands.LoginCommand{Username: "mobile-user", Password: "Password"}, "", "Dart/3.2 (dart:io)")

	Login(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}

	// Mobile path: body carries tokens AND cookies are still set (because
	// tokensInBody query flag is false).
	var appData structs.AppData
	if err := json.NewDecoder(w.Result().Body).Decode(&appData); err != nil {
		utils.PrintTestError(t, err, nil)
	}
	if appData.Jwt == "" {
		utils.PrintTestError(t, "empty jwt for mobile", "jwt set")
	}
	if appData.RefreshToken == "" {
		utils.PrintTestError(t, "empty refresh token for mobile", "refresh token set")
	}

	jwtCookieSet := false
	for _, c := range w.Result().Cookies() {
		if c.Name == constants.JwtKey {
			jwtCookieSet = true
		}
	}
	if !jwtCookieSet {
		utils.PrintTestError(t, "jwt cookie missing for mobile", "jwt cookie present")
	}
}

func TestLogin_NonAdminUserSkipsFirstAdminBranch(t *testing.T) {
	defer teardownLoginTests()

	// Seed an admin first so the next user defaults to USER.
	createAdminUser(t, "seed-admin", "Password")

	userRepo := repositories.NewUserRepository(nil)
	_, err := userRepo.CreateUser(commands.SignUpCommand{
		Username:    "regular-user",
		Password:    "Password",
		DisplayName: "Regular",
	})
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	// Baseline: no prompts exist yet.
	var beforeCount int64
	repositories.GetDB().Model(&models.Prompt{}).Count(&beforeCount)

	w := httptest.NewRecorder()
	r := loginRequest(commands.LoginCommand{Username: "regular-user", Password: "Password"}, "", "")
	Login(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}

	// Non-admin login must not create a default prompt.
	var afterCount int64
	repositories.GetDB().Model(&models.Prompt{}).Count(&afterCount)
	if afterCount != beforeCount {
		utils.PrintTestError(t, afterCount, beforeCount)
	}
}
