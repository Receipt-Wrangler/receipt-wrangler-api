package oidc

import (
	"errors"
	"regexp"
	"strings"

	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"time"

	"gorm.io/gorm"
)

var (
	// ErrNoAccount means the identity is unknown and nothing was configured that
	// would let it in: username linking is off (or matched nothing) and
	// provisioning is off.
	ErrNoAccount = errors.New("no receipt wrangler account is linked to this identity")

	// ErrAccountExists means provisioning would have had to invent a username that
	// an existing, unlinked local account already holds.
	//
	// This deliberately refuses rather than suffixing to "bob-2". On a self-hosted
	// install with a single identity provider, a username collision is almost
	// always the same human who simply has not linked yet -- and silently dropping
	// them into a second, empty account is indistinguishable from data loss from
	// their side. The remedy is one click: sign in normally and connect the
	// provider from the profile page.
	ErrAccountExists = errors.New("a local account already uses this username")

	// ErrIdentityLinkedElsewhere means this identity is already attached to a
	// different local user. Never silently re-point it.
	ErrIdentityLinkedElsewhere = errors.New("this provider identity is already linked to another account")

	// ErrAlreadyLinked means the caller already connected this provider.
	ErrAlreadyLinked = errors.New("this account is already linked to this provider")

	// ErrUserIsDummy means the resolved account is a placeholder that cannot log in.
	ErrUserIsDummy = errors.New("account cannot sign in")

	// ErrWouldLockOut means unlinking would strand a provisioned account, which
	// has no usable password to fall back on.
	ErrWouldLockOut = errors.New("this account has no password to fall back on")
)

// usernameSanitizer strips everything a username may not contain. The result is
// only ever a STARTING POINT for a provisioned account -- it is never used to
// resolve an existing identity.
var usernameSanitizer = regexp.MustCompile(`[^a-z0-9._-]+`)

// resolveUser turns a verified set of claims into a local user.
//
// The lookup order is the whole security model of this feature:
//
//  1. (provider, sub) link hit -> that user. This is the only path a returning
//     user ever takes, and it consults no other claim, so renaming at the identity
//     provider cannot re-point an account.
//  2. miss, and the provider opts into username linking -> attach to the local
//     account whose username equals preferred_username. Off by default: OIDC does
//     not promise preferred_username is stable or unique, and public providers
//     recycle released names.
//  3. miss, and the provider opts into provisioning -> create an account.
//  4. otherwise -> refuse.
//
// In every branch that succeeds the subject is persisted, so step 1 handles every
// subsequent login.
func resolveUser(provider models.OidcProvider, claims idTokenClaims) (structs.UserView, error) {
	identityRepository := repositories.NewOidcIdentityRepository(nil)
	userRepository := repositories.NewUserRepository(nil)

	identity, err := identityRepository.GetIdentityBySubject(provider.ID, claims.Subject)
	if err == nil {
		user, userErr := userRepository.GetUserById(identity.UserId)
		if userErr != nil {
			return structs.UserView{}, userErr
		}

		if user.IsDummyUser {
			return structs.UserView{}, ErrUserIsDummy
		}

		touchErr := identityRepository.TouchIdentity(identity.ID, claims.PreferredUsername, claims.Email)
		if touchErr != nil {
			return structs.UserView{}, touchErr
		}

		return user, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return structs.UserView{}, err
	}

	if provider.LinkByUsername && len(claims.PreferredUsername) > 0 {
		user, linkErr := linkByUsername(provider, claims)
		if linkErr == nil {
			return user, nil
		}

		// A miss simply falls through to provisioning; anything else is real.
		if !errors.Is(linkErr, gorm.ErrRecordNotFound) {
			return structs.UserView{}, linkErr
		}
	}

	if provider.AllowProvisioning {
		return provisionUser(provider, claims)
	}

	return structs.UserView{}, ErrNoAccount
}

// usernameMatchesClaim re-asserts, byte for byte in Go, that a row the database
// handed back really does carry the username the claim asked for.
//
// It exists because the lookup that produced that row compared in SQL, and
// users.username pins no collation -- so the same claim resolves differently per
// engine. MySQL and MariaDB, both supported here, default to a case-INSENSITIVE
// collation, where an identity provider account named "ADMIN" resolves the local
// "admin"; SQLite and Postgres compare case-sensitively and would not. This path
// is premised on an EXACT, opt-in username match, and a fuzzy one is precisely
// the account-takeover vector OidcProvider.LinkByUsername's doc comment warns
// about. A boundary that holds on one supported engine and not another is not a
// boundary, so the decisive comparison is made here rather than left to the
// column.
//
// Kept as a named function rather than inlined because the test database is
// SQLite: an inline check would be unreachable there, so the rule would ship
// with no test that can fail on it.
func usernameMatchesClaim(storedUsername string, claimedUsername string) bool {
	return storedUsername == claimedUsername
}

