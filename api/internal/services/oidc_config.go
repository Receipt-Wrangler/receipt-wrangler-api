package services

import (
	"net/url"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/repositories"
	"strings"
)

// DefaultServerPublicUrl is the externally reachable origin used when the
// operator has not configured one. It points at the API's local dev address so a
// locally-run identity provider resolves out of the box.
const DefaultServerPublicUrl = "http://localhost:8081"

// GetServerPublicUrl returns the externally reachable origin (scheme + host, no
// trailing slash) this API is reached at, read live from System Settings so a
// change takes effect without a restart.
//
// It is deliberately separate from GetMcpPublicUrl: that value is bound into the
// MCP token audience, so sharing one setting would mean an OIDC deployment tweak
// silently invalidated every live MCP connector token.
func GetServerPublicUrl() string {
	settings, err := repositories.NewSystemSettingsRepository(nil).GetSystemSettings()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Failed to read server public URL from system settings: "+err.Error())
		return DefaultServerPublicUrl
	}

	return NormalizeServerPublicUrl(settings.ServerPublicUrl)
}

// NormalizeServerPublicUrl trims a configured public URL down to a bare origin.
// Any path, query or fragment would corrupt the redirect URI built from it.
func NormalizeServerPublicUrl(raw string) string {
	return normalizePublicOrigin(raw, DefaultServerPublicUrl, "Server public URL")
}

// IsServerPublicUrlConfigured reports whether an operator has actually set the
// origin, as opposed to falling back to the dev default. Enabling an OIDC
// provider without it would register a redirect URI pointing at localhost.
func IsServerPublicUrlConfigured() bool {
	settings, err := repositories.NewSystemSettingsRepository(nil).GetSystemSettings()
	if err != nil {
		return false
	}

	return len(strings.TrimSpace(settings.ServerPublicUrl)) > 0
}

// BuildOidcRedirectUri is the exact redirect URI an administrator must register
// with the identity provider. It is shown read-only on the provider form so it
// can be copied verbatim -- an exact-match mismatch is rejected at the IdP, where
// this application never sees the error.
func BuildOidcRedirectUri(providerName string) string {
	return GetServerPublicUrl() + "/api/oidc/" + providerName + "/callback"
}

// BuildOidcLinkLaunchUri is the URL a mobile app opens in the external browser
// to start a "connect account" flow. Unlike the redirect URI it is never
// registered anywhere -- it is this server talking to itself, one hop earlier --
// but it is built from the same origin so the browser lands back on the host
// that will set and later check the binding cookie.
func BuildOidcLinkLaunchUri(providerName string, handle string) string {
	return GetServerPublicUrl() + "/api/oidc/link/" + providerName + "/launch?h=" + url.QueryEscape(handle)
}
