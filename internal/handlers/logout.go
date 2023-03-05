package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/structs"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error logging out.",
		Writer:       w,
		Request:      r,
		ResponseType: "",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			accessTokenCookie := http.Cookie{Name: constants.JWT_KEY, Value: "", HttpOnly: false, Path: "/", MaxAge: -1}
			refreshTokenCookie := http.Cookie{Name: constants.REFRESH_TOKEN_KEY, Value: "", HttpOnly: true, Path: "/", MaxAge: -1}

			http.SetCookie(w, &accessTokenCookie)
			http.SetCookie(w, &refreshTokenCookie)

			w.WriteHeader(200)

			return 0, nil
		},
	}

	HandleRequest(handler)

}
