package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error logging out.",
		Writer:       w,
		Request:      r,
		ResponseType: "",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			accessTokenCookie := services.GetEmptyAccessTokenCookie()
			refreshTokenCookie := services.GetEmptyRefreshTokenCookie()

			http.SetCookie(w, &accessTokenCookie)
			http.SetCookie(w, &refreshTokenCookie)

			w.WriteHeader(200)

			return 0, nil
		},
	}

	HandleRequest(handler)

}
