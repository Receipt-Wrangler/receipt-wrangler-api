package oidc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/permissions"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

// startLink drives GET /oidc/link/{name} as the given user.
func startLink(t *testing.T, providerName string, userId uint) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, "/api/oidc/link/"+providerName, nil)
	request = withUrlParam(request, "name", providerName)
	request = request.WithContext(context.WithValue(
		request.Context(),
		jwtmiddleware.ContextKey{},
		&validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: userId}},
	))

	recorder := httptest.NewRecorder()
	LinkStart(recorder, request)

	return recorder
}

func countAuthSessions(t *testing.T) int64 {
	t.Helper()

	var count int64
	if err := repositories.GetDB().Model(&models.OidcAuthSession{}).Count(&count).Error; err != nil {
		t.Fatalf("failed to count auth sessions: %v", err)
	}

	return count
}

// TestLinkStartRequiresAccountUpdatePermission is the gate itself.
//
// Connecting a provider ADDS a way to sign into an account, so it must be at
// least as privileged as disconnecting one -- and DeleteOidcConnection already
// requires app.account.update. Without this check an administrator could build a
// role that cannot remove a sign-in method but can still add one.
func TestLinkStartRequiresAccountUpdatePermission(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	// Holds the read half but not the update half, which is exactly the role an
	// administrator would build to make an account's bindings look-but-don't-touch.
	user := createTestUserWithPermissions(t, "restricted", permissions.AppAccountRead)

	recorder := startLink(t, provider.Name, user.ID)

	if recorder.Code != http.StatusFound {
		t.Fatalf("expected a redirect, got %d", recorder.Code)
	}

	if code := oidcErrorCode(t, recorder); code != errForbidden {
		t.Errorf("expected the %q error code, got %q", errForbidden, code)
	}

	// Nothing may be started: a denied caller must not even reach the identity
	// provider, and must not leave a redeemable session row behind.
	if count := countAuthSessions(t); count != 0 {
		t.Errorf("expected no auth session to be created for a denied caller, got %d", count)
	}
}

// TestLinkStartFailureLandsOnTheProfileNotTheLoginPage pins the redirect target.
// A link caller is already signed in, so bouncing them to a login form reads as
// "you have been signed out" rather than "that provider could not be connected".
func TestLinkStartFailureLandsOnTheProfileNotTheLoginPage(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	user := createTestUserWithPermissions(t, "restricted")

	recorder := startLink(t, provider.Name, user.ID)

	location := recorder.Header().Get("Location")
	if !strings.HasPrefix(location, desktopProfilePath) {
		t.Errorf("expected a redirect to %q, got %q", desktopProfilePath, location)
	}

	if strings.Contains(location, desktopLoginPath) {
		t.Errorf("a signed-in caller must not be sent to the login page, got %q", location)
	}
}

// TestLinkStartProceedsWithAccountUpdatePermission is the other half: the gate
// must not break the flow it guards.
func TestLinkStartProceedsWithAccountUpdatePermission(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{})

	user := createTestUserWithPermissions(t, "allowed", permissions.AppAccountUpdate)

	recorder := startLink(t, provider.Name, user.ID)

	if recorder.Code != http.StatusFound {
		t.Fatalf("expected a redirect, got %d (%s)", recorder.Code, recorder.Body.String())
	}

	location, err := url.Parse(recorder.Header().Get("Location"))
	if err != nil {
		t.Fatalf("failed to parse the redirect: %v", err)
	}

	if !strings.HasPrefix(location.String(), idp.issuer()) {
		t.Fatalf("expected a redirect to the identity provider at %q, got %q", idp.issuer(), location)
	}

	var session models.OidcAuthSession
	err = repositories.GetDB().
		Model(&models.OidcAuthSession{}).
		Where("state_hash = ?", hashSecret(location.Query().Get("state"))).
		First(&session).Error
	if err != nil {
		t.Fatalf("expected an auth session for the minted state: %v", err)
	}

	if !session.IsLink() {
		t.Fatal("expected a link session, got a plain login session")
	}

	if *session.LinkUserId != user.ID {
		t.Errorf("expected the session to carry user %d, got %d", user.ID, *session.LinkUserId)
	}
}

// startMobileLink drives GET /oidc/link/{name}?client=mobile as the given user,
// which is how the app calls it: an ordinary authenticated API request.
func startMobileLink(t *testing.T, providerName string, userId uint) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, "/api/oidc/link/"+providerName+"?client=mobile", nil)
	request = withUrlParam(request, "name", providerName)
	request = request.WithContext(context.WithValue(
		request.Context(),
		jwtmiddleware.ContextKey{},
		&validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: userId}},
	))

	recorder := httptest.NewRecorder()
	LinkStart(recorder, request)

	return recorder
}

