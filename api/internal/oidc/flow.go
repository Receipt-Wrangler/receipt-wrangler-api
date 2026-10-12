package oidc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/permissions"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
)

const (
	// bindingCookieName holds the user-agent-binding secret. Set on every leg,
	// mobile included -- the external browser round-trips it the same way.
	bindingCookieName = "oidc_session"

	// bindingCookiePath scopes the cookie to the OIDC routes: nothing else needs
	// it, and it should not ride along on every API call.
	bindingCookiePath = "/api/oidc"

	// mobileCallbackScheme is the app's private-use URL scheme, reverse-DNS of a
	// domain the project controls per RFC 8252 section 7.1. It carries a one-time,
	// PKCE-bound code -- never a token.
	mobileCallbackScheme = "io.receiptwrangler://oidc"

	// desktopCallbackPath is where the browser lands after a successful login. It
	// is a fixed, relative path: no redirect target is ever taken from the request,
	// which is what keeps this flow free of an open-redirect hole.
	desktopCallbackPath = "/auth/callback"

	// desktopLoginPath is where a failed login lands.
	desktopLoginPath = "/auth/login"

	// desktopProfilePath is where a link flow returns to.
	desktopProfilePath = "/settings/user-profile/view"

	exchangeTimeout = 15 * time.Second
)

// Error codes handed back to a client. A small fixed vocabulary -- an upstream
// identity provider's own error text is never echoed through.
const (
	errUnknownProvider = "unknown_provider"
	errInvalidRequest  = "invalid_request"
	errInvalidState    = "invalid_state"
	errNonceMismatch   = "nonce_mismatch"
	errNoIdToken       = "no_id_token"
	errProviderError   = "provider_error"
	errNoAccount       = "no_account"
	errAccountExists   = "account_exists"
	errAlreadyLinked   = "already_linked"
	errForbidden       = "forbidden"
	errServerError     = "server_error"
)

// Login starts an OIDC login. Unauthenticated.
func Login(w http.ResponseWriter, r *http.Request) {
	startFlow(w, r, nil, resolveClientType(r))
}

// LinkStart starts a "connect account" flow for the authenticated caller. It is
// the same flow as Login, differing only in that the resulting session carries
// the caller's user id -- so the callback links instead of guessing.
//
// Gated on app.account.update, resolved from the database and never from the
// JWT. Connecting a provider ADDS a way to sign into this account, so it is at
// least as privileged as disconnecting one -- and DeleteOidcConnection already
// requires exactly this permission. Without the check an administrator could
// build a role that cannot remove a sign-in method but can still add one, which
// is not a coherent thing to be able to configure.
//
// The check lives in the handler body rather than on a structs.Handler because
// this route answers with a redirect, not JSON: HandleRequest's 403 would drop a
// raw error object into a browser the user navigated here. Same shape as
// GetPagedApiKeys' app.api-keys.read-any check.
func LinkStart(w http.ResponseWriter, r *http.Request) {
	claims := structs.GetClaims(r)
	userId := claims.UserId

	clientType := resolveClientType(r)

	allowed, err := services.NewPermissionService(nil).HasAppPermissions(userId, permissions.AppAccountUpdate)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Failed to resolve permissions for an OIDC link: "+err.Error())
		failFlowStart(w, r, clientType, true, errServerError)
		return
	}

	if !allowed {
		failFlowStart(w, r, clientType, true, errForbidden)
		return
	}

	// Mobile is the one caller that is not a browser: the app makes this request
	// itself, with a bearer token, because an external user agent cannot carry one.
	// That means there is nobody here to hand a binding cookie to -- so this start
	// deliberately creates NO session. It hands back a single-use launch handle and
	// lets the external browser start the real flow at LinkLaunch, where it gets
	// bound like every other leg.
	if clientType == models.OidcClientMobile {
		startMobileLinkLaunch(w, r, userId)
		return
	}

	startFlow(w, r, &userId, clientType)
}

