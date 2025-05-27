package handlers

import (
	"errors"
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

func Login(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Invalid credentials.",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			userData := r.Context().Value("user").(commands.LoginCommand)
			var dbUser models.User

			dbUser, firstAdminToLogin, err := services.LoginUser(userData)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			if firstAdminToLogin {
				promptService := services.NewPromptService(nil)
				_, err = promptService.CreateDefaultPrompt()
				if err != nil {
					logging.LogStd(logging.LOG_LEVEL_INFO, err)
				}
			}

			if dbUser.IsDummyUser {
				return http.StatusInternalServerError, errors.New("dummy users cannot log in")
			}

			jwt, refreshToken, accessTokenClaims, err := services.GenerateJWT(dbUser.ID)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			services.PrepareAccessTokenClaims(accessTokenClaims)

			appData, err := services.GetAppData(dbUser.ID, nil)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			tokensInBodyOnly := r.URL.Query().Get("tokensInBody") == "true"

			if utils.IsMobileApp(r) || tokensInBodyOnly {
				appData.Jwt = jwt
				appData.RefreshToken = refreshToken
			}

			if !tokensInBodyOnly {
				accessTokenCookie, refreshTokenCookie := services.BuildTokenCookies(jwt, refreshToken)

				http.SetCookie(w, &accessTokenCookie)
				http.SetCookie(w, &refreshTokenCookie)
			}

			appData.Claims = accessTokenClaims

			bytes, err := utils.MarshalResponseData(appData)
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
