package oidc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/permissions"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

// startDeleteConfirmation drives GET /oidc/connections/{name}/delete-confirmation
// with claims built the way the named auth source would have built them.
func startDeleteConfirmation(t *testing.T, providerName string, userId uint, authSource string, query string) *httptest.ResponseRecorder {
	t.Helper()

	target := "/api/oidc/connections/" + providerName + "/delete-confirmation"
	if len(query) > 0 {
		target += "?" + query
	}

	request := httptest.NewRequest(http.MethodGet, target, nil)
	request = withUrlParam(request, "name", providerName)
	request = request.WithContext(context.WithValue(
		request.Context(),
		jwtmiddleware.ContextKey{},
		&validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: userId, AuthSource: authSource}},
	))

	recorder := httptest.NewRecorder()
	DeleteConfirmationStart(recorder, request)

	return recorder
}

// seedIdentity links userId to provider under subject. provisioned marks the
// account as having no password of its own.
func seedIdentity(t *testing.T, provider models.OidcProvider, userId uint, subject string, provisioned bool) {
	t.Helper()

	identity := models.OidcIdentity{
		OidcProviderId:  provider.ID,
		Subject:         subject,
		UserId:          userId,
		ProvisionedUser: provisioned,
	}
	if err := repositories.GetDB().Create(&identity).Error; err != nil {
		t.Fatalf("failed to seed the identity: %v", err)
	}
}

// seedPasswordlessUser creates an OIDC-only account allowed to delete itself.
func seedPasswordlessUser(t *testing.T, provider models.OidcProvider, username string, subject string) models.User {
	t.Helper()

	user := createTestUserWithPermissions(t, username, permissions.AppAccountDelete)
	seedIdentity(t, provider, user.ID, subject, true)

	return user
}

// deleteErrorCode pulls a delete confirmation's error code out of a redirect:
// deleteError on the desktop profile page, error on the app's scheme.
func deleteErrorCode(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()

	parsed, err := url.Parse(recorder.Header().Get("Location"))
	if err != nil {
		t.Fatalf("failed to parse redirect: %v", err)
	}

	if code := parsed.Query().Get("deleteError"); len(code) > 0 {
		return code
	}

	return parsed.Query().Get("error")
}

func countDeleteConfirmations(t *testing.T) int64 {
	t.Helper()

	var count int64
	if err := repositories.GetDB().Model(&models.OidcDeleteConfirmation{}).Count(&count).Error; err != nil {
		t.Fatalf("failed to count delete confirmations: %v", err)
	}

	return count
}

// runDeleteConfirmation starts a desktop delete confirmation and completes the
// callback with the given ID token claims, returning the callback response.
func runDeleteConfirmation(t *testing.T, idp *fakeIdp, provider models.OidcProvider, user models.User, claims func(nonce string) map[string]interface{}) *httptest.ResponseRecorder {
	t.Helper()

	start := startDeleteConfirmation(t, provider.Name, user.ID, structs.AuthSourceJwt, "")
	if start.Code != http.StatusFound {
		t.Fatalf("expected the confirmation to start, got %d (%s)", start.Code, start.Body.String())
	}

	authUrl, err := url.Parse(start.Header().Get("Location"))
	if err != nil {
		t.Fatalf("failed to parse the authorization URL: %v", err)
	}

	idp.setClaims(claims(authUrl.Query().Get("nonce")))

	return runCallback(t, provider.Name, url.Values{
		"code":  {"abc"},
		"state": {authUrl.Query().Get("state")},
	}, start.Result().Cookies())
}

// TestDeleteConfirmationStartRequiresAnInteractiveSession: a key must not be able
// to begin destroying the account that issued it, whatever its scope.
func TestDeleteConfirmationStartRequiresAnInteractiveSession(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})
	user := seedPasswordlessUser(t, provider, "keyholder", "keyholder-subject")

	for _, authSource := range []string{structs.AuthSourceApiKey, ""} {
		recorder := startDeleteConfirmation(t, provider.Name, user.ID, authSource, "client=mobile")
		if recorder.Code != http.StatusForbidden {
			t.Errorf("auth source %q: expected 403, got %d (%s)", authSource, recorder.Code, recorder.Body.String())
		}
	}

	if countAuthSessions(t) != 0 {
		t.Error("a refused start created an auth session")
	}
}

