package structs

import (
	"context"
	"fmt"
	"receipt-wrangler/api/internal/models"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Token type values carried on the TokenType claim. They keep an access token
// and a refresh token — which are otherwise minted with identical claims — from
// being used interchangeably: access tokens are rejected by the refresh paths and
// refresh tokens are rejected by the API/MCP access paths.
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// How a request authenticated, recorded on Claims.AuthSource by
// middleware.UnifiedAuthMiddleware. The zero value is neither of these, so a
// check written as "require AuthSourceJwt" denies anything unmarked.
const (
	AuthSourceJwt    = "jwt"
	AuthSourceApiKey = "api_key"
)

type Claims struct {
	DefaultAvatarColor string             `json:"defaultAvatarColor"`
	Displayname        string             `json:"displayName"`
	UserId             uint               `json:"userId"`
	Username           string             `json:"username"`
	ApiKeyScope        models.ApiKeyScope `json:"apiKeyScope"`
	// TokenType is "access" or "refresh" on tokens minted after the token-type
	// fix. It is empty on legacy tokens issued before it (see IsRefreshToken).
	TokenType string `json:"tokenType"`

	// AuthSource records HOW this request authenticated: AuthSourceJwt for an
	// interactive session, AuthSourceApiKey for a key. It exists because a handful
	// of operations are legitimate for a signed-in human and an escalation for a
	// key -- connecting an identity provider above all, which turns a scoped,
	// revocable credential into a permanent login.
	//
	// json:"-" is load bearing, not tidiness. The field must never ride on a token:
	// serialized, it would be settable by whoever minted the token and inheritable
	// by any legacy one, and the thing it answers ("how did THIS request arrive")
	// is not a property of a token at all. UnifiedAuthMiddleware sets it per
	// request, and it is the only writer.
	//
	// Do NOT infer the same thing from ApiKeyScope being non-empty. It happens to
	// hold today, it is a serialized claim, and the inference fails OPEN for any
	// path that does not set it.
	AuthSource string `json:"-"`

	jwt.RegisteredClaims
}

// IsRefreshToken reports whether these claims belong to a refresh token rather
// than an access token. Tokens minted after the token-type fix carry an explicit
// TokenType. Legacy tokens carry none; a legacy refresh token is still
// distinguishable because only refresh tokens were ever given a jti
// (RegisteredClaims.ID), so a typeless token WITH a jti is a refresh token. This
// lets the access paths reject refresh tokens immediately — closing the
// refresh-token-as-access-token hole — while still honoring already-issued legacy
// access tokens (no type, no jti) until they expire, so no user is forced to log
// in again on deploy.
func (claim *Claims) IsRefreshToken() bool {
	switch claim.TokenType {
	case TokenTypeRefresh:
		return true
	case TokenTypeAccess:
		return false
	default:
		return claim.RegisteredClaims.ID != ""
	}
}

func (claim *Claims) Validate(ctx context.Context) error {
	if claim.UserId == 0 {
		return fmt.Errorf("user ID is required")
	}

	if claim.Username == "" {
		return fmt.Errorf("username is required")
	}

	if claim.Displayname == "" {
		return fmt.Errorf("display name is required")
	}

	// Validate DefaultAvatarColor format (should be hex color)
	if claim.DefaultAvatarColor != "" {
		if !strings.HasPrefix(claim.DefaultAvatarColor, "#") || len(claim.DefaultAvatarColor) != 7 {
			return fmt.Errorf("invalid avatar color format: %s", claim.DefaultAvatarColor)
		}
		// Check if it's valid hex
		if _, err := strconv.ParseInt(claim.DefaultAvatarColor[1:], 16, 64); err != nil {
			return fmt.Errorf("invalid hex color: %s", claim.DefaultAvatarColor)
		}
	}

	// Validate API key scope if present (should be valid scope values)
	if claim.ApiKeyScope != "" {
		if !claim.ApiKeyScope.IsValid() {
			return fmt.Errorf("invalid API key scope: %s", claim.ApiKeyScope)
		}
	}

	return nil
}
