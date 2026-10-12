package oidc

import (
	"receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"
	"time"
)

const (
	// authSessionTTL bounds one in-flight login. It has to cover typing
	// credentials at the identity provider plus an MFA prompt, and every extra
	// minute widens the window in which a stolen state is still redeemable. Same
	// order of magnitude as the OAuth server's own authorization-code TTL.
	authSessionTTL = 10 * time.Minute

	// exchangeCodeTTL bounds the mobile handoff. The app already holds the browser
	// result and the redemption is a single local POST, so this is deliberately
	// tight: the code buys a full session with no further proof of user.
	exchangeCodeTTL = 2 * time.Minute

	// linkLaunchTTL bounds the gap between an app asking to connect an account and
	// the external browser actually opening. That is one hand-off with no user
	// interaction in between, so it can be tighter than the login TTL. Kept
	// separate from exchangeCodeTTL rather than shared: these bound different
	// things and should be free to diverge.
	linkLaunchTTL = 2 * time.Minute

	// randomTokenBytes is the entropy behind state, nonce, the browser binding,
	// the exchange code and the link launch handle.
	randomTokenBytes = 32
)

// newAuthSessionParams carries what createAuthSession needs to persist.
type newAuthSessionParams struct {
	ProviderId          uint
	ClientType          string
	MobileCodeChallenge string
	LinkUserId          *uint
	CodeVerifier        string
}

// createdAuthSession holds the raw secrets, which exist only in memory and in
// the redirect this request is about to write.
type createdAuthSession struct {
	State   string
	Nonce   string
	Binding string
}

// createAuthSession mints the state, nonce and browser binding, and persists the
// session with the secrets HASHED.
//
// Hashing matters because these are bearer secrets we only ever compare: a
// database dump, replica or backup then cannot be used to complete somebody's
// pending login. The PKCE verifier is the one value that cannot be hashed -- it
// has to be replayed to the identity provider verbatim -- so it is encrypted
// instead, the same treatment the IMAP password and AI provider key get.
func createAuthSession(params newAuthSessionParams) (createdAuthSession, error) {
	state, err := utils.GetRandomUrlSafeString(randomTokenBytes)
	if err != nil {
		return createdAuthSession{}, err
	}

	nonce, err := utils.GetRandomUrlSafeString(randomTokenBytes)
	if err != nil {
		return createdAuthSession{}, err
	}

	binding, err := utils.GetRandomUrlSafeString(randomTokenBytes)
	if err != nil {
		return createdAuthSession{}, err
	}

	encryptedVerifier, err := utils.EncryptAndEncodeToBase64(env.GetEncryptionKey(), params.CodeVerifier)
	if err != nil {
		return createdAuthSession{}, err
	}

	session := models.OidcAuthSession{
		OidcProviderId:      params.ProviderId,
		StateHash:           hashSecret(state),
		NonceHash:           hashSecret(nonce),
		CodeVerifier:        encryptedVerifier,
		ClientType:          params.ClientType,
		MobileCodeChallenge: params.MobileCodeChallenge,
		LinkUserId:          params.LinkUserId,
		ExpiresAt:           time.Now().Add(authSessionTTL),
	}

	// EVERY session is bound to the user agent that started it, mobile included.
	//
	// This used to be skipped for mobile, on the reasoning that an external user
	// agent may not carry a cookie. That was wrong: the app hands the login URL
	// itself to flutter_web_auth_2, so the external browser -- not the app -- is
	// what fetches /oidc/{name}/login, and it takes the Set-Cookie and returns it
	// on the callback exactly as the desktop browser does. The mobile LINK start is
	// the one request the app really does make itself, which is why it creates no
	// session at all and hands out a launch handle instead; the browser then calls
	// the launch endpoint and lands here like everyone else.
	//
	// PKCE does not substitute for this. Binding stops an attacker DRIVING your
	// flow; PKCE stops a rogue app STEALING the result of it. They defend different
	// things and the mobile login leg needs both.
	session.BindingHash = hashSecret(binding)

	err = repositories.NewOidcSessionRepository(nil).CreateAuthSession(&session)
	if err != nil {
		return createdAuthSession{}, err
	}

	return createdAuthSession{State: state, Nonce: nonce, Binding: binding}, nil
}

func consumeAuthSession(state string) (models.OidcAuthSession, bool, error) {
	return repositories.NewOidcSessionRepository(nil).ConsumeAuthSession(hashSecret(state))
}

func decryptVerifier(session models.OidcAuthSession) (string, error) {
	return utils.DecryptB64EncodedData(env.GetEncryptionKey(), session.CodeVerifier)
}

// createExchangeCode issues the mobile handoff code, bound to the challenge the
// app generated before the flow started.
func createExchangeCode(userId uint, codeChallenge string) (string, error) {
	code, err := utils.GetRandomUrlSafeString(randomTokenBytes)
	if err != nil {
		return "", err
	}

	row := models.OidcExchangeCode{
		CodeHash:      hashSecret(code),
		UserId:        userId,
		CodeChallenge: codeChallenge,
		ExpiresAt:     time.Now().Add(exchangeCodeTTL),
	}

	err = repositories.NewOidcSessionRepository(nil).CreateExchangeCode(&row)
	if err != nil {
		return "", err
	}

	return code, nil
}

func getExchangeCode(code string) (models.OidcExchangeCode, error) {
	return repositories.NewOidcSessionRepository(nil).GetExchangeCode(hashSecret(code))
}

func consumeExchangeCode(code string) (bool, error) {
	return repositories.NewOidcSessionRepository(nil).ConsumeExchangeCode(hashSecret(code))
}

// createLinkLaunch issues the handle that lets the external browser start a
// mobile link on behalf of an already-authenticated user.
func createLinkLaunch(userId uint, providerId uint) (string, error) {
	handle, err := utils.GetRandomUrlSafeString(randomTokenBytes)
	if err != nil {
		return "", err
	}

	row := models.OidcLinkLaunch{
		LaunchHash:     hashSecret(handle),
		UserId:         userId,
		OidcProviderId: providerId,
		ExpiresAt:      time.Now().Add(linkLaunchTTL),
	}

	err = repositories.NewOidcSessionRepository(nil).CreateLinkLaunch(&row)
	if err != nil {
		return "", err
	}

	return handle, nil
}

func consumeLinkLaunch(handle string) (models.OidcLinkLaunch, bool, error) {
	return repositories.NewOidcSessionRepository(nil).ConsumeLinkLaunch(hashSecret(handle))
}

// hashSecret is the single hashing entry point for every OIDC bearer secret, so
// the write side and the lookup side cannot drift.
func hashSecret(raw string) string {
	return utils.Sha256Hash([]byte(raw))
}