func TestDeleteConfirmationStartRequiresAccountDeletePermission(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	user := createTestUser(t, "nodelete")
	seedIdentity(t, provider, user.ID, "nodelete-subject", true)

	recorder := startDeleteConfirmation(t, provider.Name, user.ID, structs.AuthSourceJwt, "")
	if code := deleteErrorCode(t, recorder); code != errForbidden {
		t.Errorf("expected %q, got %q (%s)", errForbidden, code, recorder.Header().Get("Location"))
	}

	if countAuthSessions(t) != 0 {
		t.Error("a refused start created an auth session")
	}
}

// An account with a password of its own confirms with that password, never with
// a provider.
func TestDeleteConfirmationStartRefusesAnAccountWithAPassword(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	user := createTestUserWithPermissions(t, "haspassword", permissions.AppAccountDelete)
	seedIdentity(t, provider, user.ID, "haspassword-subject", false)

	recorder := startDeleteConfirmation(t, provider.Name, user.ID, structs.AuthSourceJwt, "")
	if code := deleteErrorCode(t, recorder); code != errForbidden {
		t.Errorf("expected %q, got %q (%s)", errForbidden, code, recorder.Header().Get("Location"))
	}

	if countAuthSessions(t) != 0 {
		t.Error("a refused start created an auth session")
	}
}

func TestDeleteConfirmationStartRefusesAProviderTheAccountIsNotLinkedTo(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{})

	other := createTestProvider(t, idp.issuer(), providerOptions{name: "otheridp", clientId: idp.clientId})
	user := seedPasswordlessUser(t, other, "elsewhere", "elsewhere-subject")

	recorder := startDeleteConfirmation(t, provider.Name, user.ID, structs.AuthSourceJwt, "")
	if code := deleteErrorCode(t, recorder); code != errNotLinked {
		t.Errorf("expected %q, got %q (%s)", errNotLinked, code, recorder.Header().Get("Location"))
	}

	// The mobile start is an API call, so it answers in JSON.
	mobile := startDeleteConfirmation(t, provider.Name, user.ID, structs.AuthSourceJwt, "client=mobile")
	var flowError structs.OidcFlowError
	if err := json.Unmarshal(mobile.Body.Bytes(), &flowError); err != nil || flowError.ErrorCode != errNotLinked {
		t.Errorf("expected a JSON %q, got %d (%s)", errNotLinked, mobile.Code, mobile.Body.String())
	}
}

// The delete confirmation asks the provider for a fresh sign-in with max_age=0,
// the standard parameter whose result can be checked through auth_time. It does
// not send prompt=login, which Google does not list as a supported value.
func TestDeleteConfirmationStartAsksForAFreshLogin(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{})
	user := seedPasswordlessUser(t, provider, "fresh", "fresh-subject")

	recorder := startDeleteConfirmation(t, provider.Name, user.ID, structs.AuthSourceJwt, "")
	if recorder.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d (%s)", recorder.Code, recorder.Body.String())
	}

	authUrl, err := url.Parse(recorder.Header().Get("Location"))
	if err != nil {
		t.Fatalf("failed to parse the authorization URL: %v", err)
	}

	if !strings.HasPrefix(authUrl.String(), idp.issuer()+"/auth") {
		t.Fatalf("expected a redirect to the identity provider, got %s", authUrl.String())
	}

	if got := authUrl.Query().Get("max_age"); got != "0" {
		t.Errorf("expected max_age=0, got %q", got)
	}

	if authUrl.Query().Has("prompt") {
		t.Errorf("expected no prompt parameter, got %q", authUrl.Query().Get("prompt"))
	}

	var session models.OidcAuthSession
	if err := repositories.GetDB().Where("state_hash = ?", hashSecret(authUrl.Query().Get("state"))).First(&session).Error; err != nil {
		t.Fatalf("failed to load the session: %v", err)
	}

	if !session.IsDeleteConfirmation() || session.IsLink() {
		t.Errorf("expected a delete confirmation session, got purpose %q", session.Purpose)
	}

	if session.LinkUserId == nil || *session.LinkUserId != user.ID {
		t.Errorf("expected the session to carry user %d", user.ID)
	}
}

// Only the delete confirmation forces a fresh sign-in; a login or a link must not
// start sending max_age.
func TestLoginAndLinkDoNotAskForAFreshLogin(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})
	user := createTestUserWithPermissions(t, "linker", permissions.AppAccountUpdate)

	for name, recorder := range map[string]*httptest.ResponseRecorder{
		"login": startLogin(t, provider.Name, ""),
		"link":  startLink(t, provider.Name, user.ID),
	} {
		authUrl, err := url.Parse(recorder.Header().Get("Location"))
		if err != nil {
			t.Fatalf("%s: failed to parse the authorization URL: %v", name, err)
		}

		if authUrl.Query().Has("max_age") {
			t.Errorf("%s: expected no max_age, got %q", name, authUrl.Query().Get("max_age"))
		}
	}
}

