package services

import (
	"encoding/json"
	"strings"
	"testing"

	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/repositories"
)

func seedOidcProvider(t *testing.T, name string, enabled bool) {
	t.Helper()

	t.Setenv("ENCRYPTION_KEY", "feature-config-test-key")

	secret := "a-client-secret"
	command := commands.UpsertOidcProviderCommand{
		Name:         name,
		DisplayName:  strings.ToUpper(name[:1]) + name[1:],
		IssuerUrl:    "https://accounts.example.com",
		ClientId:     "a-client-id",
		ClientSecret: &secret,
		Scope:        "openid profile email",
		Enabled:      enabled,
	}

	if _, err := repositories.NewOidcProviderRepository(nil).CreateOidcProvider(command, nil); err != nil {
		t.Fatalf("failed to seed the provider: %v", err)
	}
}

// TestFeatureConfigOidcProvidersSerializeAsEmptyArray is the guard on the mobile
// contract.
//
// GET /featureConfig is fetched unauthenticated by the mobile Connect screen, and
// the generated Dart deserializer has no null guard -- so a null here would fail
// the ENTIRE payload on every already-released build and report itself as
// "Failed to connect to server". This is the same class of failure as the two
// documented production login outages.
func TestFeatureConfigOidcProvidersSerializeAsEmptyArray(t *testing.T) {
	defer repositories.TruncateTestDb()

	featureConfig, err := NewSystemSettingsService(nil).GetFeatureConfig()
	if err != nil {
		t.Fatalf("GetFeatureConfig() returned an error: %v", err)
	}

	if featureConfig.OidcProviders == nil {
		t.Fatal("oidcProviders must never be nil")
	}

	bytes, err := json.Marshal(featureConfig)
	if err != nil {
		t.Fatalf("failed to marshal the feature config: %v", err)
	}

	body := string(bytes)
	if !strings.Contains(body, `"oidcProviders":[]`) {
		t.Errorf("expected oidcProviders to serialize as [], got %s", body)
	}
}

func TestFeatureConfigListsOnlyEnabledProviders(t *testing.T) {
	defer repositories.TruncateTestDb()

	seedOidcProvider(t, "google", true)
	seedOidcProvider(t, "twitch", false)

	featureConfig, err := NewSystemSettingsService(nil).GetFeatureConfig()
	if err != nil {
		t.Fatalf("GetFeatureConfig() returned an error: %v", err)
	}

	if len(featureConfig.OidcProviders) != 1 {
		t.Fatalf("expected exactly one enabled provider, got %d", len(featureConfig.OidcProviders))
	}

	if featureConfig.OidcProviders[0].Name != "google" {
		t.Errorf("expected the enabled provider, got %q", featureConfig.OidcProviders[0].Name)
	}
}

// TestFeatureConfigNeverPublishesProviderSecrets pins the minimal-exposure rule:
// this payload is unauthenticated, so it carries only what a login screen needs
// to name a provider and hit its login URL.
func TestFeatureConfigNeverPublishesProviderSecrets(t *testing.T) {
	defer repositories.TruncateTestDb()

	seedOidcProvider(t, "google", true)

	featureConfig, err := NewSystemSettingsService(nil).GetFeatureConfig()
	if err != nil {
		t.Fatalf("GetFeatureConfig() returned an error: %v", err)
	}

	bytes, err := json.Marshal(featureConfig)
	if err != nil {
		t.Fatalf("failed to marshal the feature config: %v", err)
	}

	body := string(bytes)

	for _, forbidden := range []string{"clientSecret", "a-client-secret", "clientId", "a-client-id", "issuerUrl", "accounts.example.com"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("the public feature config leaked %q: %s", forbidden, body)
		}
	}
}

func TestBuildOidcRedirectUri(t *testing.T) {
	defer repositories.TruncateTestDb()

	// With nothing configured it falls back to the dev origin rather than
	// producing a relative or empty URI.
	if uri := BuildOidcRedirectUri("google"); uri != DefaultServerPublicUrl+"/api/oidc/google/callback" {
		t.Errorf("unexpected fallback redirect URI %q", uri)
	}

	if IsServerPublicUrlConfigured() {
		t.Error("an unset server public URL must not report as configured")
	}
}

func TestNormalizeServerPublicUrlTrimsToAnOrigin(t *testing.T) {
	tests := []struct {
		raw      string
		expected string
	}{
		{"https://receipts.example.com", "https://receipts.example.com"},
		// A path, query or fragment would corrupt the redirect URI built from it.
		{"https://receipts.example.com/some/path?a=b#c", "https://receipts.example.com"},
		{"  https://receipts.example.com  ", "https://receipts.example.com"},
		{"http://192.168.1.10:8081", "http://192.168.1.10:8081"},
		{"", DefaultServerPublicUrl},
		{"not-a-url", DefaultServerPublicUrl},
	}

	for _, test := range tests {
		t.Run(test.raw, func(t *testing.T) {
			if actual := NormalizeServerPublicUrl(test.raw); actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}
