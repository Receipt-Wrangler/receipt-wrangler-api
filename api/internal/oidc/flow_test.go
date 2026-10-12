package oidc

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

// startLogin drives GET /oidc/{name}/login and returns the response recorder.
func startLogin(t *testing.T, providerName string, query string) *httptest.ResponseRecorder {
	t.Helper()

	target := "/api/oidc/" + providerName + "/login"
	if len(query) > 0 {
		target += "?" + query
	}

	request := httptest.NewRequest(http.MethodGet, target, nil)
	request = withUrlParam(request, "name", providerName)

	recorder := httptest.NewRecorder()
	Login(recorder, request)

	return recorder
}

// runCallback drives GET /oidc/{name}/callback with the given query and cookies.
func runCallback(t *testing.T, providerName string, query url.Values, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, "/api/oidc/"+providerName+"/callback?"+query.Encode(), nil)
	request = withUrlParam(request, "name", providerName)

	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}

	recorder := httptest.NewRecorder()
	Callback(recorder, request)

	return recorder
}

func withUrlParam(r *http.Request, key string, value string) *http.Request {
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(key, value)

	return r.WithContext(withChiContext(r, routeContext))
}

// oidcErrorCode pulls the error code out of a redirect Location.
func oidcErrorCode(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()

	location := recorder.Header().Get("Location")
	parsed, err := url.Parse(location)
	if err != nil {
		t.Fatalf("failed to parse redirect %q: %v", location, err)
	}

	if code := parsed.Query().Get("oidcError"); len(code) > 0 {
		return code
	}

	return parsed.Query().Get("error")
}

// loginAndExtract starts a login and returns the state the API minted plus the
// binding cookie it set, which is everything a callback needs.
func loginAndExtract(t *testing.T, provider models.OidcProvider, query string) (string, []*http.Cookie) {
	t.Helper()

	recorder := startLogin(t, provider.Name, query)
	if recorder.Code != http.StatusFound {
		t.Fatalf("expected a redirect from login, got %d (%s)", recorder.Code, recorder.Body.String())
	}

	location, err := url.Parse(recorder.Header().Get("Location"))
	if err != nil {
		t.Fatalf("failed to parse authorization URL: %v", err)
	}

	state := location.Query().Get("state")
	if len(state) == 0 {
		t.Fatal("login did not put a state on the authorization URL")
	}

	return state, recorder.Result().Cookies()
}

// nonceFromAuthUrl reads the nonce the API sent to the identity provider, so a
// test can mint an ID token that legitimately matches it.
func nonceFromAuthUrl(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()

	location, err := url.Parse(recorder.Header().Get("Location"))
	if err != nil {
		t.Fatalf("failed to parse authorization URL: %v", err)
	}

	return location.Query().Get("nonce")
}

func TestLoginRedirectsWithStateNonceAndPkce(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{})

	recorder := startLogin(t, provider.Name, "")

	if recorder.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", recorder.Code)
	}

	location, err := url.Parse(recorder.Header().Get("Location"))
	if err != nil {
		t.Fatalf("failed to parse redirect: %v", err)
	}

	if !strings.HasPrefix(location.String(), idp.issuer()+"/auth") {
		t.Errorf("expected a redirect to the identity provider, got %s", location.String())
	}

	query := location.Query()

	for _, key := range []string{"state", "nonce", "code_challenge"} {
		if len(query.Get(key)) == 0 {
			t.Errorf("authorization URL is missing %s", key)
		}
	}

	if query.Get("code_challenge_method") != "S256" {
		t.Errorf("expected S256 PKCE, got %q", query.Get("code_challenge_method"))
	}

	if !strings.Contains(query.Get("scope"), "openid") {
		t.Errorf("expected the openid scope, got %q", query.Get("scope"))
	}

	if query.Get("redirect_uri") != "http://localhost:8081/api/oidc/"+provider.Name+"/callback" {
		t.Errorf("unexpected redirect_uri %q", query.Get("redirect_uri"))
	}
}

