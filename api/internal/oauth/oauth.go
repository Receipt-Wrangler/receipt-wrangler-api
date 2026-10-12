// Package oauth implements a minimal, self-hosted OAuth 2.1 authorization
// server scoped to the MCP integration. It lets MCP clients such as Claude
// authenticate end users against the existing Receipt Wrangler user store and
// obtain the same HS512 JWTs the REST API already trusts.
//
// The flow Claude drives:
//
//	discovery (RFC 9728 / RFC 8414) -> dynamic client registration (RFC 7591)
//	-> authorization code + PKCE S256 (login form backed by services.LoginUser)
//	-> token exchange wrapping services.GenerateJWT.
//
// The OAuth access token is the existing Receipt Wrangler access JWT, so the
// MCP auth middleware validates it with the same services.InitTokenValidator
// used everywhere else; no parallel token system is introduced.
package oauth

import (
	"encoding/json"
	"net/http"
)

const (
	// readScope is the only scope granted to MCP tokens in v1.
	//
	// NOTE: this scope is currently DEAD as an authorization control. Nothing
	// gates access on it — read-only behavior is guaranteed structurally by the
	// fact that internal/mcp only registers read tools (no write/delete tools),
	// not by checking this scope. The moment a write or delete tool is added,
	// that structural guarantee is gone and real per-tool scope enforcement
	// must be introduced here and in internal/mcp. See the matching note on
	// mcpReadScope in internal/mcp/server.go.
	readScope = "mcp:read"

	// authCodeTTLSeconds bounds how long an issued authorization code is
	// valid before it must be redeemed at the token endpoint.
	authCodeTTLSeconds = 300

	// accessTokenExpiresIn mirrors utils.GetAccessTokenExpiryDate (20 minutes)
	// and is reported to clients so they know when to refresh.
	accessTokenExpiresIn = 20 * 60
)

// writeJson writes a value as an application/json response with the given
// status code.
func writeJson(w http.ResponseWriter, status int, body interface{}) {
	bytes, err := json.Marshal(body)
	if err != nil {
		http.Error(w, "internal_error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(bytes)
}

// writeOAuthError writes an RFC 6749 §5.2 error response. error is one of the
// registered OAuth error codes (invalid_request, invalid_grant, etc.).
func writeOAuthError(w http.ResponseWriter, status int, oauthError string, description string) {
	writeJson(w, status, map[string]string{
		"error":             oauthError,
		"error_description": description,
	})
}