// launchUrlFrom reads the launch URL out of a mobile link start response.
func launchUrlFrom(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()

	var view structs.OidcLinkStartView
	if err := json.Unmarshal(recorder.Body.Bytes(), &view); err != nil {
		t.Fatalf("failed to decode the link start response: %v", err)
	}

	return view.LaunchUrl
}

// launchMobileLink follows a launch URL the way the EXTERNAL BROWSER does, which
// is the whole point of the hop: this request is what creates the auth session,
// so this is what receives the binding cookie.
func launchMobileLink(t *testing.T, launchUrl string) *httptest.ResponseRecorder {
	t.Helper()

	parsed, err := url.Parse(launchUrl)
	if err != nil {
		t.Fatalf("failed to parse the launch URL: %v", err)
	}

	// .../api/oidc/link/{name}/launch
	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(segments) < 2 {
		t.Fatalf("unexpected launch path %q", parsed.Path)
	}
	name := segments[len(segments)-2]

	request := httptest.NewRequest(http.MethodGet, parsed.Path+"?"+parsed.RawQuery, nil)
	request = withUrlParam(request, "name", name)

	recorder := httptest.NewRecorder()
	LinkLaunch(recorder, request)

	return recorder
}

// startAndLaunchMobileLink runs both halves and returns the browser-facing one.
func startAndLaunchMobileLink(t *testing.T, providerName string, userId uint) *httptest.ResponseRecorder {
	t.Helper()

	start := startMobileLink(t, providerName, userId)
	if start.Code != http.StatusOK {
		t.Fatalf("expected the link to start, got %d (%s)", start.Code, start.Body.String())
	}

	return launchMobileLink(t, launchUrlFrom(t, start))
}

// TestMobileLinkStartReturnsALaunchUrlAndCreatesNoSession pins the shape that
// makes the mobile leg bindable at all.
//
// The app authenticates with a bearer token, and the external user agent RFC 8252
// requires cannot carry one -- so the app calls this itself. That means there is
// no browser here to hand a binding cookie to, which is exactly why this start
// must NOT create the session. It hands back a launch URL and lets the browser
// create it one hop later, where a cookie can be set.
func TestMobileLinkStartReturnsALaunchUrlAndCreatesNoSession(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	user := createTestUserWithPermissions(t, "mobilelinker", permissions.AppAccountUpdate)

	recorder := startMobileLink(t, provider.Name, user.ID)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", recorder.Code, recorder.Body.String())
	}

	launchUrl := launchUrlFrom(t, recorder)

	// It points back at US, not at the identity provider. Handing the app the
	// authorization URL directly is the bug this replaced: that URL carries the
	// state, providers leak it by Referer, and the callback could not tell the
	// app's browser from anyone else's.
	parsed, err := url.Parse(launchUrl)
	if err != nil {
		t.Fatalf("failed to parse the launch URL: %v", err)
	}

	if !strings.HasSuffix(parsed.Path, "/api/oidc/link/"+provider.Name+"/launch") {
		t.Errorf("expected a launch path for %q, got %q", provider.Name, parsed.Path)
	}

	if len(parsed.Query().Get("h")) == 0 {
		t.Error("the launch URL must carry a handle")
	}

	// Nothing is redeemable at the identity provider yet.
	if count := countAuthSessions(t); count != 0 {
		t.Errorf("a mobile link start must create no auth session, got %d", count)
	}

	// This is an API call, not a navigation. A Set-Cookie here would be a binding
	// handed to the app, which is not the agent that will return with the callback.
	if len(recorder.Result().Cookies()) > 0 {
		t.Error("a mobile link start must not set cookies")
	}
}