func TestLoginSetsHttpOnlyLaxBindingCookie(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	recorder := startLogin(t, provider.Name, "")

	cookie := findCookie(recorder.Result().Cookies(), bindingCookieName)
	if cookie == nil {
		t.Fatal("login did not set the browser-binding cookie")
	}

	if !cookie.HttpOnly {
		t.Error("the binding cookie must be HttpOnly")
	}

	// Lax, not Strict: the callback is a cross-site top-level GET from the identity
	// provider, and Strict would drop the cookie on exactly that request.
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("expected SameSite=Lax, got %v", cookie.SameSite)
	}

	if cookie.Path != bindingCookiePath {
		t.Errorf("expected the cookie scoped to %s, got %s", bindingCookiePath, cookie.Path)
	}
}

func TestLoginStoresSecretsHashedNotPlaintext(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	recorder := startLogin(t, provider.Name, "")
	state, _ := loginAndExtractFrom(t, recorder)
	nonce := nonceFromAuthUrl(t, recorder)

	var session models.OidcAuthSession
	if err := repositories.GetDB().Where("oidc_provider_id = ?", provider.ID).Last(&session).Error; err != nil {
		t.Fatalf("failed to load the persisted session: %v", err)
	}

	// A database dump must not be usable to complete somebody's pending login.
	if session.StateHash == state {
		t.Error("the state was stored in plaintext")
	}

	if session.NonceHash == nonce {
		t.Error("the nonce was stored in plaintext")
	}

	if session.StateHash != hashSecret(state) {
		t.Error("the stored state hash does not match the state that was issued")
	}

	// The verifier cannot be hashed -- it has to be replayed to the identity
	// provider verbatim -- so it must at least be encrypted.
	if len(session.CodeVerifier) == 0 {
		t.Fatal("no code verifier was persisted")
	}

	decrypted, err := decryptVerifier(session)
	if err != nil {
		t.Fatalf("failed to decrypt the stored verifier: %v", err)
	}

	if decrypted == session.CodeVerifier {
		t.Error("the code verifier was stored in plaintext")
	}
}

func TestLoginRejectsUnknownAndDisabledProviders(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	recorder := startLogin(t, "does-not-exist", "")
	if code := oidcErrorCode(t, recorder); code != errUnknownProvider {
		t.Errorf("expected %s for an unknown provider, got %q", errUnknownProvider, code)
	}

	// A disabled provider is deliberately indistinguishable from a missing one.
	if err := repositories.GetDB().Model(&models.OidcProvider{}).Where("id = ?", provider.ID).Update("enabled", false).Error; err != nil {
		t.Fatalf("failed to disable the provider: %v", err)
	}

	recorder = startLogin(t, provider.Name, "")
	if code := oidcErrorCode(t, recorder); code != errUnknownProvider {
		t.Errorf("expected %s for a disabled provider, got %q", errUnknownProvider, code)
	}
}

func TestMobileLoginRequiresCodeChallenge(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	// Without the app's own PKCE challenge the exchange code would be bearer-only,
	// and any app that registered the same URL scheme could redeem it.
	recorder := startLogin(t, provider.Name, "client=mobile")

	if code := oidcErrorCode(t, recorder); code != errInvalidRequest {
		t.Errorf("expected %s without a code challenge, got %q", errInvalidRequest, code)
	}

	if !strings.HasPrefix(recorder.Header().Get("Location"), mobileCallbackScheme) {
		t.Errorf("a mobile error must come back on the app's scheme, got %s", recorder.Header().Get("Location"))
	}
}

