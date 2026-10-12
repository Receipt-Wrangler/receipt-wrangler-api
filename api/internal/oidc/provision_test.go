package oidc

import (
	"errors"
	"testing"

	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"

	"gorm.io/gorm"
)

func claims(subject string, preferredUsername string) idTokenClaims {
	return idTokenClaims{
		Subject:           subject,
		PreferredUsername: preferredUsername,
		Name:              preferredUsername,
		Email:             preferredUsername + "@example.com",
	}
}

func TestResolveUserRefusesUnknownIdentityWhenNothingIsEnabled(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	createTestUser(t, "existing")

	// Neither provisioning nor username linking is on, so an unseen identity has no
	// way in -- not even one whose username happens to match.
	_, err := resolveUser(provider, claims("unseen-subject", "existing"))
	if !errors.Is(err, ErrNoAccount) {
		t.Errorf("expected ErrNoAccount, got %v", err)
	}
}

func TestResolveUserProvisionsWhenEnabled(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{allowProvisioning: true})

	user, err := resolveUser(provider, claims("new-subject", "newperson"))
	if err != nil {
		t.Fatalf("expected provisioning to succeed, got %v", err)
	}

	if user.Username != "newperson" {
		t.Errorf("expected the username to come from preferred_username, got %q", user.Username)
	}

	identity, err := repositories.NewOidcIdentityRepository(nil).GetIdentityBySubject(provider.ID, "new-subject")
	if err != nil {
		t.Fatalf("expected an identity to be linked: %v", err)
	}

	if !identity.ProvisionedUser {
		t.Error("a provisioned account's identity must be marked as such, or the unlink lockout guard cannot fire")
	}
}

// TestProvisionedPasswordIsUnusable is the single most important test in this
// file.
//
// UserRepository.CreateUser bcrypts whatever it is handed, so a sentinel password
// would become a WORKING password for anyone who knew the sentinel -- through the
// normal login form and through the MCP OAuth login form, which shares
// services.LoginUser. The value handed over must be real randomness, and must be
// discarded.
func TestProvisionedPasswordIsUnusable(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{allowProvisioning: true})

	subject := "provisioned-subject"
	user, err := resolveUser(provider, claims(subject, "provisioned"))
	if err != nil {
		t.Fatalf("expected provisioning to succeed, got %v", err)
	}

	var stored models.User
	if err := repositories.GetDB().Where("id = ?", user.ID).First(&stored).Error; err != nil {
		t.Fatalf("failed to load the provisioned user: %v", err)
	}

	if len(stored.Password) == 0 {
		t.Fatal("the password column must not be empty")
	}

	for _, guess := range []string{
		"",
		"password",
		"oidc",
		"!oidc",
		subject,
		stored.Username,
		provider.Name,
	} {
		if utils.VerifyPassword(stored.Password, guess) == nil {
			t.Errorf("a provisioned account must not accept the password %q", guess)
		}
	}
}

func TestResolveUserLinksByUsernameOnlyWhenEnabled(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{linkByUsername: true})

	existing := createTestUser(t, "alice")

	user, err := resolveUser(provider, claims("alice-subject", "alice"))
	if err != nil {
		t.Fatalf("expected the username match to link, got %v", err)
	}

	if user.ID != existing.ID {
		t.Errorf("expected to land on the existing account %d, got %d", existing.ID, user.ID)
	}

	identity, err := repositories.NewOidcIdentityRepository(nil).GetIdentityBySubject(provider.ID, "alice-subject")
	if err != nil {
		t.Fatalf("expected an identity to be linked: %v", err)
	}

	if identity.ProvisionedUser {
		t.Error("linking to an existing account must not mark it as provisioned -- that account has its own password")
	}
}

// TestSecondLoginUsesSubjectNotUsername is the guarantee that makes
// linkByUsername survivable: it only ever applies to a FIRST login. Once the
// subject is stored, a rename at the identity provider cannot re-point anything.
func TestSecondLoginUsesSubjectNotUsername(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{linkByUsername: true, allowProvisioning: true})

	existing := createTestUser(t, "bob")
	createTestUser(t, "carol")

	first, err := resolveUser(provider, claims("bob-subject", "bob"))
	if err != nil {
		t.Fatalf("first login failed: %v", err)
	}

	// The same person renames themselves at the identity provider to a name that
	// belongs to somebody else here. The subject is unchanged, so the login must
	// still land on the original account.
	second, err := resolveUser(provider, claims("bob-subject", "carol"))
	if err != nil {
		t.Fatalf("second login failed: %v", err)
	}

	if second.ID != first.ID || second.ID != existing.ID {
		t.Errorf("a rename at the identity provider re-pointed the account: first %d, second %d, expected %d", first.ID, second.ID, existing.ID)
	}
}

