package middleware

import (
	"net/http"
	"receipt-wrangler/api/internal/constants"
	"strings"
)

func UnifiedAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if ApiKeyExists(*r) {
			// validate api key
		} else if JwtExists(*r) {
			// validate jwt
		}

		next.ServeHTTP(w, r)
	})
}

func ApiKeyExists(r http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	parts := strings.Split(authHeader, ".")

	if len(parts) != 4 {
		return false
	}

	return parts[0] == constants.V1Prefix
}

func JwtExists(r http.Request) bool {
	bearerExists := strings.Contains(r.Header.Get("Authorization"), "Bearer")
	authCookieExists, err := r.Cookie(constants.JwtKey)

	return bearerExists || (authCookieExists != nil && err == nil)
}
