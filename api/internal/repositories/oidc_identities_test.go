package repositories

import (
	"receipt-wrangler/api/internal/models"
	"testing"
)

// IsAccountPasswordless is the one definition the unlink guard and the password
// login refusal both ask, so it is pinned here on its own.
func TestIsAccountPasswordless(t *testing.T) {
	defer TruncateTestDb()
	db := GetDB()
	repository := NewOidcIdentityRepository(nil)

	user := models.User{Username: "passwordless", DisplayName: "Passwordless", Password: "x"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create the user: %v", err)
	}

	passwordless, err := repository.IsAccountPasswordless(user.ID)
	if err != nil {
		t.Fatalf("IsAccountPasswordless failed: %v", err)
	}
	if passwordless {
		t.Errorf("an account with no identities is a local account, not a passwordless one")
	}

	provider := models.OidcProvider{
		Name:         "keycloak",
		DisplayName:  "Keycloak",
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
		Subject:         "a-subject",
		UserId:          user.ID,
		ProvisionedUser: true,
	}
	if err := db.Create(&identity).Error; err != nil {
		t.Fatalf("failed to create the identity: %v", err)
	}

	passwordless, err = repository.IsAccountPasswordless(user.ID)
	if err != nil {
		t.Fatalf("IsAccountPasswordless failed: %v", err)
	}
	if !passwordless {
		t.Errorf("expected a provisioned identity to make the account passwordless")
	}

	// Setting a password clears the flag, which is what ResetPassword does.
	if err := repository.ClearProvisionedFlagForUser(user.ID); err != nil {
		t.Fatalf("failed to clear the provisioned flag: %v", err)
	}

	passwordless, err = repository.IsAccountPasswordless(user.ID)
	if err != nil {
		t.Fatalf("IsAccountPasswordless failed: %v", err)
	}
	if passwordless {
		t.Errorf("expected the account to have a password of its own once the flag is cleared")
	}
}