// TestMobileLinkLaunchBindsTheBrowserThatOpensIt is the fix for the finding: the
// session is created by the browser, so the browser gets bound.
func TestMobileLinkLaunchBindsTheBrowserThatOpensIt(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{})

	user := createTestUserWithPermissions(t, "mobilelaunch", permissions.AppAccountUpdate)

	recorder := startAndLaunchMobileLink(t, provider.Name, user.ID)

	if recorder.Code != http.StatusFound {
		t.Fatalf("expected a 302 to the provider, got %d (%s)", recorder.Code, recorder.Body.String())
	}

	location := recorder.Header().Get("Location")
	if !strings.HasPrefix(location, idp.issuer()) {
		t.Fatalf("expected a redirect to %q, got %q", idp.issuer(), location)
	}

	if findCookie(recorder.Result().Cookies(), bindingCookieName) == nil {
		t.Fatal("the launch must set the binding cookie -- without it the callback is unbound")
	}

	parsed, err := url.Parse(location)
	if err != nil {
		t.Fatalf("failed to parse the authorization URL: %v", err)
	}

	var session models.OidcAuthSession
	err = repositories.GetDB().
		Model(&models.OidcAuthSession{}).
		Where("state_hash = ?", hashSecret(parsed.Query().Get("state"))).
		First(&session).Error
	if err != nil {
		t.Fatalf("expected an auth session for the minted state: %v", err)
	}

	if !session.IsLink() || *session.LinkUserId != user.ID {
		t.Errorf("expected a link session for user %d, got %+v", user.ID, session.LinkUserId)
	}

	if session.ClientType != models.OidcClientMobile {
		t.Errorf("expected a mobile session, got %q", session.ClientType)
	}

	if len(session.BindingHash) == 0 {
		t.Error("a mobile link session must carry a binding hash")
	}
}

// TestMobileLinkCallbackRefusesAnUnboundBrowser is the regression test for the
// finding itself.
//
// An attacker who obtains the authorization URL -- identity providers leak it by
// Referer from their own login pages -- can authenticate as THEMSELVES and drive
// this callback. Their browser holds no binding cookie, so it must be refused. If
// it is not, their identity is grafted onto the victim's account and they can
// sign in as the victim from then on.
func TestMobileLinkCallbackRefusesAnUnboundBrowser(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{})

	user := createTestUserWithPermissions(t, "mobileunbound", permissions.AppAccountUpdate)

	launch := startAndLaunchMobileLink(t, provider.Name, user.ID)
	authUrl, err := url.Parse(launch.Header().Get("Location"))
	if err != nil {
		t.Fatalf("failed to parse the authorization URL: %v", err)
	}

	idp.setClaims(claimsFor(idp, "attacker-subject", authUrl.Query().Get("nonce"), "attacker"))

	// The attacker's browser: correct state, no binding cookie.
	callback := runCallback(t, provider.Name, url.Values{
		"code":  {"abc"},
		"state": {authUrl.Query().Get("state")},
	}, nil)

	if code := oidcErrorCode(t, callback); code != errInvalidState {
		t.Errorf("expected %q for an unbound callback, got %q", errInvalidState, code)
	}

	_, err = repositories.NewOidcIdentityRepository(nil).GetIdentityBySubject(provider.ID, "attacker-subject")
	if err == nil {
		t.Fatal("an unbound callback must not link an identity")
	}
}

// TestMobileLinkLaunchHandleIsSingleUse: the handle is a bearer capability to
// start a link for its user, so a replay must not mint a second session.
func TestMobileLinkLaunchHandleIsSingleUse(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	user := createTestUserWithPermissions(t, "mobilereplay", permissions.AppAccountUpdate)

	start := startMobileLink(t, provider.Name, user.ID)
	if start.Code != http.StatusOK {
		t.Fatalf("expected the link to start, got %d (%s)", start.Code, start.Body.String())
	}

	launchUrl := launchUrlFrom(t, start)

	if first := launchMobileLink(t, launchUrl); first.Code != http.StatusFound {
		t.Fatalf("expected the first launch to redirect, got %d", first.Code)
	}

	second := launchMobileLink(t, launchUrl)
	if code := oidcErrorCode(t, second); code != errInvalidState {
		t.Errorf("expected %q on replay, got %q", errInvalidState, code)
	}

	if count := countAuthSessions(t); count != 1 {
		t.Errorf("a replayed handle must not create a second session, got %d", count)
	}
}

// TestMobileLinkStartNeedsNoCodeChallenge: the challenge binds the mobile LOGIN's
// exchange code to the app that started it. A link's callback hands back only
// ?linked={name}, which is not a credential, so there is nothing for the app to
// protect -- the session is bound by cookie like every other leg instead.
func TestMobileLinkStartNeedsNoCodeChallenge(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	user := createTestUserWithPermissions(t, "nochallenge", permissions.AppAccountUpdate)

	// startMobileLink deliberately sends no codeChallenge.
	if recorder := startMobileLink(t, provider.Name, user.ID); recorder.Code != http.StatusOK {
		t.Fatalf("expected 200 without a code challenge, got %d (%s)", recorder.Code, recorder.Body.String())
	}
}

