package auth_middleware

import (
	"context"
	"net/http"
	"receipt-wrangler/api/internal/utils"
)

func ValidateRefreshToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenValidator, err := utils.InitTokenValidator()
		errMessage := "Error refreshing token"

		if err != nil {
			panic(err)
		}

		refreshTokenCookie, err := r.Cookie("refresh_token")
		if err != nil {
			utils.WriteCustomErrorResponse(w, errMessage, 500)
			return
		}

		refreshToken, err := tokenValidator.ValidateToken(context.TODO(), refreshTokenCookie.Value)
		if err != nil {
			utils.WriteCustomErrorResponse(w, errMessage, 500)
			return
		}

		ctx := context.WithValue(r.Context(), "refreshToken", refreshToken)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
