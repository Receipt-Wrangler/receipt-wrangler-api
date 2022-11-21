package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/constants"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	accessTokenCookie := http.Cookie{Name: constants.JWT_KEY, Value: "", HttpOnly: false, Path: "/", MaxAge: -1}
	refreshTokenCookie := http.Cookie{Name: constants.REFRESH_TOKEN_KEY, Value: "", HttpOnly: true, Path: "/", MaxAge: -1}

	http.SetCookie(w, &accessTokenCookie)
	http.SetCookie(w, &refreshTokenCookie)

	w.WriteHeader(200)
}
