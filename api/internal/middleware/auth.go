package middleware

import (
	"context"
	"errors"
	"net/http"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strings"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

// isRefreshToken reports whether the validated JWT claims belong to a refresh
// token (which must never be accepted as an API access credential). A shape it
// does not recognize is treated as NOT a refresh token, leaving the normal
// downstream validation in charge.
func isRefreshToken(validated interface{}) bool {
	validatedClaims, ok := validated.(*validator.ValidatedClaims)
	if !ok {
		return false
	}
	claims, ok := validatedClaims.CustomClaims.(*structs.Claims)
	if !ok {
		return false
	}
	return claims.IsRefreshToken()
}

// setAuthSource records how this request authenticated, on the claims the
// handlers will read. This middleware is the only thing that knows which branch
// ran, so it is the only thing that may set it -- and the field is json:"-", so
// nothing a caller sends can reach it.
//
// A shape it does not recognize is left unmarked, which denies: structs.Claims'
// zero AuthSource matches neither source, and the checks are written as "require
// AuthSourceJwt" rather than "reject AuthSourceApiKey" for exactly that reason.
func setAuthSource(validated interface{}, source string) {
	validatedClaims, ok := validated.(*validator.ValidatedClaims)
	if !ok {
		return
	}

	claims, ok := validatedClaims.CustomClaims.(*structs.Claims)
	if !ok {
		return
	}

	claims.AuthSource = source
}

func UnifiedAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const unauthorized = "Unauthorized"

		apiKey := getApiKey(*r)
		jwt := getJwt(*r)

		if len(apiKey) != 0 {
			dbApiKey, err := validateApiKey(apiKey)
			if err != nil {
				logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
				utils.WriteCustomErrorResponse(w, unauthorized, http.StatusForbidden)
				return
			}

			apiKeyService := services.NewApiKeyService(nil)
			claims, err := apiKeyService.GetClaimsFromApiKey(dbApiKey)
			if err != nil {
				logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
				utils.WriteCustomErrorResponse(w, unauthorized, http.StatusForbidden)
				return
			}

			go func() {
				err := apiKeyService.UpdateApiKeyLastUsedDate(dbApiKey.ID)
				if err != nil {
					logging.LogStd(logging.LOG_LEVEL_ERROR, "Failed to update API key last used date: "+err.Error())
				}
			}()

			// The scope gate. It sits AFTER the last-used stamp on purpose: that
			// write is authentication bookkeeping, not the caller's operation, so a
			// read-only key must still be able to record that it was used. Deriving
			// the predicate from "does this request write anything" would get that
			// backwards.
			if !apiKeyScopePermits(dbApiKey.Scope, classifyRequest(r)) {
				logging.LogStd(
					logging.LOG_LEVEL_ERROR,
					"API key with scope \""+dbApiKey.Scope+"\" refused for "+r.Method+" "+r.URL.Path,
				)
				utils.WriteCustomErrorResponse(w, "API key scope does not permit this operation", http.StatusForbidden)
				return
			}

			setAuthSource(&claims, structs.AuthSourceApiKey)

			r = r.Clone(context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &claims))
			next.ServeHTTP(w, r)
			return
		} else if len(jwt) != 0 {
			claims, err := validateJwt(jwt)
			if err != nil {
				logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
				utils.WriteCustomErrorResponse(w, unauthorized, http.StatusForbidden)
				return
			}

			// A refresh token is signed with the same key/issuer/audience as an
			// access token, so it validates here — but it must NOT be usable as an
			// API access credential. Reject it explicitly. (Legacy access tokens,
			// which carry no type and no jti, are still accepted; see
			// structs.Claims.IsRefreshToken.)
			if isRefreshToken(claims) {
				logging.LogStd(logging.LOG_LEVEL_ERROR, "refresh token presented as an access token")
				utils.WriteCustomErrorResponse(w, unauthorized, http.StatusForbidden)
				return
			}

			setAuthSource(claims, structs.AuthSourceJwt)

			r = r.Clone(context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, claims))
			next.ServeHTTP(w, r)
			return
		}

		utils.WriteCustomErrorResponse(w, "Unauthorized", http.StatusForbidden)
		return
	})
}

func validateJwt(jwt string) (interface{}, error) {
	validator, err := services.InitTokenValidator()
	if err != nil {
		return nil, err
	}

	return validator.ValidateToken(context.Background(), jwt)
}

func validateApiKey(apiKey string) (models.ApiKey, error) {
	apiKeyService := services.NewApiKeyService(nil)
	return apiKeyService.ValidateV1ApiKey(apiKey)
}

func getApiKey(r http.Request) string {
	authHeader := r.Header.Get("Authorization")
	parts := strings.Split(authHeader, ".")

	if len(parts) != 4 {
		return ""
	}

	if parts[0] == constants.V1Prefix {
		return authHeader
	}

	return ""
}

func getJwt(r http.Request) string {
	authCookie, err := r.Cookie(constants.JwtKey)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
	}

	if authCookie != nil && len(authCookie.Value) != 0 {
		return authCookie.Value
	}

	authHeader := r.Header.Get("Authorization")
	if len(authHeader) == 0 {
		return ""
	}

	if strings.HasPrefix(authHeader, "Bearer") {
		split := strings.Split(authHeader, "Bearer")
		if len(split) == 2 {
			authHeader = strings.TrimSpace(split[1])
		} else {
			logging.LogStd(logging.LOG_LEVEL_ERROR, errors.New("malformed auth header found"+authHeader))
		}
	}

	return authHeader
}
