package middleware

import (
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"

	"github.com/go-chi/chi/v5"
)

func ValidateGroupRole(role models.GroupRole) (mw func(http.Handler) http.Handler) {

	mw = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			groupId := chi.URLParam(r, "groupId")
			errMsg := "Unauthorized access to receipt image."

			if len(groupId) > 0 {
				token := utils.GetJWT(r)

				groupMember, err := repositories.GetGroupMemberByUserIdAndGroupId(utils.UintToString(token.UserId), groupId)
				if err != nil {
					middleware_logger.Print(err.Error())
					utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
					return
				}

				if groupMember.GroupRole != role {
					middleware_logger.Print(err.Error())
					utils.WriteCustomErrorResponse(w, errMsg, http.StatusForbidden)
					return
				}
			}
			h.ServeHTTP(w, r)
		})
	}
	return
}