// linkByUsername attaches an unseen identity to an existing local account whose
// username matches the preferred_username claim.
func linkByUsername(provider models.OidcProvider, claims idTokenClaims) (structs.UserView, error) {
	identityRepository := repositories.NewOidcIdentityRepository(nil)

	user, err := repositories.NewUserRepository(nil).GetUserByUsername(claims.PreferredUsername)
	if err != nil {
		return structs.UserView{}, err
	}

	// Treated as a miss rather than an error, so an inexact match falls through to
	// provisioning exactly like a lookup that found nothing.
	if !usernameMatchesClaim(user.Username, claims.PreferredUsername) {
		return structs.UserView{}, gorm.ErrRecordNotFound
	}

	// A dummy user is a placeholder that is blocked from logging in anyway;
	// attaching an identity to one would only produce a confusing failure later.
	if user.IsDummyUser {
		return structs.UserView{}, gorm.ErrRecordNotFound
	}

	// If this account already holds a DIFFERENT subject at this provider, the
	// username match is pointing at the wrong person -- refuse rather than create a
	// second identity (which the unique index would reject anyway).
	_, err = identityRepository.GetIdentityForUser(provider.ID, user.ID)
	if err == nil {
		return structs.UserView{}, ErrIdentityLinkedElsewhere
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return structs.UserView{}, err
	}

	err = createIdentity(provider, user.ID, claims, false)
	if err != nil {
		return structs.UserView{}, err
	}

	return user, nil
}

// provisionUser creates a local account for an unseen identity.
func provisionUser(provider models.OidcProvider, claims idTokenClaims) (structs.UserView, error) {
	username := deriveUsername(provider, claims)

	// Username uniqueness is normally enforced by middleware.ValidateUserData,
	// which this path does not go through, so check here -- and refuse rather than
	// suffix (see ErrAccountExists).
	_, err := repositories.NewUserRepository(nil).GetUserByUsername(username)
	if err == nil {
		return structs.UserView{}, ErrAccountExists
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return structs.UserView{}, err
	}

	// The password column is NOT NULL and CreateUser bcrypts whatever it is handed.
	// A sentinel such as "!oidc" would therefore store bcrypt("!oidc") and let
	// anyone who knows the sentinel sign in as this account through the normal
	// login form -- and through the MCP OAuth login form, which shares
	// services.LoginUser. So: generate real randomness, hand it over, and discard
	// it. Nothing anywhere retains a value that VerifyPassword will accept.
	password, err := utils.GetRandomUrlSafeString(48)
	if err != nil {
		return structs.UserView{}, err
	}

	displayName := claims.Name
	if len(displayName) == 0 {
		displayName = username
	}

	created, err := repositories.NewUserRepository(nil).CreateUser(commands.SignUpCommand{
		Username:    username,
		Password:    password,
		DisplayName: displayName,
		IsDummyUser: false,
		// Nil so the account picks up the configured default app role, exactly as a
		// local sign-up does. A provisioned user is deliberately never an admin.
		AppRoleID: nil,
	})
	if err != nil {
		return structs.UserView{}, err
	}

	err = createIdentity(provider, created.ID, claims, true)
	if err != nil {
		return structs.UserView{}, err
	}

	return structs.UserView{
		ID:                 created.ID,
		Username:           created.Username,
		DisplayName:        created.DisplayName,
		DefaultAvatarColor: created.DefaultAvatarColor,
		IsDummyUser:        created.IsDummyUser,
		AppRoleID:          created.AppRoleID,
	}, nil
}

func createIdentity(provider models.OidcProvider, userId uint, claims idTokenClaims, provisioned bool) error {
	now := time.Now()

	identity := models.OidcIdentity{
		OidcProviderId:    provider.ID,
		Subject:           claims.Subject,
		UserId:            userId,
		PreferredUsername: claims.PreferredUsername,
		Email:             claims.Email,
		ProvisionedUser:   provisioned,
		LastLoginAt:       &now,
	}

	return repositories.NewOidcIdentityRepository(nil).CreateIdentity(&identity)
}

// deriveUsername builds a candidate username from the claims, preferring the
// provider's own username, then the local part of the email, then the subject.
func deriveUsername(provider models.OidcProvider, claims idTokenClaims) string {
	candidates := []string{
		claims.PreferredUsername,
		emailLocalPart(claims.Email),
		provider.Name + "-" + claims.Subject,
	}

	for _, candidate := range candidates {
		sanitized := sanitizeUsername(candidate)
		if len(sanitized) >= 3 {
			return sanitized
		}
	}

	return sanitizeUsername(provider.Name + "-user")
}

