package middleware

import (
	"net/http"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/utils"
)

func MoveJWTCookieToHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errMessage := "JWT not found"

		if utils.IsMobileApp(r) {
			next.ServeHTTP(w, r)
			return
		}

		accessTokenCookie, err := r.Cookie(constants.JWT_KEY)
		if err != nil {
			utils.WriteCustomErrorResponse(w, errMessage, http.StatusBadRequest)
			middleware_logger.Println(errMessage)
			return
		}

		r.Header.Add("Authorization", "Bearer "+accessTokenCookie.Value)

		next.ServeHTTP(w, r)
	})
}
