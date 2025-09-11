package middleware

import (
	"net/http"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/utils"
	"strings"
)

func UnifiedAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const unauthorized = "Unauthorized"

		apiKey := getApiKey(*r)
		jwt := getJwt(*r)

		if len(apiKey) != 0 {
			err := validateApiKey(*r)
			if err != nil {
				logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
				utils.WriteCustomErrorResponse(w, unauthorized, http.StatusForbidden)
				return
			}
		} else if len(jwt) != 0 {
			err := validateJwt(*r)
			if err != nil {
				logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
				utils.WriteCustomErrorResponse(w, unauthorized, http.StatusForbidden)
				return
			}
		}

		utils.WriteCustomErrorResponse(w, "Unauthorized", http.StatusForbidden)
		return
	})
}

func validateJwt(r http.Request) error {
	return nil
}

func validateApiKey(r http.Request) error {
	return nil
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
