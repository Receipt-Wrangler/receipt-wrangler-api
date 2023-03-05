package handlers

import (
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
		ResponseType: "",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			userData := r.Context().Value("user").(models.User)
			var dbUser models.User

			dbUser, err := services.LoginUser(userData)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			jwt, refreshToken, err := utils.GenerateJWT(dbUser.ID)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			accessTokenCookie := http.Cookie{Name: constants.JWT_KEY, Value: jwt, HttpOnly: false, Path: "/"}
			refreshTokenCookie := http.Cookie{Name: constants.REFRESH_TOKEN_KEY, Value: refreshToken, HttpOnly: true, Path: "/"}

			http.SetCookie(w, &accessTokenCookie)
			http.SetCookie(w, &refreshTokenCookie)

			w.WriteHeader(200)

			return 0, nil
		},
	}

	HandleRequest(handler)
}
