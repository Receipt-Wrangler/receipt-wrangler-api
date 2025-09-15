package middleware

import (
	"context"
	"net/http"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/utils"
	"strings"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
)

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

	return authHeader
}