// startMobileLinkLaunch answers a mobile link start with the URL the app should
// open in the external browser. The handle in it is the browser's authorization
// to start a link for this user, and it is single-use and short-lived.
//
// It is safe to put in a URL the browser navigates to: a 302 does not make the
// redirecting URL the next request's Referer, so the handle never reaches the
// identity provider the way the authorization URL's own state does. That Referer
// leak is exactly the attack this whole hop exists to defeat.
func startMobileLinkLaunch(w http.ResponseWriter, r *http.Request, userId uint) {
	name := chi.URLParam(r, "name")

	providerRow, err := repositories.NewOidcProviderRepository(nil).GetEnabledOidcProviderByName(name)
	if err != nil {
		writeOidcFlowError(w, errUnknownProvider)
		return
	}

	handle, err := createLinkLaunch(userId, providerRow.ID)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Failed to create an OIDC link launch: "+err.Error())
		writeOidcFlowError(w, errServerError)
		return
	}

	writeOidcJson(w, http.StatusOK, structs.OidcLinkStartView{
		LaunchUrl: services.BuildOidcLinkLaunchUri(providerRow.Name, handle),
	})
}

// LinkLaunch is the browser-facing half of a mobile link, and the reason the
// mobile leg can be bound at all.
//
// Unauthenticated on purpose: the single-use launch handle IS the authorization.
// An authenticated LinkStart minted it moments ago for this user, after checking
// app.account.update. Consuming it here lets the EXTERNAL BROWSER be the thing
// that creates the auth session, which is what makes a binding cookie possible.
func LinkLaunch(w http.ResponseWriter, r *http.Request) {
	handle := strings.TrimSpace(r.URL.Query().Get("h"))
	if len(handle) == 0 {
		redirectWithError(w, r, models.OidcClientMobile, errInvalidRequest)
		return
	}

	launch, claimed, err := consumeLinkLaunch(handle)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Failed to consume an OIDC link launch: "+err.Error())
		redirectWithError(w, r, models.OidcClientMobile, errServerError)
		return
	}

	// Unknown, expired or already spent all collapse to the same answer, so a
	// caller cannot probe which handles exist.
	if !claimed {
		redirectWithError(w, r, models.OidcClientMobile, errInvalidState)
		return
	}

	// The handle is bound to one provider. Presenting it at another provider's
	// launch URL is a caller confusing itself rather than an attack -- only the
	// user who asked for it ever held it -- but the mismatch is still not a thing
	// we should act on.
	providerRow, err := repositories.NewOidcProviderRepository(nil).GetEnabledOidcProviderByName(chi.URLParam(r, "name"))
	if err != nil || providerRow.ID != launch.OidcProviderId {
		redirectWithError(w, r, models.OidcClientMobile, errInvalidState)
		return
	}

	startFlow(w, r, &launch.UserId, models.OidcClientMobile)
}

// resolveClientType reads which client is driving the flow. Anything that is not
// an explicit "mobile" is a browser.
func resolveClientType(r *http.Request) string {
	if r.URL.Query().Get("client") == models.OidcClientMobile {
		return models.OidcClientMobile
	}

	return models.OidcClientDesktop
}

