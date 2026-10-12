package handlers

import (
	"net/http"
	"strings"
	"testing"

	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
)

func storedServerPublicUrl(t *testing.T) string {
	t.Helper()

	settings, err := repositories.NewSystemSettingsRepository(nil).GetSystemSettings()
	if err != nil {
		t.Fatal(err)
	}

	return settings.ServerPublicUrl
}

func createEnabledOidcProvider(t *testing.T) {
	t.Helper()

	provider := models.OidcProvider{
		Name:         "keycloak",
		DisplayName:  "Keycloak",
		IssuerUrl:    "https://idp.example.com",
		ClientId:     "client",
		ClientSecret: "secret",
		Enabled:      true,
	}

	if err := repositories.GetDB().Create(&provider).Error; err != nil {
		t.Fatalf("failed to create the provider: %v", err)
	}
}

// TestUpdateSystemSettingsCannotClearServerPublicUrlWhileAProviderIsEnabled
// closes the cross-setting rule from the side it was open on.
//
// OidcProviderService refuses to ENABLE a provider with no server public URL,
// but it only runs on a provider save -- so the same broken configuration was
// reachable by doing the two steps in the other order. The consequence is
// invisible from here: GetServerPublicUrl falls back to http://localhost:8081,
// every provider starts sending a redirect_uri nobody registered, and the
// rejection happens at the identity provider, where this API never sees it.
func TestUpdateSystemSettingsCannotClearServerPublicUrlWhileAProviderIsEnabled(t *testing.T) {
	defer tearDownSystemSettingsTest()

	repositories.GetDB().Create(&models.SystemSettings{})
	grantAllAppPerms(t, 1)

	configured := "https://receipts.example.com"

	w := putSystemSettings(t, map[string]interface{}{"serverPublicUrl": configured})
	if w.Code != http.StatusOK {
		t.Fatalf("setting the URL: status %d, body %s", w.Code, w.Body.String())
	}

	createEnabledOidcProvider(t)

	// Blanking it now is a field-level 400 and changes nothing.
	w = putSystemSettings(t, map[string]interface{}{"serverPublicUrl": ""})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("clearing the URL: status %d, want 400 (%s)", w.Code, w.Body.String())
	}

	if !strings.Contains(w.Body.String(), "serverPublicUrl") {
		t.Errorf("the 400 should name serverPublicUrl: %s", w.Body.String())
	}

	if got := storedServerPublicUrl(t); got != configured {
		t.Errorf("a rejected save changed the stored URL to %q", got)
	}

	// Whitespace is not a loophole: the guard trims, exactly as
	// IsServerPublicUrlConfigured does, so the two halves of the rule cannot
	// disagree about what "configured" means.
	w = putSystemSettings(t, map[string]interface{}{"serverPublicUrl": "   "})
	if w.Code != http.StatusBadRequest {
		t.Errorf("clearing with whitespace: status %d, want 400 (%s)", w.Code, w.Body.String())
	}

	// CHANGING it stays allowed -- a deployment legitimately moves, and the admin
	// re-registers the redirect URI at the identity provider.
	w = putSystemSettings(t, map[string]interface{}{"serverPublicUrl": "https://moved.example.com"})
	if w.Code != http.StatusOK {
		t.Fatalf("changing the URL: status %d, body %s", w.Code, w.Body.String())
	}

	if got := storedServerPublicUrl(t); got != "https://moved.example.com" {
		t.Errorf("stored = %q, want the new URL", got)
	}

	// And omitting the key is untouched, which is what keeps every settings save
	// that has nothing to do with OIDC working.
	w = putSystemSettings(t, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("omitted key: status %d, body %s", w.Code, w.Body.String())
	}

	if got := storedServerPublicUrl(t); got != "https://moved.example.com" {
		t.Errorf("an omitted key changed the stored URL to %q", got)
	}
}

// TestUpdateSystemSettingsClearsServerPublicUrlWithNoEnabledProvider is the
// control. The rule is about providers that would break, not about the setting
// being precious -- an install with none, or with only disabled ones, must still
// be able to clear it.
func TestUpdateSystemSettingsClearsServerPublicUrlWithNoEnabledProvider(t *testing.T) {
	defer tearDownSystemSettingsTest()

	repositories.GetDB().Create(&models.SystemSettings{})
	grantAllAppPerms(t, 1)

	w := putSystemSettings(t, map[string]interface{}{"serverPublicUrl": "https://receipts.example.com"})
	if w.Code != http.StatusOK {
		t.Fatalf("setting the URL: status %d, body %s", w.Code, w.Body.String())
	}

	// A DISABLED provider sends no redirect_uri, so it cannot be broken by this.
	provider := models.OidcProvider{
		Name:         "disabled-idp",
		DisplayName:  "Disabled",
		IssuerUrl:    "https://idp.example.com",
		ClientId:     "client",
		ClientSecret: "secret",
		Enabled:      false,
	}
	if err := repositories.GetDB().Create(&provider).Error; err != nil {
		t.Fatalf("failed to create the provider: %v", err)
	}

	w = putSystemSettings(t, map[string]interface{}{"serverPublicUrl": ""})
	if w.Code != http.StatusOK {
		t.Fatalf("clearing the URL: status %d, body %s", w.Code, w.Body.String())
	}

	if got := storedServerPublicUrl(t); got != "" {
		t.Errorf("expected the URL to be cleared, got %q", got)
	}
}
