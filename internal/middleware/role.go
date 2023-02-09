package middleware

import (
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
)

func ValidateRole(role models.UserRole) (mw func(http.Handler) http.Handler) {

	mw = func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					errMsg := "Not allowed to perform this action."
					jwt := utils.GetJWT(r)
					hasRole := role == jwt.UserRole

					if (!hasRole) {
						utils.WriteCustomErrorResponse(w, errMsg, http.StatusForbidden)
						middleware_logger.Print(errMsg, r)
						return
					}
					h.ServeHTTP(w, r)
			})
	}
	return
}