func TestDeleteConfirmationDesktopCallbackIssuesACode(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{})
	user := seedPasswordlessUser(t, provider, "desktopdelete", "desktopdelete-subject")

	callback := runDeleteConfirmation(t, idp, provider, user, func(nonce string) map[string]interface{} {
		return claimsFor(idp, "desktopdelete-subject", nonce, "desktopdelete")
	})

	location := callback.Header().Get("Location")
	if !strings.HasPrefix(location, desktopProfilePath) {
		t.Fatalf("expected a return to the profile page, got %q", location)
	}

	parsed, err := url.Parse(location)
	if err != nil {
		t.Fatalf("failed to parse the redirect: %v", err)
	}

	code := parsed.Query().Get("deleteConfirmation")
	if len(code) == 0 {
		t.Fatalf("expected a deleteConfirmation code, got %q (error %q)", location, parsed.Query().Get("deleteError"))
	}

	var stored models.OidcDeleteConfirmation
	if err := repositories.GetDB().Where("code_hash = ?", hashSecret(code)).First(&stored).Error; err != nil {
		t.Fatalf("expected the code to be stored hashed: %v", err)
	}

	if stored.UserId != user.ID || stored.Used {
		t.Errorf("expected an unused code for user %d, got user %d used=%v", user.ID, stored.UserId, stored.Used)
	}

	// A confirmation is not a sign-in: nothing credential-like may ride back.
	for _, name := range []string{constants.JwtKey, constants.RefreshTokenKey} {
		if findCookie(callback.Result().Cookies(), name) != nil {
			t.Errorf("a delete confirmation must not set the %s cookie", name)
		}
	}
}

func TestDeleteConfirmationMobileFlowIssuesACodeToTheApp(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{})
	user := seedPasswordlessUser(t, provider, "mobiledelete", "mobiledelete-subject")

	start := startDeleteConfirmation(t, provider.Name, user.ID, structs.AuthSourceJwt, "client=mobile")
	if start.Code != http.StatusOK {
		t.Fatalf("expected a launch URL, got %d (%s)", start.Code, start.Body.String())
	}

	if countAuthSessions(t) != 0 {
		t.Error("the mobile start must create no session; the browser's launch does")
	}

	var launchRow models.OidcLinkLaunch
	if err := repositories.GetDB().First(&launchRow).Error; err != nil {
		t.Fatalf("expected a launch row: %v", err)
	}

	if launchRow.Purpose != models.OidcPurposeDeleteAccount {
		t.Errorf("expected the launch to carry the delete purpose, got %q", launchRow.Purpose)
	}

	launch := launchMobileLink(t, launchUrlFrom(t, start))

	authUrl, err := url.Parse(launch.Header().Get("Location"))
	if err != nil {
		t.Fatalf("failed to parse the authorization URL: %v", err)
	}

	if authUrl.Query().Get("max_age") != "0" {
		t.Errorf("expected the launched flow to ask for a fresh login, got %q", authUrl.String())
	}

	idp.setClaims(claimsFor(idp, "mobiledelete-subject", authUrl.Query().Get("nonce"), "mobiledelete"))

	callback := runCallback(t, provider.Name, url.Values{
		"code":  {"abc"},
		"state": {authUrl.Query().Get("state")},
	}, launch.Result().Cookies())

	location := callback.Header().Get("Location")
	if !strings.HasPrefix(location, mobileCallbackScheme) {
		t.Fatalf("expected a return to %q, got %q", mobileCallbackScheme, location)
	}

	parsed, err := url.Parse(location)
	if err != nil {
		t.Fatalf("failed to parse the app redirect: %v", err)
	}

	if len(parsed.Query().Get("deleteConfirmation")) == 0 {
		t.Errorf("expected a deleteConfirmation code, got %q", location)
	}

	if countDeleteConfirmations(t) != 1 {
		t.Errorf("expected exactly one confirmation row")
	}
}

