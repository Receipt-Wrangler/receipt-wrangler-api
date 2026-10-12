package services

import (
	"net/url"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/repositories"
	"strings"
)

// DefaultMcpPublicUrl is the externally reachable origin used when the operator
// has not configured one in System Settings. It points at the API's local dev
// address so the OAuth/metadata/redirect URLs resolve out of the box locally.
const DefaultMcpPublicUrl = "http://localhost:8081"

// IsMcpEnabled reports whether the MCP server is currently enabled, read live
// from System Settings. The MCP/OAuth handlers are always mounted; this is the
// runtime gate that lets an admin toggle the server on and off without a
// restart (mirroring the live email-polling / task-server pattern).
func IsMcpEnabled() bool {
	settings, err := repositories.NewSystemSettingsRepository(nil).GetSystemSettings()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Failed to read MCP enabled flag from system settings: "+err.Error())
		return false
	}

	return settings.McpEnabled
}

// GetMcpPublicUrl returns the externally reachable origin (scheme + host, no
// trailing slash) configured in System Settings, used to build the OAuth
// issuer, resource, metadata, and redirect URLs advertised to MCP clients such
// as Claude. It is read live so changing the setting takes effect without a
// restart. Changing it intentionally invalidates existing connector tokens,
// because the token audience is derived from this value.
func GetMcpPublicUrl() string {
	settings, err := repositories.NewSystemSettingsRepository(nil).GetSystemSettings()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Failed to read MCP public URL from system settings: "+err.Error())
		return DefaultMcpPublicUrl
	}

	return NormalizeMcpPublicUrl(settings.McpPublicUrl)
}

// GetMcpResourceUrl is the protected resource identifier (the /mcp endpoint)
// advertised to clients, echoed back as the RFC 8707 resource indicator, and
// bound as the audience of MCP tokens.
func GetMcpResourceUrl() string {
	return GetMcpPublicUrl() + "/mcp"
}

// NormalizeMcpPublicUrl trims a configured public URL down to a bare origin
// (scheme://host). Any path, query, or fragment would corrupt the
// issuer/metadata/redirect URLs derived from it, so they are dropped. An empty
// or unparseable value falls back to DefaultMcpPublicUrl.
func NormalizeMcpPublicUrl(raw string) string {
	return normalizePublicOrigin(raw, DefaultMcpPublicUrl, "MCP public URL")
}

// normalizePublicOrigin is the shared body of NormalizeMcpPublicUrl and
// NormalizeServerPublicUrl. Both settings are operator-supplied origins that
// downstream code concatenates paths onto, so they must normalize identically —
// keeping one implementation is what stops them drifting.
func normalizePublicOrigin(raw string, fallback string, label string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) == 0 {
		return fallback
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		logging.LogStd(logging.LOG_LEVEL_ERROR, label+" must be an absolute origin like https://host (got "+raw+"); falling back to "+fallback)
		return fallback
	}

	return parsed.Scheme + "://" + parsed.Host
}