// startFlow creates the auth session and sends the user agent to the identity
// provider. Every caller is a BROWSER -- a login navigation, a desktop link
// navigation, or LinkLaunch -- which is what lets it bind the session to the
// agent unconditionally. The client type is passed in rather than sniffed from
// the query because LinkLaunch knows it from the handle, not from the URL.
func startFlow(w http.ResponseWriter, r *http.Request, linkUserId *uint, clientType string) {
	name := chi.URLParam(r, "name")

	isLink := linkUserId != nil

	providerRow, err := repositories.NewOidcProviderRepository(nil).GetEnabledOidcProviderByName(name)
	if err != nil {
		failBrowserFlowStart(w, r, clientType, isLink, errUnknownProvider)
		return
	}

	mobileChallenge := ""

	if clientType == models.OidcClientMobile && !isLink {
		mobileChallenge = strings.TrimSpace(r.URL.Query().Get("codeChallenge"))

		// A mobile LOGIN hands the app a one-time exchange code, which buys a full
		// session. The app's PKCE challenge is what ties that code to the app that
		// started the flow, so a rogue app claiming the private-use scheme gets
		// something it cannot redeem. This is separate from the binding cookie above,
		// which stops an attacker DRIVING the flow rather than stealing its result.
		//
		// A mobile LINK needs none: its callback hands back only ?linked={name},
		// which is not a credential. Nothing is minted for the app to protect.
		if len(mobileChallenge) == 0 {
			failBrowserFlowStart(w, r, clientType, isLink, errInvalidRequest)
			return
		}
	}

	discovered, err := GetProvider(providerRow)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "OIDC discovery failed for provider "+providerRow.Name+": "+err.Error())
		failBrowserFlowStart(w, r, clientType, isLink, errProviderError)
		return
	}

	config, err := buildOauthConfig(providerRow, discovered, services.BuildOidcRedirectUri(providerRow.Name))
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Failed to build OIDC config for provider "+providerRow.Name+": "+err.Error())
		failBrowserFlowStart(w, r, clientType, isLink, errServerError)
		return
	}

	verifier := oauth2.GenerateVerifier()

	created, err := createAuthSession(newAuthSessionParams{
		ProviderId:          providerRow.ID,
		ClientType:          clientType,
		MobileCodeChallenge: mobileChallenge,
		LinkUserId:          linkUserId,
		CodeVerifier:        verifier,
	})
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Failed to create OIDC auth session: "+err.Error())
		failBrowserFlowStart(w, r, clientType, isLink, errServerError)
		return
	}

	// Unconditional, mobile included. The authorization URL about to be built
	// carries the state, and identity providers leak it -- their login pages load
	// third-party resources and hand the whole URL over as a Referer. This cookie
	// is what makes a leaked state worthless: the attacker cannot produce it, so
	// their callback is refused.
	http.SetCookie(w, buildBindingCookie(created.Binding))

	authUrl := config.AuthCodeURL(
		created.State,
		oidc.Nonce(created.Nonce),
		oauth2.S256ChallengeOption(verifier),
	)

	http.Redirect(w, r, authUrl, http.StatusFound)
}