// TestMobileLoginSetsBindingCookie inverts an assertion that used to read "the
// mobile leg must not depend on a cookie the external user agent may not carry".
//
// That premise was wrong, and it cost the mobile legs their login-CSRF defense.
// The app hands the LOGIN URL itself to flutter_web_auth_2, so the external
// browser is what fetches this start -- it takes the Set-Cookie and returns it on
// the callback exactly as a desktop browser does. Skipping it meant anyone who
// obtained the authorization URL could complete the flow as themselves and have
// the victim's app sign into THEIR account.
func TestMobileLoginSetsBindingCookie(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	recorder := startLogin(t, provider.Name, "client=mobile&codeChallenge="+challengeFor("verifier-for-the-mobile-app-1234567890"))

	if findCookie(recorder.Result().Cookies(), bindingCookieName) == nil {
		t.Error("the mobile login leg must be bound to the browser that started it")
	}
}

// TestMobileLoginCallbackRefusesAnUnboundBrowser is the login-side twin of the
// link regression: a correct state without the binding cookie must not complete.
func TestMobileLoginCallbackRefusesAnUnboundBrowser(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{allowProvisioning: true})

	recorder := startLogin(t, provider.Name, "client=mobile&codeChallenge="+challengeFor("verifier-for-the-mobile-app-1234567890"))

	state, _ := loginAndExtractFrom(t, recorder)
	idp.setClaims(claimsFor(idp, "unbound-mobile-subject", nonceFromAuthUrl(t, recorder), "unbound"))

	callback := runCallback(t, provider.Name, url.Values{
		"code":  {"abc"},
		"state": {state},
	}, nil)

	if code := oidcErrorCode(t, callback); code != errInvalidState {
		t.Errorf("expected %q for an unbound callback, got %q", errInvalidState, code)
	}
}

func TestCallbackRejectsMissingCodeOrState(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	recorder := runCallback(t, provider.Name, url.Values{"code": {"abc"}}, nil)
	if code := oidcErrorCode(t, recorder); code != errInvalidRequest {
		t.Errorf("expected %s without a state, got %q", errInvalidRequest, code)
	}

	recorder = runCallback(t, provider.Name, url.Values{"state": {"abc"}}, nil)
	if code := oidcErrorCode(t, recorder); code != errInvalidRequest {
		t.Errorf("expected %s without a code, got %q", errInvalidRequest, code)
	}
}

func TestCallbackRejectsUnknownState(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	recorder := runCallback(t, provider.Name, url.Values{
		"code":  {"abc"},
		"state": {"a-state-that-was-never-issued"},
	}, nil)

	if code := oidcErrorCode(t, recorder); code != errInvalidState {
		t.Errorf("expected %s, got %q", errInvalidState, code)
	}
}

func TestCallbackRejectsReplayedState(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{allowProvisioning: true})

	loginRecorder := startLogin(t, provider.Name, "")
	state, cookies := loginAndExtractFrom(t, loginRecorder)
	idp.setClaims(claimsFor(idp, "subject-replay", nonceFromAuthUrl(t, loginRecorder), "replayuser"))

	query := url.Values{"code": {"abc"}, "state": {state}}

	first := runCallback(t, provider.Name, query, cookies)
	if first.Header().Get("Location") != desktopCallbackPath {
		t.Fatalf("the first callback should have succeeded, got %s (%s)", first.Header().Get("Location"), oidcErrorCode(t, first))
	}

	// The state is claimed atomically BEFORE the token exchange, so a replay never
	// even reaches the identity provider.
	second := runCallback(t, provider.Name, query, cookies)
	if code := oidcErrorCode(t, second); code != errInvalidState {
		t.Errorf("expected a replayed state to be refused with %s, got %q", errInvalidState, code)
	}
}

func TestCallbackRejectsExpiredState(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	loginRecorder := startLogin(t, provider.Name, "")
	state, cookies := loginAndExtractFrom(t, loginRecorder)

	if err := repositories.GetDB().
		Model(&models.OidcAuthSession{}).
		Where("state_hash = ?", hashSecret(state)).
		Update("expires_at", timeInPast()).Error; err != nil {
		t.Fatalf("failed to expire the session: %v", err)
	}

	recorder := runCallback(t, provider.Name, url.Values{"code": {"abc"}, "state": {state}}, cookies)
	if code := oidcErrorCode(t, recorder); code != errInvalidState {
		t.Errorf("expected %s for an expired state, got %q", errInvalidState, code)
	}
}

