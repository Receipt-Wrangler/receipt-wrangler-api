package middleware

import (
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/simpleutils"
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
				token := structs.GetJWT(r)

				err := services.ValidateGroupRole(role, groupId, simpleutils.UintToString(token.UserId))

				if err != nil {
					middleware_logger.Print("Unauthorized request", r)
					utils.WriteCustomErrorResponse(w, errMsg, http.StatusForbidden)
					return
				}
			}
			h.ServeHTTP(w, r)
		})
	}
	return
}

func BulkValidateGroupRole(role models.GroupRole) (mw func(http.Handler) http.Handler) {

	mw = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			errMsg := "Unauthorized access to entity."
			groupMap := models.BuildGroupMap()
			groupIds := r.Context().Value("groupIds").([]string)

			if len(groupIds) > 0 {
				token := structs.GetJWT(r)
				for i := 0; i < len(groupIds); i++ {
					groupId := groupIds[i]

					groupMemberRepository := repositories.NewGroupMemberRepository(nil)
					groupMember, err := groupMemberRepository.GetGroupMemberByUserIdAndGroupId(simpleutils.UintToString(token.UserId), groupId)
					if err != nil {
						middleware_logger.Print(err.Error())
						utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
						return
					}

					var hasAccess = groupMap[groupMember.GroupRole] >= groupMap[role]

					if !hasAccess {
						middleware_logger.Print("Unauthorized request", r)
						utils.WriteCustomErrorResponse(w, errMsg, http.StatusForbidden)
						return
					}
				}

			}
			h.ServeHTTP(w, r)
		})
	}
	return
}

func CanDeleteGroup(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := structs.GetJWT(r)
		errMsg := "User must be a part of at least one group."

		groupMemberRepository := repositories.NewGroupMemberRepository(nil)
		groupMembers, err := groupMemberRepository.GetGroupMembersByUserId(simpleutils.UintToString(token.UserId))
		if err != nil {
			middleware_logger.Print(err.Error())
			utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
			return
		}

		if len(groupMembers) <= 1 {
			middleware_logger.Print(errMsg, r)
			utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r)
	})
}