// Callback is where the identity provider returns the user. Every check below is
// load-bearing; they run in this order and fail closed.
func Callback(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	providerRow, err := repositories.NewOidcProviderRepository(nil).GetEnabledOidcProviderByName(name)
	if err != nil {
		redirectWithError(w, r, models.OidcClientDesktop, errUnknownProvider)
		return
	}

	// 1. The identity provider itself refused. Nothing is consumed; the pending
	// session simply expires. Never echo the upstream text.
	if len(r.URL.Query().Get("error")) > 0 {
		clearBindingCookie(w, r)
		redirectWithError(w, r, models.OidcClientDesktop, errProviderError)
		return
	}

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if len(state) == 0 || len(code) == 0 {
		clearBindingCookie(w, r)
		redirectWithError(w, r, models.OidcClientDesktop, errInvalidRequest)
		return
	}

	// 2. Claim the session ATOMICALLY, before any network call. A replayed state
	// therefore cannot even reach the identity provider, let alone mint a second
	// session. Unknown, expired and already-used states are indistinguishable.
	session, claimed, err := consumeAuthSession(state)
	clearBindingCookie(w, r)

	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Failed to consume OIDC auth session: "+err.Error())
		redirectWithError(w, r, models.OidcClientDesktop, errServerError)
		return
	}

	if !claimed {
		redirectWithError(w, r, models.OidcClientDesktop, errInvalidState)
		return
	}

	clientType := session.ClientType

	// 3. The state must belong to THIS provider. Otherwise a state minted for a
	// permissive provider could be redeemed at a stricter one's callback.
	if session.OidcProviderId != providerRow.ID {
		redirectWithError(w, r, clientType, errInvalidState)
		return
	}

	// 4. User-agent binding -- the login-CSRF defense, and the single most
	// load-bearing check here. Without it an attacker who obtains the state (the
	// identity provider's own login page leaks the authorization URL by Referer)
	// can complete the flow as THEMSELVES and have the result land on the victim:
	// signed into the attacker's account on a login, or the attacker's identity
	// grafted onto the victim's account on a link.
	//
	// Unconditional by design. This used to read `len(session.BindingHash) > 0 &&`,
	// which turned the check into a no-op for every mobile session, because
	// createAuthSession skipped the hash for mobile. No session is created unbound
	// any more, so an empty hash is a bug rather than a mode, and it must fail
	// closed here rather than wave the callback through.
	if !bindingMatches(r, session.BindingHash) {
		redirectWithError(w, r, clientType, errInvalidState)
		return
	}

	verifier, err := decryptVerifier(session)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Failed to decrypt OIDC code verifier: "+err.Error())
		redirectWithError(w, r, clientType, errServerError)
		return
	}

	discovered, err := GetProvider(providerRow)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "OIDC discovery failed on callback: "+err.Error())
		redirectWithError(w, r, clientType, errProviderError)
		return
	}

	config, err := buildOauthConfig(providerRow, discovered, services.BuildOidcRedirectUri(providerRow.Name))
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Failed to build OIDC config on callback: "+err.Error())
		redirectWithError(w, r, clientType, errServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), exchangeTimeout)
	defer cancel()
	ctx = oidc.ClientContext(ctx, &http.Client{Timeout: exchangeTimeout})

	// 5. Exchange the code, proving possession of our PKCE verifier.
	token, err := config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "OIDC code exchange failed: "+err.Error())
		redirectWithError(w, r, clientType, errProviderError)
		return
	}

	// 6. No ID token means this was an OAuth response, not an OIDC one -- there is
	// no verified identity to act on. Hard fail rather than falling back to
	// userinfo, which is not signed.
	rawIdToken, ok := token.Extra("id_token").(string)
	if !ok || len(rawIdToken) == 0 {
		redirectWithError(w, r, clientType, errNoIdToken)
		return
	}

	// 7. Verify signature (via JWKS), issuer, audience and expiry.
	idToken, err := discovered.Verifier(&oidc.Config{ClientID: providerRow.ClientId}).Verify(ctx, rawIdToken)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "OIDC ID token verification failed: "+err.Error())
		redirectWithError(w, r, clientType, errProviderError)
		return
	}

	// 8. Verify the nonce OURSELVES. go-oidc documents that Verify does NOT do
	// nonce validation -- "that is the caller's responsibility" -- and skipping it
	// re-opens ID token replay. The empty check matters independently: an empty
	// nonce hashes to a fixed value, and must never be allowed to match.
	if len(idToken.Nonce) == 0 || !utils.SecureCompare(hashSecret(idToken.Nonce), session.NonceHash) {
		redirectWithError(w, r, clientType, errNonceMismatch)
		return
	}

	claims, err := extractClaims(ctx, discovered, config, token, idToken)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Failed to read OIDC claims: "+err.Error())
		redirectWithError(w, r, clientType, errProviderError)
		return
	}

	if len(claims.Subject) == 0 {
		redirectWithError(w, r, clientType, errProviderError)
		return
	}

	if session.IsLink() {
		finishLink(w, r, providerRow, claims, session)
		return
	}

	finishLogin(w, r, providerRow, claims, session)
}

// extractClaims decodes the ID token's claims, falling back to the userinfo
// endpoint only to fill in an email the ID token omitted. Identity itself always
// comes from the signed ID token.
func extractClaims(
	ctx context.Context,
	discovered *oidc.Provider,
	config *oauth2.Config,
	token *oauth2.Token,
	idToken *oidc.IDToken,
) (idTokenClaims, error) {
	var raw json.RawMessage
	if err := idToken.Claims(&raw); err != nil {
		return idTokenClaims{}, err
	}

	claims, err := decodeIdTokenClaims(raw)
	if err != nil {
		return idTokenClaims{}, err
	}

	if len(claims.Email) > 0 || len(discovered.UserInfoEndpoint()) == 0 {
		return claims, nil
	}

	// Best effort: several providers (Twitch notably) will not put email in the ID
	// token. Email is display-only here, so a failure is logged and ignored rather
	// than failing an otherwise valid login.
	info, err := discovered.UserInfo(ctx, config.TokenSource(ctx, token))
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "OIDC userinfo lookup failed: "+err.Error())
		return claims, nil
	}

	if info.Subject == claims.Subject {
		claims.Email = info.Email
		claims.EmailVerified = tolerantBool(info.EmailVerified)
	}

	return claims, nil
}