func TestCallbackRejectsMissingBindingCookie(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{allowProvisioning: true})

	loginRecorder := startLogin(t, provider.Name, "")
	state, _ := loginAndExtractFrom(t, loginRecorder)

	// This is the login-CSRF case: an attacker who plants their own code and state
	// in a victim's browser cannot also produce the victim's binding cookie.
	recorder := runCallback(t, provider.Name, url.Values{"code": {"abc"}, "state": {state}}, nil)

	if code := oidcErrorCode(t, recorder); code != errInvalidState {
		t.Errorf("expected %s without the binding cookie, got %q", errInvalidState, code)
	}
}

func TestCallbackRejectsWrongBindingCookie(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{allowProvisioning: true})

	loginRecorder := startLogin(t, provider.Name, "")
	state, _ := loginAndExtractFrom(t, loginRecorder)

	wrong := []*http.Cookie{{Name: bindingCookieName, Value: "a-different-browsers-binding"}}

	recorder := runCallback(t, provider.Name, url.Values{"code": {"abc"}, "state": {state}}, wrong)
	if code := oidcErrorCode(t, recorder); code != errInvalidState {
		t.Errorf("expected %s with a foreign binding cookie, got %q", errInvalidState, code)
	}
}

func TestCallbackRejectsStateFromAnotherProvider(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{allowProvisioning: true})

	other := createTestProvider(t, provider.IssuerUrl, providerOptions{
		name:              "other",
		clientId:          "other-client",
		allowProvisioning: true,
	})

	loginRecorder := startLogin(t, provider.Name, "")
	state, cookies := loginAndExtractFrom(t, loginRecorder)

	// A state minted for one provider must not be redeemable at another's callback.
	recorder := runCallback(t, other.Name, url.Values{"code": {"abc"}, "state": {state}}, cookies)
	if code := oidcErrorCode(t, recorder); code != errInvalidState {
		t.Errorf("expected %s for a cross-provider state, got %q", errInvalidState, code)
	}
}

func TestCallbackRejectsNonceMismatch(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{allowProvisioning: true})

	loginRecorder := startLogin(t, provider.Name, "")
	state, cookies := loginAndExtractFrom(t, loginRecorder)

	// go-oidc's Verify explicitly does NOT check the nonce; without our own check
	// this ID token would be accepted and replay would be back.
	idp.setClaims(claimsFor(idp, "subject-nonce", "a-nonce-we-never-issued", "nonceuser"))

	recorder := runCallback(t, provider.Name, url.Values{"code": {"abc"}, "state": {state}}, cookies)
	if code := oidcErrorCode(t, recorder); code != errNonceMismatch {
		t.Errorf("expected %s, got %q", errNonceMismatch, code)
	}
}

func TestCallbackRejectsEmptyNonce(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{allowProvisioning: true})

	loginRecorder := startLogin(t, provider.Name, "")
	state, cookies := loginAndExtractFrom(t, loginRecorder)

	// An empty nonce hashes to a fixed value, so it must be rejected before the
	// comparison rather than allowed to match anything.
	claims := claimsFor(idp, "subject-empty-nonce", "", "emptynonce")
	delete(claims, "nonce")
	idp.setClaims(claims)

	recorder := runCallback(t, provider.Name, url.Values{"code": {"abc"}, "state": {state}}, cookies)
	if code := oidcErrorCode(t, recorder); code != errNonceMismatch {
		t.Errorf("expected %s for an absent nonce, got %q", errNonceMismatch, code)
	}
}

