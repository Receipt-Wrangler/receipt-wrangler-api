package middleware

import (
	"net/http"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	"github.com/go-chi/chi/v5"
)

func ValidateGroupRole(role models.GroupRole) (mw func(http.Handler) http.Handler) {

	mw = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			var groupId string
			if len(chi.URLParam(r, "groupId")) > 0 {
				groupId = chi.URLParam(r, "groupId")
			} else {
				groupId = r.Context().Value("groupId").(string)
			}
			if groupId == "all" {
				h.ServeHTTP(w, r)
				return
			}
			errMsg := "Unauthorized access to entity."

			if len(groupId) > 0 {
				groupService := services.NewGroupService(nil)
				token := structs.GetClaims(r)
				err := groupService.ValidateGroupRole(role, groupId, utils.UintToString(token.UserId))

				if err != nil {
					logging.LogStd(logging.LOG_LEVEL_ERROR, "Unauthorized request", r)
					utils.WriteCustomErrorResponse(w, errMsg, http.StatusForbidden)
					return
				}
			}
			h.ServeHTTP(w, r)
		})
	}
	return
}

func CanDeleteGroup(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := structs.GetClaims(r)
		errMsg := "User must be a part of at least one group."

		groupMemberRepository := repositories.NewGroupMemberRepository(nil)
		groupMembers, err := groupMemberRepository.GetGroupMembersByUserId(utils.UintToString(token.UserId))
		if err != nil {
			logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
			utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
			return
		}

		if len(groupMembers) <= 1 {
			logging.LogStd(logging.LOG_LEVEL_ERROR, errMsg, r)
			utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r)
	})
}
