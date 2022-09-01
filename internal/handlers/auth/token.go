package auth

import (
	"net/http"
	auth_utils "receipt-wrangler/api/internal/utils/auth"
	httpUtils "receipt-wrangler/api/internal/utils/http"
)

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	// if we get here then we have teh jwt in token
	// next we want to write a custom middle ware as  a wrapper for the jwt validator
	// and validate the refresh token, and custom claims
	// at this point we can generate a new set, and return them as cookies
	// et voiala
	oldJwt := auth_utils.GetJWT(r)

	jwt, refreshToken, err := auth_utils.GenerateJWT(oldJwt.Username)
	if err != nil {
		httpUtils.WriteErrorResponse(w, err, 500)
		return
	}

	accessTokenCookie := http.Cookie{Name: "jwt", Value: jwt, HttpOnly: false, Path: "/"}
	refreshTokenCookie := http.Cookie{Name: "refresh_token", Value: refreshToken, HttpOnly: true, Path: "/"}

	http.SetCookie(w, &accessTokenCookie)
	http.SetCookie(w, &refreshTokenCookie)

	w.WriteHeader(200)
}