// Signing in at the provider as anyone but the account's own identity confirms
// nothing: not another user's identity, and not an identity no one has linked.
func TestDeleteConfirmationRefusesADifferentIdentity(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{})
	user := seedPasswordlessUser(t, provider, "victim", "victim-subject")

	someoneElse := createTestUser(t, "someoneelse")
	seedIdentity(t, provider, someoneElse.ID, "someoneelse-subject", true)

	for _, subject := range []string{"someoneelse-subject", "unlinked-subject"} {
		callback := runDeleteConfirmation(t, idp, provider, user, func(nonce string) map[string]interface{} {
			return claimsFor(idp, subject, nonce, "whoever")
		})

		if code := deleteErrorCode(t, callback); code != errIdentityMismatch {
			t.Errorf("subject %q: expected %q, got %q (%s)", subject, errIdentityMismatch, code, callback.Header().Get("Location"))
		}

		if location := callback.Header().Get("Location"); !strings.HasPrefix(location, desktopProfilePath) {
			t.Errorf("subject %q: expected the failure on the profile page, got %q", subject, location)
		}
	}

	if countDeleteConfirmations(t) != 0 {
		t.Error("a refused confirmation issued a code")
	}
}

// auth_time, when the provider sends it, must show a sign-in made after the flow
// began. When it is absent the round trip is accepted: refusing would leave users
// of a provider that never sends it unable to delete their account.
func TestDeleteConfirmationChecksAuthTimeWhenPresent(t *testing.T) {
	now := time.Now()

	for _, testCase := range []struct {
		name     string
		authTime interface{}
		wantCode bool
	}{
		{name: "absent", authTime: nil, wantCode: true},
		{name: "fresh", authTime: now.Unix(), wantCode: true},
		{name: "fresh, as a float", authTime: float64(now.Unix()) + 0.5, wantCode: true},
		{name: "fresh, as a string", authTime: strconv.FormatInt(now.Unix(), 10), wantCode: true},
		{name: "unreadable reads as absent", authTime: "not-a-time", wantCode: true},
		{name: "stale", authTime: now.Add(-10 * time.Minute).Unix(), wantCode: false},
		{name: "stale, as a string", authTime: strconv.FormatInt(now.Add(-10*time.Minute).Unix(), 10), wantCode: false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			defer teardownOidcTest()
			idp, provider := setupOidcTest(t, oidcTestOptions{})
			user := seedPasswordlessUser(t, provider, "authtime", "authtime-subject")

			callback := runDeleteConfirmation(t, idp, provider, user, func(nonce string) map[string]interface{} {
				claims := claimsFor(idp, "authtime-subject", nonce, "authtime")
				if testCase.authTime != nil {
					claims["auth_time"] = testCase.authTime
				}
				return claims
			})

			parsed, err := url.Parse(callback.Header().Get("Location"))
			if err != nil {
				t.Fatalf("failed to parse the redirect: %v", err)
			}

			gotCode := len(parsed.Query().Get("deleteConfirmation")) > 0
			if gotCode != testCase.wantCode {
				t.Errorf("expected a code: %v, got %q", testCase.wantCode, parsed.String())
			}

			if !testCase.wantCode && deleteErrorCode(t, callback) != errStaleLogin {
				t.Errorf("expected %q, got %q", errStaleLogin, deleteErrorCode(t, callback))
			}
		})
	}
}

// A user who backs out at the provider is already signed in here, so the failure
// belongs on the profile page they started from -- for a delete confirmation and
// for a link alike. Nothing has been consumed at that point, so the pending
// session is the only thing that can say which flow it was.
func TestSignedInFlowsReturnToTheProfileWhenTheProviderRefuses(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	deleter := seedPasswordlessUser(t, provider, "backsout", "backsout-subject")
	linker := createTestUserWithPermissions(t, "linkbacksout", permissions.AppAccountUpdate)

	for _, testCase := range []struct {
		name      string
		start     *httptest.ResponseRecorder
		errorName string
	}{
		{name: "delete confirmation", start: startDeleteConfirmation(t, provider.Name, deleter.ID, structs.AuthSourceJwt, ""), errorName: "deleteError"},
		{name: "link", start: startLink(t, provider.Name, linker.ID), errorName: "oidcError"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			state, cookies := loginAndExtractFrom(t, testCase.start)

			callback := runCallback(t, provider.Name, url.Values{
				"state": {state},
				"error": {"access_denied"},
			}, cookies)

			location := callback.Header().Get("Location")
			if !strings.HasPrefix(location, desktopProfilePath) {
				t.Fatalf("expected a return to the profile page, got %q", location)
			}

			parsed, err := url.Parse(location)
			if err != nil {
				t.Fatalf("failed to parse the redirect: %v", err)
			}

			if code := parsed.Query().Get(testCase.errorName); code != errProviderError {
				t.Errorf("expected %s=%q, got %q", testCase.errorName, errProviderError, location)
			}
		})
	}
}