func TestResolveUserRefusesUsernameCollisionRatherThanSuffixing(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{allowProvisioning: true})

	createTestUser(t, "dave")

	// Username linking is OFF, so this identity cannot attach to the existing
	// "dave". Provisioning must refuse rather than quietly create "dave-2", which
	// would look like data loss to the user.
	_, err := resolveUser(provider, claims("dave-subject", "dave"))
	if !errors.Is(err, ErrAccountExists) {
		t.Fatalf("expected ErrAccountExists, got %v", err)
	}

	var count int64
	if err := repositories.GetDB().Model(&models.User{}).Count(&count).Error; err != nil {
		t.Fatalf("failed to count users: %v", err)
	}

	if count != 1 {
		t.Errorf("expected no second account to be created, found %d users", count)
	}
}

func TestResolveUserSkipsDummyUsersOnUsernameMatch(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{linkByUsername: true})

	dummy := models.User{Username: "placeholder", DisplayName: "placeholder", Password: "x", IsDummyUser: true}
	if err := repositories.GetDB().Create(&dummy).Error; err != nil {
		t.Fatalf("failed to create the dummy user: %v", err)
	}

	// A dummy user is blocked from logging in anyway, so attaching an identity to
	// one would only produce a confusing failure later.
	_, err := resolveUser(provider, claims("placeholder-subject", "placeholder"))
	if !errors.Is(err, ErrNoAccount) {
		t.Errorf("expected ErrNoAccount for a dummy-user match, got %v", err)
	}
}

func TestResolveUserRefusesDummyUserOnReturningLogin(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	user := createTestUser(t, "willbecomedummy")
	if err := createIdentity(provider, user.ID, claims("dummy-subject", "willbecomedummy"), false); err != nil {
		t.Fatalf("failed to seed the identity: %v", err)
	}

	if err := repositories.GetDB().Model(&models.User{}).Where("id = ?", user.ID).Update("is_dummy_user", true).Error; err != nil {
		t.Fatalf("failed to flip the dummy flag: %v", err)
	}

	_, err := resolveUser(provider, claims("dummy-subject", "willbecomedummy"))
	if !errors.Is(err, ErrUserIsDummy) {
		t.Errorf("expected ErrUserIsDummy, got %v", err)
	}
}

func TestResolveLinkAttachesWithoutMatchingOrProvisioning(t *testing.T) {
	defer teardownOidcTest()
	// Both toggles OFF: linking works anyway, because the session already proved
	// who the caller is. This is what makes linkByUsername safe to default off.
	_, provider := setupOidcTest(t, oidcTestOptions{})

	user := createTestUser(t, "erin")

	err := resolveLink(provider, claims("erin-subject", "somethingelse"), user.ID)
	if err != nil {
		t.Fatalf("expected linking to succeed, got %v", err)
	}

	identity, err := repositories.NewOidcIdentityRepository(nil).GetIdentityBySubject(provider.ID, "erin-subject")
	if err != nil {
		t.Fatalf("expected an identity: %v", err)
	}

	if identity.UserId != user.ID {
		t.Errorf("expected the identity on user %d, got %d", user.ID, identity.UserId)
	}
}

func TestResolveLinkRefusesAnIdentityOwnedByAnotherUser(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	owner := createTestUser(t, "owner")
	intruder := createTestUser(t, "intruder")

	if err := createIdentity(provider, owner.ID, claims("shared-subject", "owner"), false); err != nil {
		t.Fatalf("failed to seed the identity: %v", err)
	}

	// Re-pointing would silently transfer the account.
	err := resolveLink(provider, claims("shared-subject", "owner"), intruder.ID)
	if !errors.Is(err, ErrIdentityLinkedElsewhere) {
		t.Errorf("expected ErrIdentityLinkedElsewhere, got %v", err)
	}
}

func TestResolveLinkRefusesASecondIdentityAtTheSameProvider(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	user := createTestUser(t, "frank")
	if err := createIdentity(provider, user.ID, claims("frank-subject", "frank"), false); err != nil {
		t.Fatalf("failed to seed the identity: %v", err)
	}

	err := resolveLink(provider, claims("frank-other-subject", "frank"), user.ID)
	if !errors.Is(err, ErrAlreadyLinked) {
		t.Errorf("expected ErrAlreadyLinked, got %v", err)
	}
}