func TestCallbackRejectsMissingIdToken(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{allowProvisioning: true})

	loginRecorder := startLogin(t, provider.Name, "")
	state, cookies := loginAndExtractFrom(t, loginRecorder)

	// An OAuth-only response carries no verified identity. Refuse rather than
	// falling back to the unsigned userinfo endpoint.
	idp.setOmitIdToken(true)

	recorder := runCallback(t, provider.Name, url.Values{"code": {"abc"}, "state": {state}}, cookies)
	if code := oidcErrorCode(t, recorder); code != errNoIdToken {
		t.Errorf("expected %s, got %q", errNoIdToken, code)
	}
}

func TestCallbackRejectsIdTokenSignedByAnUnknownKey(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{allowProvisioning: true})

	loginRecorder := startLogin(t, provider.Name, "")
	state, cookies := loginAndExtractFrom(t, loginRecorder)

	idp.setClaims(claimsFor(idp, "subject-forged", nonceFromAuthUrl(t, loginRecorder), "forged"))
	idp.setSignWithWrongKey(true)

	recorder := runCallback(t, provider.Name, url.Values{"code": {"abc"}, "state": {state}}, cookies)
	if code := oidcErrorCode(t, recorder); code != errProviderError {
		t.Errorf("expected %s for a forged signature, got %q", errProviderError, code)
	}
}

func TestCallbackRejectsIdTokenForAnotherAudience(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{allowProvisioning: true})

	loginRecorder := startLogin(t, provider.Name, "")
	state, cookies := loginAndExtractFrom(t, loginRecorder)

	claims := claimsFor(idp, "subject-aud", nonceFromAuthUrl(t, loginRecorder), "auduser")
	claims["aud"] = "some-other-clients-id"
	idp.setClaims(claims)

	recorder := runCallback(t, provider.Name, url.Values{"code": {"abc"}, "state": {state}}, cookies)
	if code := oidcErrorCode(t, recorder); code != errProviderError {
		t.Errorf("expected %s for a foreign audience, got %q", errProviderError, code)
	}
}

func TestCallbackHappyPathSetsSessionCookies(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{allowProvisioning: true})

	loginRecorder := startLogin(t, provider.Name, "")
	state, cookies := loginAndExtractFrom(t, loginRecorder)
	idp.setClaims(claimsFor(idp, "subject-happy", nonceFromAuthUrl(t, loginRecorder), "happyuser"))

	recorder := runCallback(t, provider.Name, url.Values{"code": {"abc"}, "state": {state}}, cookies)

	if recorder.Header().Get("Location") != desktopCallbackPath {
		t.Fatalf("expected a redirect to %s, got %s (error %q)", desktopCallbackPath, recorder.Header().Get("Location"), oidcErrorCode(t, recorder))
	}

	result := recorder.Result().Cookies()
	if findCookie(result, "jwt") == nil || findCookie(result, "refresh_token") == nil {
		t.Error("expected the session cookies to be set")
	}

	// And the binding cookie must be cleared once it has served its purpose.
	binding := findCookie(result, bindingCookieName)
	if binding == nil || binding.MaxAge >= 0 {
		t.Error("expected the binding cookie to be cleared")
	}
}

