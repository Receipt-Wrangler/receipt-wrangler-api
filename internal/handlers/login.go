package handlers

import (
	"errors"
	"net/http"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

func Login(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Invalid credentials",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			userData := r.Context().Value("user").(models.User)
			var dbUser models.User

			dbUser, err := services.LoginUser(userData)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			if dbUser.IsDummyUser {
				return http.StatusInternalServerError, errors.New("dummy users cannot log in")
			}

			jwt, refreshToken, accessTokenClaims, err := utils.GenerateJWT(dbUser.ID)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			services.PrepareAccessTokenClaims(accessTokenClaims)
			bytes, err := utils.MarshalResponseData(accessTokenClaims)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			accessTokenCookie, refreshTokenCookie := services.BuildTokenCookies(jwt, refreshToken)

			http.SetCookie(w, &accessTokenCookie)
			http.SetCookie(w, &refreshTokenCookie)

			w.WriteHeader(200)
			w.Write(bytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}