// TestMobileLoginStillRequiresACodeChallenge guards the other half of that split.
// Relaxing it for the link must not relax it for the login, where the challenge
// is the only thing making an intercepted handoff code worthless.
func TestMobileLoginStillRequiresACodeChallenge(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	recorder := startLogin(t, provider.Name, "client=mobile")

	if code := oidcErrorCode(t, recorder); code != errInvalidRequest {
		t.Errorf("expected %q, got %q", errInvalidRequest, code)
	}

	if count := countAuthSessions(t); count != 0 {
		t.Errorf("no session may be created without a challenge, got %d", count)
	}
}

// TestMobileLinkStartDeniedAnswersWithJson: a denial has to reach the app in the
// same shape as a success, or it surfaces as an unparseable body rather than a
// message.
func TestMobileLinkStartDeniedAnswersWithJson(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	user := createTestUserWithPermissions(t, "mobilerestricted", permissions.AppAccountRead)

	recorder := startMobileLink(t, provider.Name, user.ID)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}

	if location := recorder.Header().Get("Location"); len(location) > 0 {
		t.Errorf("an API call must not be answered with a redirect, got %q", location)
	}

	var body structs.OidcFlowError
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode the error: %v", err)
	}

	if body.ErrorCode != errForbidden {
		t.Errorf("expected %q, got %q", errForbidden, body.ErrorCode)
	}

	if count := countAuthSessions(t); count != 0 {
		t.Errorf("expected no auth session for a denied caller, got %d", count)
	}
}

// TestMobileLinkStartUnknownProviderAnswersWithJson pins the reordering that made
// the client type known before the provider lookup. Without it this failure took
// the desktop redirect branch and the app received HTML it could not read.
func TestMobileLinkStartUnknownProviderAnswersWithJson(t *testing.T) {
	defer teardownOidcTest()
	setupOidcTest(t, oidcTestOptions{})

	user := createTestUserWithPermissions(t, "unknownprovider", permissions.AppAccountUpdate)

	recorder := startMobileLink(t, "no-such-provider", user.ID)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d (%s)", recorder.Code, recorder.Body.String())
	}

	var body structs.OidcFlowError
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode the error: %v", err)
	}

	if body.ErrorCode != errUnknownProvider {
		t.Errorf("expected %q, got %q", errUnknownProvider, body.ErrorCode)
	}
}

// TestMobileLinkCallbackReturnsToTheAppScheme is the far end of the flow: the
// callback lands in the external browser, and the only way back into the app is
// its private-use scheme. Redirecting to the desktop profile path -- as this did
// before -- strands the browser on a page the app never sees.
func TestMobileLinkCallbackReturnsToTheAppScheme(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{})

	user := createTestUserWithPermissions(t, "mobilecallback", permissions.AppAccountUpdate)

	launch := startAndLaunchMobileLink(t, provider.Name, user.ID)

	authUrl, err := url.Parse(launch.Header().Get("Location"))
	if err != nil {
		t.Fatalf("failed to parse the authorization URL: %v", err)
	}

	idp.setClaims(claimsFor(idp, "mobile-link-subject", authUrl.Query().Get("nonce"), "whoever"))

	// The binding cookie the launch set, carried back the way the browser does.
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

	if parsed.Query().Get("linked") != provider.Name {
		t.Errorf("expected linked=%q, got %q (error %q)", provider.Name, parsed.Query().Get("linked"), parsed.Query().Get("error"))
	}

	// A link mints no session, so nothing credential-like may ride back. The
	// callback does legitimately emit one Set-Cookie -- the expiry that retires the
	// binding cookie now that it has served its purpose -- so this asserts against
	// the token cookies specifically rather than against any cookie at all.
	for _, name := range []string{"jwt", "refreshToken"} {
		if findCookie(callback.Result().Cookies(), name) != nil {
			t.Errorf("a link must not set the %s cookie", name)
		}
	}

	for _, key := range []string{"code", "jwt", "token", "access_token", "refresh_token", "refreshToken"} {
		if len(parsed.Query().Get(key)) > 0 {
			t.Errorf("the app redirect must not carry %s", key)
		}
	}

	identity, err := repositories.NewOidcIdentityRepository(nil).GetIdentityBySubject(provider.ID, "mobile-link-subject")
	if err != nil {
		t.Fatalf("expected the identity to be linked: %v", err)
	}

	if identity.UserId != user.ID {
		t.Errorf("expected the link to land on user %d, got %d", user.ID, identity.UserId)
	}

	if identity.ProvisionedUser {
		t.Error("a link must not mark the account as provisioned -- it has its own password")
	}
}
