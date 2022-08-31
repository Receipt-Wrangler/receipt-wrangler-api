package auth_middleware

import (
	"context"
	"fmt"
	"net/http"
	"receipt-wrangler/api/internal/handlers/auth"
	auth_utils "receipt-wrangler/api/internal/utils/auth"
	httpUtils "receipt-wrangler/api/internal/utils/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

func ValidateRefreshToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenRefreshValidator, err := auth_utils.InitTokenRefreshValidator()
		errMessage := "Error refreshing token"

		if err != nil {
			panic(err)
		}

		jwt := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims).CustomClaims.(*auth.Claims)

		// get cookie
		refreshTokenCookie, err := r.Cookie("refresh_token")
		if err != nil {
			httpUtils.WriteCustomErrorResponse(w, errMessage, 500)
			return
		}

		refreshToken, err := tokenRefreshValidator.ValidateToken(context.TODO(), refreshTokenCookie.Value)
		if err != nil {
			httpUtils.WriteCustomErrorResponse(w, errMessage, 500)
			return
		}

		if refreshToken.(*validator.ValidatedClaims).RegisteredClaims.Subject != fmt.Sprint(jwt.UserID) {
			httpUtils.WriteCustomErrorResponse(w, errMessage, 500)
			return
		}

		next.ServeHTTP(w, r)
	})
}
