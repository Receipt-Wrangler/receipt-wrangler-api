package oidc

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"testing"
	"time"

	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/utils"

	"github.com/go-chi/chi/v5"
)

type oidcTestOptions struct {
	allowProvisioning bool
	linkByUsername    bool
}

type providerOptions struct {
	name              string
	clientId          string
	allowProvisioning bool
	linkByUsername    bool
}

// setupOidcTest stands up a fake identity provider and a provider row pointed at
// it, and seeds the roles a provisioned user needs.
func setupOidcTest(t *testing.T, options oidcTestOptions) (*fakeIdp, models.OidcProvider) {
	t.Helper()

	repositories.CreateTestRoles()

	idp := newFakeIdp(t, "test-client-id")
	provider := createTestProvider(t, idp.issuer(), providerOptions{
		name:              "testidp",
		clientId:          idp.clientId,
		allowProvisioning: options.allowProvisioning,
		linkByUsername:    options.linkByUsername,
	})

	return idp, provider
}

func createTestProvider(t *testing.T, issuerUrl string, options providerOptions) models.OidcProvider {
	t.Helper()

	secret := "test-client-secret"
	command := commands.UpsertOidcProviderCommand{
		Name:              options.name,
		DisplayName:       options.name,
		IssuerUrl:         issuerUrl,
		ClientId:          options.clientId,
		ClientSecret:      &secret,
		Scope:             "openid profile email",
		AllowProvisioning: options.allowProvisioning,
		LinkByUsername:    options.linkByUsername,
		Enabled:           true,
	}

	provider, err := repositories.NewOidcProviderRepository(nil).CreateOidcProvider(command, nil)
	if err != nil {
		t.Fatalf("failed to create the test provider: %v", err)
	}

	return provider
}

// teardownOidcTest resets both the database and the discovery cache.
//
// Clearing the cache is not optional: the test database is truncated between
// cases and reuses row ids, so a cached provider would be served to a later test
// pointed at a different fake identity provider.
func teardownOidcTest() {
	ClearProviderCacheForTests()
	// Same hazard as the provider cache: the role-permission cache is keyed by
	// scope + role id, and the truncated test database reuses ids, so one case's
	// permission set would be served to the next.
	services.ClearRolePermissionCacheForTests()
	repositories.TruncateTestDb()
}

func withChiContext(r *http.Request, routeContext *chi.Context) context.Context {
	return context.WithValue(r.Context(), chi.RouteCtxKey, routeContext)
}

func timeInPast() time.Time {
	return time.Now().Add(-time.Hour)
}

// challengeFor derives the S256 challenge for a verifier independently of
// utils.VerifyPkceS256, so the exchange tests are not judged by the code they
// exercise.
func challengeFor(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func createTestUser(t *testing.T, username string) models.User {
	t.Helper()

	hashed, err := utils.HashPassword("password")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	user := models.User{
		Username:    username,
		DisplayName: username,
		Password:    string(hashed),
	}

	if err := repositories.GetDB().Create(&user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	return user
}

var _ = env.GetEncryptionKey

// createTestUserWithPermissions creates a user holding an app role that grants
// exactly the permissions given. CreateTestRoles' shared role carries only
// app.users.create, so a case that turns on a specific permission needs its own.
func createTestUserWithPermissions(t *testing.T, username string, granted ...string) models.User {
	t.Helper()

	user := createTestUser(t, username)

	rolePermissions := make([]models.AppRolePermission, 0, len(granted))
	for _, permission := range granted {
		rolePermissions = append(rolePermissions, models.AppRolePermission{Permission: permission})
	}

	role := models.AppRole{
		Name:        "role-for-" + username,
		Description: "test role",
		Permissions: rolePermissions,
	}

	if err := repositories.GetDB().Create(&role).Error; err != nil {
		t.Fatalf("failed to create the app role: %v", err)
	}

	err := repositories.GetDB().
		Model(&models.User{}).
		Where("id = ?", user.ID).
		Update("app_role_id", role.ID).Error
	if err != nil {
		t.Fatalf("failed to assign the app role: %v", err)
	}

	user.AppRoleID = &role.ID

	return user
}
