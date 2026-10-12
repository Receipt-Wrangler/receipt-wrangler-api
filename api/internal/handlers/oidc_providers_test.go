package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/permissions"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/go-chi/chi/v5"
)

const (
	oidcTestClientSecret = "super-secret-client-secret"
	oidcTestUserId       = uint(1)
)

func setupOidcProviderHandlerTest(t *testing.T) {
	t.Helper()

	// The provider's client secret is encrypted at rest, so this package needs a
	// key for these cases the way internal/ai does.
	t.Setenv("ENCRYPTION_KEY", "oidc-handler-test-key")

	// A server public URL must be configured before a provider can be enabled --
	// otherwise there is no redirect URI to register with the identity provider.
	settingsRepository := repositories.NewSystemSettingsRepository(nil)
	settings, err := settingsRepository.GetSystemSettings()
	if err != nil {
		t.Fatalf("failed to read system settings: %v", err)
	}

	if err := repositories.GetDB().
		Model(&models.SystemSettings{}).
		Where("id = ?", settings.ID).
		Update("server_public_url", "https://receipts.example.com").Error; err != nil {
		t.Fatalf("failed to set the server public URL: %v", err)
	}
}

func teardownOidcProviderHandlerTest() {
	repositories.TruncateTestDb()
}

func oidcRequest(t *testing.T, method string, url string, body string, urlParams map[string]string) (*http.Request, *httptest.ResponseRecorder) {
	t.Helper()

	request := httptest.NewRequest(method, url, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	var vClaims validator.ValidatedClaims
	vClaims.CustomClaims = &structs.Claims{
		UserId:      oidcTestUserId,
		Username:    "perm-user-1",
		Displayname: "Perm User",
	}

	ctx := context.WithValue(request.Context(), jwtmiddleware.ContextKey{}, &vClaims)

	if len(urlParams) > 0 {
		routeContext := chi.NewRouteContext()
		for key, value := range urlParams {
			routeContext.URLParams.Add(key, value)
		}
		ctx = context.WithValue(ctx, chi.RouteCtxKey, routeContext)
	}

	return request.WithContext(ctx), httptest.NewRecorder()
}

func createProviderBody(name string) string {
	command := commands.UpsertOidcProviderCommand{
		Name:              name,
		DisplayName:       "Test Provider",
		IssuerUrl:         "https://accounts.example.com",
		ClientId:          "a-client-id",
		Scope:             "openid profile email",
		AllowProvisioning: true,
		Enabled:           true,
	}

	secret := oidcTestClientSecret
	command.ClientSecret = &secret

	bytes, _ := json.Marshal(command)
	return string(bytes)
}

func TestOidcProviderEndpointsRequireTheirPermissions(t *testing.T) {
	defer teardownOidcProviderHandlerTest()
	setupOidcProviderHandlerTest(t)

	// Registering an identity provider decides who can sign in to the whole
	// install, so every one of these must be gated.
	tests := []struct {
		name       string
		permission string
		handler    http.HandlerFunc
		method     string
		url        string
		body       string
		urlParams  map[string]string
	}{
		{"create", permissions.AppOidcProvidersCreate, CreateOidcProvider, http.MethodPost, "/api/oidcProvider/", createProviderBody("gated-create"), nil},
		{"read one", permissions.AppOidcProvidersRead, GetOidcProviderById, http.MethodGet, "/api/oidcProvider/1", "", map[string]string{"oidcProviderId": "1"}},
		{"read paged", permissions.AppOidcProvidersRead, GetPagedOidcProviders, http.MethodPost, "/api/oidcProvider/getPagedOidcProviders", `{"page":1,"pageSize":10,"orderBy":"name","sortDirection":"asc"}`, nil},
		{"update", permissions.AppOidcProvidersUpdate, UpdateOidcProvider, http.MethodPut, "/api/oidcProvider/1", createProviderBody("gated-update"), map[string]string{"oidcProviderId": "1"}},
		{"delete", permissions.AppOidcProvidersDelete, DeleteOidcProvider, http.MethodDelete, "/api/oidcProvider/1", "", map[string]string{"oidcProviderId": "1"}},
	}

	// Grant an unrelated permission once, so each case proves the specific gate
	// rather than merely that the user has no role at all. Granting inside the
	// subtests would collide on the helper's per-user role name.
	grantAppPerms(t, oidcTestUserId, permissions.AppAccountRead)

	for _, test := range tests {
		t.Run(test.name+" is denied without the permission", func(t *testing.T) {
			request, recorder := oidcRequest(t, test.method, test.url, test.body, test.urlParams)
			test.handler(recorder, request)

			if recorder.Code != http.StatusForbidden {
				t.Errorf("expected 403 without %s, got %d", test.permission, recorder.Code)
			}
		})
	}
}

// TestGetOidcProviderNeverReturnsClientSecret is the leak guard. The secret lets
// anyone impersonate this deployment to the identity provider, so it must not
// appear in any response body under any key.
func TestGetOidcProviderNeverReturnsClientSecret(t *testing.T) {
	defer teardownOidcProviderHandlerTest()
	setupOidcProviderHandlerTest(t)
	grantAppPerms(t, oidcTestUserId, permissions.AppOidcProvidersCreate, permissions.AppOidcProvidersRead)

	request, recorder := oidcRequest(t, http.MethodPost, "/api/oidcProvider/", createProviderBody("secretless"), nil)
	CreateOidcProvider(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected the create to succeed, got %d (%s)", recorder.Code, recorder.Body.String())
	}

	assertNoSecret(t, recorder.Body.String(), "create response")

	var created structs.OidcProviderView
	if err := json.Unmarshal(recorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode the create response: %v", err)
	}

	if !created.HasClientSecret {
		t.Error("hasClientSecret should report that a secret is stored")
	}

	if created.RedirectUri != "https://receipts.example.com/api/oidc/secretless/callback" {
		t.Errorf("unexpected redirect URI %q", created.RedirectUri)
	}

	request, recorder = oidcRequest(t, http.MethodGet, "/api/oidcProvider/1", "", map[string]string{"oidcProviderId": "1"})
	GetOidcProviderById(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected the read to succeed, got %d (%s)", recorder.Code, recorder.Body.String())
	}

	assertNoSecret(t, recorder.Body.String(), "get response")

	request, recorder = oidcRequest(t, http.MethodPost, "/api/oidcProvider/getPagedOidcProviders", `{"page":1,"pageSize":10,"orderBy":"name","sortDirection":"asc"}`, nil)
	GetPagedOidcProviders(recorder, request)

	assertNoSecret(t, recorder.Body.String(), "paged response")
}

func assertNoSecret(t *testing.T, body string, label string) {
	t.Helper()

	if strings.Contains(body, oidcTestClientSecret) {
		t.Errorf("the %s leaked the plaintext client secret", label)
	}

	if strings.Contains(body, "clientSecret") {
		t.Errorf("the %s carries a clientSecret key", label)
	}
}

func TestClientSecretIsEncryptedAtRest(t *testing.T) {
	defer teardownOidcProviderHandlerTest()
	setupOidcProviderHandlerTest(t)
	grantAppPerms(t, oidcTestUserId, permissions.AppOidcProvidersCreate)

	request, recorder := oidcRequest(t, http.MethodPost, "/api/oidcProvider/", createProviderBody("encrypted"), nil)
	CreateOidcProvider(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected the create to succeed, got %d (%s)", recorder.Code, recorder.Body.String())
	}

	var stored models.OidcProvider
	if err := repositories.GetDB().Where("name = ?", "encrypted").First(&stored).Error; err != nil {
		t.Fatalf("failed to load the provider: %v", err)
	}

	if stored.ClientSecret == oidcTestClientSecret {
		t.Fatal("the client secret was stored in plaintext")
	}

	decrypted, err := repositories.NewOidcProviderRepository(nil).GetDecryptedClientSecret(stored)
	if err != nil {
		t.Fatalf("failed to decrypt the stored secret: %v", err)
	}

	if decrypted != oidcTestClientSecret {
		t.Errorf("the secret did not round-trip: got %q", decrypted)
	}
}

// TestUpdateClearsBooleanFlags is the GORM map-form regression guard.
//
// The struct form of Updates skips zero values, so turning any of these OFF would
// be silently ignored -- and each one is a security toggle, so an administrator
// revoking it would believe they had.
func TestUpdateClearsBooleanFlags(t *testing.T) {
	defer teardownOidcProviderHandlerTest()
	setupOidcProviderHandlerTest(t)
	grantAppPerms(t, oidcTestUserId,
		permissions.AppOidcProvidersCreate,
		permissions.AppOidcProvidersUpdate,
		permissions.AppOidcProvidersRead,
	)

	request, recorder := oidcRequest(t, http.MethodPost, "/api/oidcProvider/", createProviderBody("toggles"), nil)
	CreateOidcProvider(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected the create to succeed, got %d (%s)", recorder.Code, recorder.Body.String())
	}

	var created structs.OidcProviderView
	if err := json.Unmarshal(recorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode the create response: %v", err)
	}

	if !created.AllowProvisioning {
		t.Fatal("expected provisioning to start on")
	}

	// Now turn every toggle off, and omit the secret entirely.
	command := commands.UpsertOidcProviderCommand{
		Name:              "toggles",
		DisplayName:       "Test Provider",
		IssuerUrl:         "https://accounts.example.com",
		ClientId:          "a-client-id",
		Scope:             "openid profile email",
		AllowProvisioning: false,
		LinkByUsername:    false,
		Enabled:           false,
	}

	body, _ := json.Marshal(command)
	request, recorder = oidcRequest(t, http.MethodPut, "/api/oidcProvider/1", string(body), map[string]string{"oidcProviderId": "1"})
	UpdateOidcProvider(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected the update to succeed, got %d (%s)", recorder.Code, recorder.Body.String())
	}

	var stored models.OidcProvider
	if err := repositories.GetDB().Where("name = ?", "toggles").First(&stored).Error; err != nil {
		t.Fatalf("failed to reload the provider: %v", err)
	}

	if stored.AllowProvisioning {
		t.Error("allowProvisioning was not turned off")
	}

	if stored.Enabled {
		t.Error("enabled was not turned off")
	}

	// And omitting the secret must keep the stored ciphertext byte for byte.
	decrypted, err := repositories.NewOidcProviderRepository(nil).GetDecryptedClientSecret(stored)
	if err != nil {
		t.Fatalf("failed to decrypt after the update: %v", err)
	}

	if decrypted != oidcTestClientSecret {
		t.Errorf("an update that omitted the secret changed it: got %q", decrypted)
	}
}

func TestEnablingAProviderRequiresAServerPublicUrl(t *testing.T) {
	defer teardownOidcProviderHandlerTest()
	setupOidcProviderHandlerTest(t)
	grantAppPerms(t, oidcTestUserId, permissions.AppOidcProvidersCreate)

	// Without an origin there is no redirect URI to register, so the provider would
	// point at localhost and fail at the identity provider rather than here.
	if err := repositories.GetDB().
		Model(&models.SystemSettings{}).
		Where("id > 0").
		Update("server_public_url", "").Error; err != nil {
		t.Fatalf("failed to clear the server public URL: %v", err)
	}

	request, recorder := oidcRequest(t, http.MethodPost, "/api/oidcProvider/", createProviderBody("nourl"), nil)
	CreateOidcProvider(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected 400 without a server public URL, got %d (%s)", recorder.Code, recorder.Body.String())
	}
}
