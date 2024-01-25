package handlers

import (
	"errors"
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

func Login(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Invalid credentials.",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			userData := r.Context().Value("user").(commands.LoginCommand)
			clientData := make(map[string]interface{})
			var dbUser models.User

			dbUser, err := services.LoginUser(userData)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			if dbUser.IsDummyUser {
				return http.StatusInternalServerError, errors.New("dummy users cannot log in")
			}

			jwt, refreshToken, accessTokenClaims, err := services.GenerateJWT(dbUser.ID)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			services.PrepareAccessTokenClaims(accessTokenClaims)

			// Add claims data to clientData
			clientData["claims"] = accessTokenClaims

			if utils.IsMobileDevice(r) {
				clientData["jwt"] = jwt
				clientData["refreshToken"] = refreshToken
			} else {
				accessTokenCookie, refreshTokenCookie := services.BuildTokenCookies(jwt, refreshToken)

				http.SetCookie(w, &accessTokenCookie)
				http.SetCookie(w, &refreshTokenCookie)
			}

			userId := simpleutils.UintToString(dbUser.ID)
			groupService := services.NewGroupService(repositories.GetDB())
			groups, err := groupService.GetGroupsForUser(userId)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			clientData["groups"] = groups

			// TODO: update frontend to use clientData
			bytes, err := utils.MarshalResponseData(userData)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			w.WriteHeader(200)
			w.Write(bytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}