func TestUnlinkRefusesToStrandAProvisionedAccount(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{allowProvisioning: true})

	user, err := resolveUser(provider, claims("stranded-subject", "stranded"))
	if err != nil {
		t.Fatalf("expected provisioning to succeed, got %v", err)
	}

	// This account has only a random, discarded password, so removing its last
	// identity would leave it with no way in at all.
	err = UnlinkIdentity(user.ID, provider.Name)
	if !errors.Is(err, ErrWouldLockOut) {
		t.Errorf("expected ErrWouldLockOut, got %v", err)
	}

	if _, err := repositories.NewOidcIdentityRepository(nil).GetIdentityForUser(provider.ID, user.ID); err != nil {
		t.Errorf("the identity should still exist after a refused unlink: %v", err)
	}
}

func TestUnlinkSucceedsForAnAccountThatHasAPassword(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	user := createTestUser(t, "grace")
	if err := createIdentity(provider, user.ID, claims("grace-subject", "grace"), false); err != nil {
		t.Fatalf("failed to seed the identity: %v", err)
	}

	if err := UnlinkIdentity(user.ID, provider.Name); err != nil {
		t.Fatalf("expected the unlink to succeed, got %v", err)
	}

	_, err := repositories.NewOidcIdentityRepository(nil).GetIdentityForUser(provider.ID, user.ID)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("expected the identity to be gone, got %v", err)
	}
}

func TestDeriveUsernameFallsBackThroughTheClaims(t *testing.T) {
	provider := models.OidcProvider{Name: "acme"}

	tests := []struct {
		name     string
		claims   idTokenClaims
		expected string
	}{
		{"prefers preferred_username", idTokenClaims{Subject: "s", PreferredUsername: "Hank", Email: "other@x.com"}, "hank"},
		{"falls back to the email local part", idTokenClaims{Subject: "s", Email: "Ivy.Jones@x.com"}, "ivy.jones"},
		{"falls back to the subject", idTokenClaims{Subject: "abc123"}, "acme-abc123"},
		{"strips characters a username may not contain", idTokenClaims{Subject: "s", PreferredUsername: "a b/c!d"}, "abcd"},
		{"skips a candidate that sanitizes too short", idTokenClaims{Subject: "s", PreferredUsername: "a", Email: "longenough@x.com"}, "longenough"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := deriveUsername(provider, test.claims); actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

// TestLinkByUsernameRequiresAnExactUsernameMatch pins the match as byte-for-byte
// in Go rather than whatever the column's collation happens to be.
//
// GetUserByUsername compares in SQL, and users.username pins no collation, so
// MySQL and MariaDB -- both supported engines -- default to a case-insensitive
// one where an identity provider account named "ADMIN" resolves the local
// "admin". SQLite (this suite) and Postgres compare case-sensitively, so without
// the Go-side re-assertion this test passes here and the hole exists only in
// production on MySQL.
func TestLinkByUsernameRequiresAnExactUsernameMatch(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{linkByUsername: true})

	existing := createTestUser(t, "admin")

	// Provisioning is off, so a refused match has nowhere to fall through to and
	// surfaces as ErrNoAccount -- which is what an unmatched claim should do.
	_, err := resolveUser(provider, claims("attacker-subject", "ADMIN"))
	if !errors.Is(err, ErrNoAccount) {
		t.Fatalf("expected a case-differing username to be refused, got %v", err)
	}

	_, err = repositories.NewOidcIdentityRepository(nil).GetIdentityBySubject(provider.ID, "attacker-subject")
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("no identity may be linked on a refused match, got %v", err)
	}

	// The exact claim still links, so the check narrows rather than breaks it.
	user, err := resolveUser(provider, claims("owner-subject", "admin"))
	if err != nil {
		t.Fatalf("expected the exact username to link, got %v", err)
	}

	if user.ID != existing.ID {
		t.Errorf("expected to land on %d, got %d", existing.ID, user.ID)
	}
}

// TestUsernameMatchesClaim is the decisive comparison, tested directly.
//
// It has to be tested directly: this suite runs on SQLite, whose `=` is
// case-sensitive, so the lookup in linkByUsername never returns a case-differing
// row here and an assertion driven through resolveUser would pass whether or not
// the rule exists. The engine that actually needs it, MySQL, is not exercised by
// the suite at all.
func TestUsernameMatchesClaim(t *testing.T) {
	tests := []struct {
		name   string
		stored string
		claim  string
		want   bool
	}{
		{"identical", "admin", "admin", true},
		{"claim upper-cased -- the MySQL collation hazard", "admin", "ADMIN", false},
		{"stored upper-cased", "ADMIN", "admin", false},
		{"mixed case", "Admin", "aDmIn", false},
		{"leading whitespace", "admin", " admin", false},
		{"empty against empty", "", "", true},
		{"unicode case fold is not equality", "\u00e5ngstr\u00f6m", "\u00c5NGSTR\u00d6M", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := usernameMatchesClaim(test.stored, test.claim); got != test.want {
				t.Errorf("usernameMatchesClaim(%q, %q) = %v, want %v", test.stored, test.claim, got, test.want)
			}
		})
	}
}

