package oauth

import (
	"context"
	"errors"
	"net/http"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	"github.com/auth0/go-jwt-middleware/v2/validator"
)

// Token handles POST /oauth/token for the authorization_code and refresh_token
// grants. Both ultimately mint a standard Receipt Wrangler access JWT via
// services.GenerateJWT, so the issued access token is validated by the same
// machinery the REST API uses.
func Token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "Malformed request body")
		return
	}

	switch r.PostForm.Get("grant_type") {
	case "authorization_code":
		handleAuthorizationCodeGrant(w, r)
	case "refresh_token":
		handleRefreshTokenGrant(w, r)
	default:
		writeOAuthError(w, http.StatusBadRequest, "unsupported_grant_type", "Unsupported grant_type")
	}
}

func handleAuthorizationCodeGrant(w http.ResponseWriter, r *http.Request) {
	code := r.PostForm.Get("code")
	clientId := r.PostForm.Get("client_id")
	redirectUri := r.PostForm.Get("redirect_uri")
	codeVerifier := r.PostForm.Get("code_verifier")

	if len(code) == 0 || len(codeVerifier) == 0 {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "code and code_verifier are required")
		return
	}

	// Load the code and validate the client binding and PKCE BEFORE consuming
	// it, so a mismatched or forged redemption attempt cannot burn a valid code
	// out from under the legitimate client.
	authCode, err := getAuthorizationCode(code)
	if err != nil {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "Invalid or expired authorization code")
		return
	}

	if authCode.ClientId != clientId || authCode.RedirectUri != redirectUri {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "client_id or redirect_uri mismatch")
		return
	}

	if !utils.VerifyPkceS256(codeVerifier, authCode.CodeChallenge) {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "PKCE verification failed")
		return
	}

	// Atomically consume the code. A losing race (or a replayed/expired code)
	// affects zero rows and is rejected here, guaranteeing single use.
	consumed, err := markAuthorizationCodeUsed(code)
	if err != nil {
		writeOAuthError(w, http.StatusInternalServerError, "server_error", "Failed to redeem authorization code")
		return
	}
	if !consumed {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "Invalid or expired authorization code")
		return
	}

	issueTokenResponse(w, authCode.UserId, authCode.Scope)
}

func handleRefreshTokenGrant(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.PostForm.Get("refresh_token")
	if len(refreshToken) == 0 {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "refresh_token is required")
		return
	}

	userId, err := rotateRefreshToken(refreshToken)
	if err != nil {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "Invalid or expired refresh token")
		return
	}

	issueTokenResponse(w, userId, readScope)
}

// rotateRefreshToken validates a refresh token JWT, then atomically verifies it
// exists in the database and is unused and marks it used, mirroring the
// existing single-use refresh flow (middleware.RevokeRefreshToken). It returns
// the owning user id on success.
func rotateRefreshToken(refreshTokenString string) (uint, error) {
	// Validate against the MCP audience, not the normal REST audience. This is
	// what stops a client from POSTing an MCP refresh token to /api/token to
	// trade it for a full-access token (and vice versa): each endpoint only
	// accepts refresh tokens minted for its own audience.
	tokenValidator, err := services.InitMcpTokenValidator(mcpResourceUrl())
	if err != nil {
		return 0, err
	}

	validated, err := tokenValidator.ValidateToken(context.Background(), refreshTokenString)
	if err != nil {
		return 0, err
	}

	claims := validated.(*validator.ValidatedClaims).CustomClaims.(*structs.Claims)

	hashedToken := utils.Sha256Hash([]byte(refreshTokenString))

	// Atomically mark the stored token used via the WHERE clause so two
	// concurrent refreshes of the same token cannot both succeed. A zero
	// row count means the token is unknown or already used.
	result := repositories.GetDB().
		Model(&models.RefreshToken{}).
		Where("token = ? AND is_used = ?", hashedToken, false).
		Update("is_used", true)
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected != 1 {
		return 0, errors.New("refresh token is invalid or already used")
	}

	return claims.UserId, nil
}

// issueTokenResponse mints a fresh access + refresh JWT pair for the user and
// writes the OAuth token response. The pair is bound to the MCP resource as its
// audience (rather than the normal REST audience) so these tokens are accepted
// only at the MCP endpoints.
func issueTokenResponse(w http.ResponseWriter, userId uint, scope string) {
	accessToken, refreshToken, _, err := services.GenerateMcpJWT(userId, mcpResourceUrl())
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Failed to generate token for MCP client: "+err.Error())
		writeOAuthError(w, http.StatusInternalServerError, "server_error", "Failed to issue token")
		return
	}

	writeJson(w, http.StatusOK, map[string]interface{}{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    accessTokenExpiresIn,
		"refresh_token": refreshToken,
		"scope":         scopeOrDefault(scope),
	})
}
