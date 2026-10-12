package repositories

import (
	"testing"

	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
)

func oidcProviderCommand(name string, enabled bool) commands.UpsertOidcProviderCommand {
	secret := "a-client-secret"

	return commands.UpsertOidcProviderCommand{
		Name:              name,
		DisplayName:       name,
		IssuerUrl:         "https://accounts.example.com",
		ClientId:          "a-client-id",
		ClientSecret:      &secret,
		Scope:             "openid profile email",
		AllowProvisioning: true,
		LinkByUsername:    true,
		Enabled:           enabled,
	}
}

// TestCreateOidcProviderPersistsADisabledProvider is a regression guard.
//
// GORM skips zero-value fields on Create when the column carries a `default`
// tag, so `enabled: false` against a `default:true` column silently came back
// enabled -- meaning an administrator who registered a provider without enabling
// it would have published a live login button.
func TestCreateOidcProviderPersistsADisabledProvider(t *testing.T) {
	defer TruncateTestDb()
	t.Setenv("ENCRYPTION_KEY", "oidc-repo-test-key")

	repository := NewOidcProviderRepository(nil)

	provider, err := repository.CreateOidcProvider(oidcProviderCommand("disabled", false), nil)
	if err != nil {
		t.Fatalf("failed to create the provider: %v", err)
	}

	if provider.Enabled {
		t.Error("a provider created with enabled=false must not come back enabled")
	}

	var stored models.OidcProvider
	if err := GetDB().Where("id = ?", provider.ID).First(&stored).Error; err != nil {
		t.Fatalf("failed to reload the provider: %v", err)
	}

	if stored.Enabled {
		t.Error("the disabled flag was not persisted")
	}
}

func TestGetEnabledOidcProviderByNameIgnoresDisabledOnes(t *testing.T) {
	defer TruncateTestDb()
	t.Setenv("ENCRYPTION_KEY", "oidc-repo-test-key")

	repository := NewOidcProviderRepository(nil)

	if _, err := repository.CreateOidcProvider(oidcProviderCommand("off", false), nil); err != nil {
		t.Fatalf("failed to create the provider: %v", err)
	}

	// A disabled provider must be indistinguishable from a missing one, so turning
	// one off closes every route it serves at once.
	if _, err := repository.GetEnabledOidcProviderByName("off"); err == nil {
		t.Error("a disabled provider must not resolve")
	}

	if _, err := repository.GetOidcProviderByName("off"); err != nil {
		t.Errorf("the administrative lookup should still find it: %v", err)
	}
}

// TestDeleteOidcProviderCascades pins the explicit cascade. SQLite does not
// enforce foreign keys unless the pragma is set per connection, so an orphaned
// identity row would otherwise survive and could be re-adopted if the slug were
// reused.
func TestDeleteOidcProviderCascades(t *testing.T) {
	defer TruncateTestDb()
	t.Setenv("ENCRYPTION_KEY", "oidc-repo-test-key")

	repository := NewOidcProviderRepository(nil)

	provider, err := repository.CreateOidcProvider(oidcProviderCommand("cascade", true), nil)
	if err != nil {
		t.Fatalf("failed to create the provider: %v", err)
	}

	user := models.User{Username: "cascadeuser", DisplayName: "Cascade", Password: "x"}
	if err := GetDB().Create(&user).Error; err != nil {
		t.Fatalf("failed to create the user: %v", err)
	}

	identity := models.OidcIdentity{OidcProviderId: provider.ID, Subject: "s", UserId: user.ID}
	if err := GetDB().Create(&identity).Error; err != nil {
		t.Fatalf("failed to create the identity: %v", err)
	}

	session := models.OidcAuthSession{
		OidcProviderId: provider.ID,
		StateHash:      "state-hash",
		NonceHash:      "nonce-hash",
		CodeVerifier:   "verifier",
	}
	if err := GetDB().Create(&session).Error; err != nil {
		t.Fatalf("failed to create the session: %v", err)
	}

	if err := repository.DeleteOidcProvider(provider.ID); err != nil {
		t.Fatalf("failed to delete the provider: %v", err)
	}

	var identityCount int64
	if err := GetDB().Model(&models.OidcIdentity{}).Count(&identityCount).Error; err != nil {
		t.Fatalf("failed to count identities: %v", err)
	}

	if identityCount != 0 {
		t.Errorf("expected the identities to cascade, found %d", identityCount)
	}

	var sessionCount int64
	if err := GetDB().Model(&models.OidcAuthSession{}).Count(&sessionCount).Error; err != nil {
		t.Fatalf("failed to count sessions: %v", err)
	}

	if sessionCount != 0 {
		t.Errorf("expected the sessions to cascade, found %d", sessionCount)
	}

	// The user is a separate entity and must survive.
	var userCount int64
	if err := GetDB().Model(&models.User{}).Count(&userCount).Error; err != nil {
		t.Fatalf("failed to count users: %v", err)
	}

	if userCount != 1 {
		t.Errorf("deleting a provider must not delete users, found %d", userCount)
	}
}

// TestOidcIdentityUniqueIndexes pins the two invariants the whole feature rests
// on: one IdP account can never be claimed by two users, and one user can never
// accumulate two identities at the same provider.
func TestOidcIdentityUniqueIndexes(t *testing.T) {
	defer TruncateTestDb()
	t.Setenv("ENCRYPTION_KEY", "oidc-repo-test-key")

	provider, err := NewOidcProviderRepository(nil).CreateOidcProvider(oidcProviderCommand("unique", true), nil)
	if err != nil {
		t.Fatalf("failed to create the provider: %v", err)
	}

	first := models.User{Username: "first", DisplayName: "First", Password: "x"}
	second := models.User{Username: "second", DisplayName: "Second", Password: "x"}

	if err := GetDB().Create(&first).Error; err != nil {
		t.Fatalf("failed to create the first user: %v", err)
	}

	if err := GetDB().Create(&second).Error; err != nil {
		t.Fatalf("failed to create the second user: %v", err)
	}

	if err := GetDB().Create(&models.OidcIdentity{OidcProviderId: provider.ID, Subject: "shared", UserId: first.ID}).Error; err != nil {
		t.Fatalf("failed to create the first identity: %v", err)
	}

	if err := GetDB().Create(&models.OidcIdentity{OidcProviderId: provider.ID, Subject: "shared", UserId: second.ID}).Error; err == nil {
		t.Error("two users must not be able to claim the same provider subject")
	}

	if err := GetDB().Create(&models.OidcIdentity{OidcProviderId: provider.ID, Subject: "another", UserId: first.ID}).Error; err == nil {
		t.Error("one user must not hold two identities at the same provider")
	}
}

func TestValidateOrderByRejectsUnknownColumns(t *testing.T) {
	repository := NewOidcProviderRepository(nil)

	// The column is interpolated into the query, so the allow-list is the guard.
	if err := repository.validateOrderBy("name"); err != nil {
		t.Errorf("expected name to be allowed: %v", err)
	}

	if err := repository.validateOrderBy("client_secret"); err == nil {
		t.Error("expected an unlisted column to be rejected")
	}

	if err := repository.validateOrderBy("name; drop table users"); err == nil {
		t.Error("expected an injection attempt to be rejected")
	}
}