// TestLinkByUsernameInexactMatchRefusesRatherThanDuplicating covers the end of
// the road for a case-differing claim when provisioning is ON.
//
// The claim does not link (that is the point of the exactness rule), and it does
// not quietly create a lookalike account either: deriveUsername lower-cases the
// candidate, so it collides with the existing row and provisioning refuses with
// ErrAccountExists. The user is told to sign in and connect from their profile,
// which is the one path that proves who they are.
func TestLinkByUsernameInexactMatchRefusesRatherThanDuplicating(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{linkByUsername: true, allowProvisioning: true})

	existing := createTestUser(t, "dana")

	_, err := resolveUser(provider, claims("dana-upper-subject", "DANA"))
	if !errors.Is(err, ErrAccountExists) {
		t.Fatalf("expected ErrAccountExists, got %v", err)
	}

	_, err = repositories.NewOidcIdentityRepository(nil).GetIdentityBySubject(provider.ID, "dana-upper-subject")
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("no identity may be linked on a refused match, got %v", err)
	}

	var userCount int64
	if err := repositories.GetDB().Model(&models.User{}).Count(&userCount).Error; err != nil {
		t.Fatalf("failed to count users: %v", err)
	}

	if userCount != 1 {
		t.Errorf("expected only the original account to exist, got %d users", userCount)
	}

	var stored models.User
	if err := repositories.GetDB().Where("id = ?", existing.ID).First(&stored).Error; err != nil {
		t.Fatalf("failed to reload the existing user: %v", err)
	}

	if stored.Username != "dana" {
		t.Errorf("the existing account must be untouched, got username %q", stored.Username)
	}
}

// TestUnlinkRefusesEvenAfterTheProvisioningIdentityIsGone is the regression test
// for the lockout finding, and the shape no existing test covered: every other
// unlink case has exactly ONE identity, which is the one arrangement where the
// buggy guard happened to behave correctly.
//
// The old guard asked the row being removed whether it had provisioned the
// account. With a provisioned first provider and a second linked later, dropping
// the first passed (two identities remained) and dropping the second passed too
// (that row was not the provisioning one) -- leaving an account with no
// identities and a bcrypt hash of a password nobody ever knew.
func TestUnlinkRefusesEvenAfterTheProvisioningIdentityIsGone(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{allowProvisioning: true})

	second := createTestProvider(t, idp.issuer(), providerOptions{
		name:     "secondidp",
		clientId: "test-client-id",
	})

	user, err := resolveUser(provider, claims("two-provider-subject", "twoprovider"))
	if err != nil {
		t.Fatalf("expected provisioning to succeed, got %v", err)
	}

	if err := resolveLink(second, claims("two-provider-second", "twoprovider"), user.ID); err != nil {
		t.Fatalf("expected the second provider to link, got %v", err)
	}

	// Dropping the provisioning identity is fine -- one way in remains.
	if err := UnlinkIdentity(user.ID, provider.Name); err != nil {
		t.Fatalf("expected the first unlink to succeed, got %v", err)
	}

	// Dropping the last one is not, and this is what used to be allowed.
	err = UnlinkIdentity(user.ID, second.Name)
	if !errors.Is(err, ErrWouldLockOut) {
		t.Fatalf("expected ErrWouldLockOut on the last identity, got %v", err)
	}

	count, err := repositories.NewOidcIdentityRepository(nil).CountIdentitiesForUser(user.ID)
	if err != nil {
		t.Fatalf("failed to count identities: %v", err)
	}

	if count != 1 {
		t.Errorf("expected the account to keep its last way in, got %d identities", count)
	}
}