func finishLogin(
	w http.ResponseWriter,
	r *http.Request,
	providerRow models.OidcProvider,
	claims idTokenClaims,
	session models.OidcAuthSession,
) {
	user, err := resolveUser(providerRow, claims)
	if err != nil {
		redirectWithError(w, r, session.ClientType, resolutionErrorCode(err))
		return
	}

	issueSession(w, r, user.ID, session)
}

func finishLink(
	w http.ResponseWriter,
	r *http.Request,
	providerRow models.OidcProvider,
	claims idTokenClaims,
	session models.OidcAuthSession,
) {
	err := resolveLink(providerRow, claims, *session.LinkUserId)
	if err != nil {
		redirectLinkError(w, r, session.ClientType, resolutionErrorCode(err))
		return
	}

	redirectLinkSuccess(w, r, session.ClientType, providerRow.Name)
}

// redirectLinkSuccess returns a completed "connect account" flow to wherever it
// was started from. The mobile app is waiting on its private-use scheme, so it
// gets the same redirect a mobile login would -- carrying only the provider's
// name, since a link mints no session and there is nothing to hand over.
func redirectLinkSuccess(w http.ResponseWriter, r *http.Request, clientType string, providerName string) {
	if clientType == models.OidcClientMobile {
		http.Redirect(w, r, mobileCallbackScheme+"?linked="+url.QueryEscape(providerName), http.StatusFound)
		return
	}

	http.Redirect(w, r, desktopProfilePath+"?tab=user-profile&oidcLinked="+url.QueryEscape(providerName), http.StatusFound)
}

// redirectLinkError reports a failed "connect account" attempt on the profile
// page the user started from.
//
// Deliberately NOT redirectWithError, which lands on the login screen: a link
// flow's caller is already signed in, so bouncing them to a login form reads as
// "you have been signed out" rather than "that provider could not be connected".
func redirectLinkError(w http.ResponseWriter, r *http.Request, clientType string, code string) {
	if clientType == models.OidcClientMobile {
		redirectWithError(w, r, clientType, code)
		return
	}

	http.Redirect(w, r, desktopProfilePath+"?tab=user-profile&oidcError="+url.QueryEscape(code), http.StatusFound)
}

// issueSession mints Receipt Wrangler tokens and hands them to the client in the
// way that client can actually receive them.
func issueSession(w http.ResponseWriter, r *http.Request, userId uint, session models.OidcAuthSession) {
	if session.ClientType == models.OidcClientMobile {
		// Deliberately NOT the tokens themselves: a private-use scheme is
		// unverifiable on Android, so any installed app can read this redirect. The
		// code is single-use, short-lived and PKCE-bound, so intercepting it is
		// useless without the verifier the app kept.
		code, err := createExchangeCode(userId, session.MobileCodeChallenge)
		if err != nil {
			logging.LogStd(logging.LOG_LEVEL_ERROR, "Failed to create OIDC exchange code: "+err.Error())
			redirectWithError(w, r, session.ClientType, errServerError)
			return
		}

		http.Redirect(w, r, mobileCallbackScheme+"?code="+url.QueryEscape(code), http.StatusFound)
		return
	}

	jwt, refreshToken, _, err := services.GenerateJWT(userId)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Failed to mint tokens after OIDC login: "+err.Error())
		redirectWithError(w, r, session.ClientType, errServerError)
		return
	}

	_, err = repositories.NewUserRepository(nil).UpdateUserLastLoginDate(userId)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Failed to stamp last login date after OIDC login: "+err.Error())
	}

	accessTokenCookie, refreshTokenCookie := services.BuildTokenCookies(jwt, refreshToken)
	http.SetCookie(w, &accessTokenCookie)
	http.SetCookie(w, &refreshTokenCookie)

	http.Redirect(w, r, desktopCallbackPath, http.StatusFound)
}

func resolutionErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrNoAccount), errors.Is(err, ErrUserIsDummy):
		return errNoAccount
	case errors.Is(err, ErrAccountExists):
		return errAccountExists
	case errors.Is(err, ErrAlreadyLinked), errors.Is(err, ErrIdentityLinkedElsewhere):
		return errAlreadyLinked
	default:
		logging.LogStd(logging.LOG_LEVEL_ERROR, "OIDC identity resolution failed: "+err.Error())
		return errServerError
	}
}

// failFlowStart reports a failure from a start that the APP called directly --
// LinkStart, which a mobile client reaches as an ordinary bearer-authenticated
// API request and a desktop client reaches as a browser navigation. The mobile
// case gets JSON so the failure arrives in the same shape as the success.
//
// Not for startFlow: every caller of that is a browser, including LinkLaunch,
// and handing a browser a JSON body would strand it on a blank page.
func failFlowStart(w http.ResponseWriter, r *http.Request, clientType string, isLink bool, code string) {
	if clientType == models.OidcClientMobile && isLink {
		writeOidcFlowError(w, code)
		return
	}

	failBrowserFlowStart(w, r, clientType, isLink, code)
}

// failBrowserFlowStart reports a start failure to a user agent, sending it back
// wherever it came from: the profile page for a link, the login screen for a
// login, and the app's own scheme for either on mobile.
func failBrowserFlowStart(w http.ResponseWriter, r *http.Request, clientType string, isLink bool, code string) {
	if isLink {
		redirectLinkError(w, r, clientType, code)
		return
	}

	redirectWithError(w, r, clientType, code)
}

// writeOidcFlowError writes a flow error code as JSON under a status that
// describes the failure, for the one start that is an API call.
func writeOidcFlowError(w http.ResponseWriter, code string) {
	status := http.StatusInternalServerError

	switch code {
	case errUnknownProvider:
		status = http.StatusNotFound
	case errInvalidRequest:
		status = http.StatusBadRequest
	case errForbidden:
		status = http.StatusForbidden
	case errProviderError:
		status = http.StatusBadGateway
	}

	writeOidcJson(w, status, structs.OidcFlowError{ErrorCode: code})
}

func writeOidcJson(w http.ResponseWriter, status int, body any) {
	bytes, err := json.Marshal(body)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Failed to marshal an OIDC response: "+err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", constants.ApplicationJson)
	w.WriteHeader(status)
	w.Write(bytes)
}

func redirectWithError(w http.ResponseWriter, r *http.Request, clientType string, code string) {
	if clientType == models.OidcClientMobile {
		http.Redirect(w, r, mobileCallbackScheme+"?error="+url.QueryEscape(code), http.StatusFound)
		return
	}

	http.Redirect(w, r, desktopLoginPath+"?oidcError="+url.QueryEscape(code), http.StatusFound)
}

func buildBindingCookie(value string) *http.Cookie {
	secure := false
	if env.GetDeployEnv() == "dev" {
		secure = true
	}

	return &http.Cookie{
		Name:     bindingCookieName,
		Value:    value,
		HttpOnly: true,
		Path:     bindingCookiePath,
		MaxAge:   int(authSessionTTL.Seconds()),
		Secure:   secure,
		// Lax, NOT Strict. BuildTokenCookies uses Strict in production, but the
		// callback arrives as a cross-site top-level GET from the identity provider:
		// Strict would drop this cookie on exactly the request that needs it, and the
		// flow would fail only in production.
		SameSite: http.SameSiteLaxMode,
	}
}

// clearBindingCookie expires the browser binding once it has served its purpose.
//
// It is a no-op when the request did not carry one, so the mobile leg -- which
// never sets a binding cookie, because an external user agent may not carry it --
// emits no Set-Cookie header at all. A native client should get a redirect and
// nothing else.
func clearBindingCookie(w http.ResponseWriter, r *http.Request) {
	if _, err := r.Cookie(bindingCookieName); err != nil {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     bindingCookieName,
		Value:    "",
		HttpOnly: true,
		Path:     bindingCookiePath,
		MaxAge:   -1,
	})
}

func bindingMatches(r *http.Request, bindingHash string) bool {
	cookie, err := r.Cookie(bindingCookieName)
	if err != nil || len(cookie.Value) == 0 {
		return false
	}

	return utils.SecureCompare(hashSecret(cookie.Value), bindingHash)
}
