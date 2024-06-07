package middleware

import (
	"net/http"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

func ValidateRole(role models.UserRole) (mw func(http.Handler) http.Handler) {

	mw = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			errMsg := "Not allowed to perform this action."
			jwt := structs.GetJWT(r)
			hasRole := models.HasRole(jwt.UserRole, role)

			if !hasRole {
				utils.WriteCustomErrorResponse(w, errMsg, http.StatusForbidden)
				logging.LogStd(logging.LOG_LEVEL_ERROR, errMsg, r)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
	return
}