// TestLinkingToAPasswordlessAccountInheritsTheFlag pins the invariant that makes
// the guard above work: the fact lives on every identity, so it survives the
// deletion of the one that recorded it first.
func TestLinkingToAPasswordlessAccountInheritsTheFlag(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{allowProvisioning: true})

	second := createTestProvider(t, idp.issuer(), providerOptions{
		name:     "inheritidp",
		clientId: "test-client-id",
	})

	user, err := resolveUser(provider, claims("inherit-subject", "inherit"))
	if err != nil {
		t.Fatalf("expected provisioning to succeed, got %v", err)
	}

	if err := resolveLink(second, claims("inherit-second", "inherit"), user.ID); err != nil {
		t.Fatalf("expected the second provider to link, got %v", err)
	}

	identity, err := repositories.NewOidcIdentityRepository(nil).GetIdentityForUser(second.ID, user.ID)
	if err != nil {
		t.Fatalf("failed to load the linked identity: %v", err)
	}

	if !identity.ProvisionedUser {
		t.Error("an identity linked to a passwordless account must inherit the flag")
	}
}

// TestLinkingToAnAccountWithAPasswordDoesNotSetTheFlag is the other half: the
// invariant must not leak onto accounts that were never passwordless, or every
// ordinary user would be refused their last unlink.
func TestLinkingToAnAccountWithAPasswordDoesNotSetTheFlag(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	user := createTestUser(t, "haspassword")

	if err := resolveLink(provider, claims("haspassword-subject", "haspassword"), user.ID); err != nil {
		t.Fatalf("expected the link to succeed, got %v", err)
	}

	identity, err := repositories.NewOidcIdentityRepository(nil).GetIdentityForUser(provider.ID, user.ID)
	if err != nil {
		t.Fatalf("failed to load the linked identity: %v", err)
	}

	if identity.ProvisionedUser {
		t.Error("an account with its own password must not be marked passwordless")
	}

	if err := UnlinkIdentity(user.ID, provider.Name); err != nil {
		t.Errorf("expected the unlink to be allowed, got %v", err)
	}
}

// TestClearingTheProvisionedFlagUnblocksTheUnlink covers the remedy the lockout
// error actually promises -- "ask an administrator to set one before
// disconnecting". Before this, an administrator could comply and the unlink was
// still refused, because nothing observed the reset.
func TestClearingTheProvisionedFlagUnblocksTheUnlink(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{allowProvisioning: true})

	user, err := resolveUser(provider, claims("remedy-subject", "remedy"))
	if err != nil {
		t.Fatalf("expected provisioning to succeed, got %v", err)
	}

	if err := UnlinkIdentity(user.ID, provider.Name); !errors.Is(err, ErrWouldLockOut) {
		t.Fatalf("expected ErrWouldLockOut before a password is set, got %v", err)
	}

	if err := repositories.NewOidcIdentityRepository(nil).ClearProvisionedFlagForUser(user.ID); err != nil {
		t.Fatalf("failed to clear the provisioned flag: %v", err)
	}

	if err := UnlinkIdentity(user.ID, provider.Name); err != nil {
		t.Errorf("expected the unlink to be allowed once a password exists, got %v", err)
	}
}

// TestUnlinkRepairsALegacyAccountThatPredatesTheInvariant covers the database
// that already exists.
//
// The flag is now held on every identity of a passwordless account, but rows
// written before that was true carry it on only ONE -- the identity that
// provisioned the account. Deleting that row first would take the fact with it
// and the next unlink would strand the account, which is the original bug in a
// different disguise. Unlinking re-asserts the flag across what remains, so such
// an account heals the first time it is touched.
func TestUnlinkRepairsALegacyAccountThatPredatesTheInvariant(t *testing.T) {
	defer teardownOidcTest()
	idp, provider := setupOidcTest(t, oidcTestOptions{allowProvisioning: true})

	second := createTestProvider(t, idp.issuer(), providerOptions{
		name:     "legacyidp",
		clientId: "test-client-id",
	})

	user, err := resolveUser(provider, claims("legacy-subject", "legacy"))
	if err != nil {
		t.Fatalf("expected provisioning to succeed, got %v", err)
	}

	// The legacy shape: a second identity linked before resolveLink inherited the
	// flag, so it sits at false while the account still has no password.
	if err := createIdentity(second, user.ID, claims("legacy-second", "legacy"), false); err != nil {
		t.Fatalf("failed to seed the legacy identity: %v", err)
	}

	if err := UnlinkIdentity(user.ID, provider.Name); err != nil {
		t.Fatalf("expected the first unlink to succeed, got %v", err)
	}

	err = UnlinkIdentity(user.ID, second.Name)
	if !errors.Is(err, ErrWouldLockOut) {
		t.Fatalf("expected ErrWouldLockOut on the last identity, got %v", err)
	}
}
