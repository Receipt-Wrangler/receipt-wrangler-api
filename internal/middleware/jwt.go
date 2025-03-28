package middleware

import (
	"net/http"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/utils"
)

func MoveJWTCookieToHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errMessage := "JWT not found"

		if utils.IsMobileApp(r) {
			next.ServeHTTP(w, r)
			return
		}

		accessTokenCookie, err := r.Cookie(constants.JwtKey)
		if err != nil {
			utils.WriteCustomErrorResponse(w, errMessage, http.StatusForbidden)
			logging.LogStd(logging.LOG_LEVEL_ERROR, errMessage)
			return
		}

		r.Header.Add("Authorization", "Bearer "+accessTokenCookie.Value)

		next.ServeHTTP(w, r)
	})
}