func sanitizeUsername(raw string) string {
	sanitized := usernameSanitizer.ReplaceAllString(strings.ToLower(strings.TrimSpace(raw)), "")
	sanitized = strings.Trim(sanitized, "._-")

	if len(sanitized) > 40 {
		sanitized = strings.Trim(sanitized[:40], "._-")
	}

	return sanitized
}

func emailLocalPart(email string) string {
	at := strings.Index(email, "@")
	if at <= 0 {
		return ""
	}

	return email[:at]
}

// resolveLink attaches a verified identity to an ALREADY AUTHENTICATED user.
//
// This is what makes linkByUsername safe to leave off. The caller proved who they
// are with a session, so nothing has to be inferred from a claim: the whole
// match/provision decision tree above is skipped and the link is written
// directly. Neither AllowProvisioning nor LinkByUsername is consulted here.
func resolveLink(provider models.OidcProvider, claims idTokenClaims, userId uint) error {
	identityRepository := repositories.NewOidcIdentityRepository(nil)

	// This identity may not already belong to somebody else. Re-pointing it would
	// silently transfer an account.
	existing, err := identityRepository.GetIdentityBySubject(provider.ID, claims.Subject)
	if err == nil {
		if existing.UserId == userId {
			return ErrAlreadyLinked
		}

		return ErrIdentityLinkedElsewhere
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// And the caller may not already hold a different identity at this provider.
	_, err = identityRepository.GetIdentityForUser(provider.ID, userId)
	if err == nil {
		return ErrAlreadyLinked
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// The new identity INHERITS whether the account is passwordless, rather than
	// assuming a linked account always has a password of its own.
	//
	// This is what keeps the lockout guard honest. The flag used to mean "this
	// identity created the account", so it lived on exactly one row -- and
	// unlinking that row first destroyed the only record that the account has no
	// password, letting the next unlink strand it. Held on every identity, the fact
	// survives until the last one goes.
	passwordless, err := accountIsPasswordless(userId)
	if err != nil {
		return err
	}

	return createIdentity(provider, userId, claims, passwordless)
}

// accountIsPasswordless reports whether this account still has no password of its
// own -- it was created by an OIDC provisioning login and nobody has set one
// since. Such an account cannot afford to lose its last identity.
func accountIsPasswordless(userId uint) (bool, error) {
	count, err := repositories.NewOidcIdentityRepository(nil).CountProvisionedIdentitiesForUser(userId)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// UnlinkIdentity removes one of the caller's connected providers.
//
// The guard is the reason ProvisionedUser exists: an account created by OIDC only
// ever had a random, discarded password, so unlinking its LAST identity would
// strand it with no way back in. Refuse and point at an administrator instead.
//
// The question is asked of the ACCOUNT, never of the row being removed. Asking
// the row was the original bug: with a provisioned Google and a later-linked
// GitHub, dropping Google passed the guard (two identities remained) and then
// dropping GitHub passed it too (that row was not the provisioning one), leaving
// an account with no identities and a bcrypt hash of a password nobody knew.
func UnlinkIdentity(userId uint, providerName string) error {
	providerRow, err := repositories.NewOidcProviderRepository(nil).GetOidcProviderByName(providerName)
	if err != nil {
		return err
	}

	identityRepository := repositories.NewOidcIdentityRepository(nil)

	identity, err := identityRepository.GetIdentityForUser(providerRow.ID, userId)
	if err != nil {
		return err
	}

	count, err := identityRepository.CountIdentitiesForUser(userId)
	if err != nil {
		return err
	}

	if count <= 1 {
		passwordless, passwordlessErr := accountIsPasswordless(userId)
		if passwordlessErr != nil {
			return passwordlessErr
		}

		if passwordless {
			return ErrWouldLockOut
		}
	}

	// Removing an identity must not take the account's passwordless fact with it.
	//
	// resolveLink holds that fact on every identity precisely so it survives a
	// delete, but rows written before that invariant existed carry it on only the
	// one that provisioned the account -- and deleting that one first is the
	// original bug in a different disguise. Re-asserting it across what remains
	// heals such an account the first time it is touched, which is why this needs
	// no data migration.
	if identity.ProvisionedUser {
		err = identityRepository.SetProvisionedFlagForUser(userId)
		if err != nil {
			return err
		}
	}

	return identityRepository.DeleteIdentity(identity.ID, userId)
}