func TestCallbackMobileReturnsCodeNotTokens(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{allowProvisioning: true})

	verifier := "a-mobile-app-pkce-verifier-that-is-long-enough"
	loginRecorder := startLogin(t, provider.Name, "client=mobile&codeChallenge="+challengeFor(verifier))
	state, _ := loginAndExtractFrom(t, loginRecorder)
	idp.setClaims(claimsFor(idp, "subject-mobile", nonceFromAuthUrl(t, loginRecorder), "mobileuser"))

	recorder := runCallback(t, provider.Name, url.Values{"code": {"abc"}, "state": {state}}, loginRecorder.Result().Cookies())

	location := recorder.Header().Get("Location")
	if !strings.HasPrefix(location, mobileCallbackScheme) {
		t.Fatalf("expected the app's scheme, got %s", location)
	}

	parsed, err := url.Parse(location)
	if err != nil {
		t.Fatalf("failed to parse the app redirect: %v", err)
	}

	code := parsed.Query().Get("code")
	if len(code) == 0 {
		t.Fatal("no exchange code was delivered to the app")
	}

	// A private-use scheme is unverifiable on Android, so this URL must never carry
	// anything that is a credential on its own.
	//
	// Asserted against the TOKEN cookies rather than against any cookie at all:
	// the mobile leg is now bound like every other one, so the callback
	// legitimately emits the expiry that retires the binding cookie.
	for _, name := range []string{"jwt", "refreshToken"} {
		if findCookie(recorder.Result().Cookies(), name) != nil {
			t.Errorf("the mobile leg must not set the %s cookie", name)
		}
	}

	for _, key := range []string{"jwt", "token", "access_token", "refresh_token", "refreshToken"} {
		if len(parsed.Query().Get(key)) > 0 {
			t.Errorf("the app redirect must not carry %s", key)
		}
	}

	// And the code on the wire must not be what is stored, so a database read
	// cannot redeem it.
	var stored models.OidcExchangeCode
	if err := repositories.GetDB().Last(&stored).Error; err != nil {
		t.Fatalf("failed to load the exchange code: %v", err)
	}

	if stored.CodeHash == code {
		t.Error("the exchange code was stored in plaintext")
	}
}

// loginAndExtractFrom pulls the state and cookies out of an already-run login.
func loginAndExtractFrom(t *testing.T, recorder *httptest.ResponseRecorder) (string, []*http.Cookie) {
	t.Helper()

	if recorder.Code != http.StatusFound {
		t.Fatalf("expected a redirect from login, got %d (%s)", recorder.Code, recorder.Body.String())
	}

	location, err := url.Parse(recorder.Header().Get("Location"))
	if err != nil {
		t.Fatalf("failed to parse authorization URL: %v", err)
	}

	state := location.Query().Get("state")
	if len(state) == 0 {
		t.Fatal("login did not put a state on the authorization URL")
	}

	return state, recorder.Result().Cookies()
}

func claimsFor(idp *fakeIdp, subject string, nonce string, preferredUsername string) jwt.MapClaims {
	claims := idp.baseClaims(subject, nonce)
	claims["preferred_username"] = preferredUsername
	claims["name"] = preferredUsername
	claims["email"] = preferredUsername + "@example.com"

	return claims
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}

	return nil
}

var _ = commands.UpsertOidcProviderCommand{}

// TestCallbackRefusesASessionWithNoBinding isolates the fail-closed half of the
// binding defense.
//
// createAuthSession sets a hash on every session, so in normal operation this
// can't happen -- which is exactly why it needs its own test. The check used to
// read `len(session.BindingHash) > 0 && !bindingMatches(...)`, so an unbound
// session was waved straight through rather than refused, and anything that
// managed to create one (a code path that forgot, a row left by an older build)
// silently lost the defense. A missing binding is a bug, not a mode.
func TestCallbackRefusesASessionWithNoBinding(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{allowProvisioning: true})

	recorder := startLogin(t, provider.Name, "")
	state, cookies := loginAndExtractFrom(t, recorder)
	idp.setClaims(claimsFor(idp, "no-binding-subject", nonceFromAuthUrl(t, recorder), "nobinding"))

	// Strip the binding the session was created with, leaving the row in the shape
	// the old escape hatch would have accepted.
	err := repositories.GetDB().
		Model(&models.OidcAuthSession{}).
		Where("state_hash = ?", hashSecret(state)).
		Update("binding_hash", "").Error
	if err != nil {
		t.Fatalf("failed to blank the binding hash: %v", err)
	}

	// Even WITH the original cookie, an unbound session must not complete.
	callback := runCallback(t, provider.Name, url.Values{
		"code":  {"abc"},
		"state": {state},
	}, cookies)

	if code := oidcErrorCode(t, callback); code != errInvalidState {
		t.Errorf("expected %q for a session with no binding, got %q", errInvalidState, code)
	}
}
