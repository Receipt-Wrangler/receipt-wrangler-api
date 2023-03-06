package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/utils"

	"github.com/auth0/go-jwt-middleware/v2/validator"
)

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	oldRefreshToken := r.Context().Value("refreshToken").(*validator.ValidatedClaims).CustomClaims.(*utils.Claims)

	jwt, refreshToken, accessTokenClaims, err := utils.GenerateJWT(oldRefreshToken.UserId)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteErrorResponse(w, err, 500)
		return
	}

	services.PrepareAccessTokenClaims(accessTokenClaims)
	bytes, err := utils.MarshalResponseData(accessTokenClaims)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteErrorResponse(w, err, 500)
		return
	}

	accessTokenCookie, refreshTokenCookie := services.BuildTokenCookies(jwt, refreshToken)

	http.SetCookie(w, &accessTokenCookie)
	http.SetCookie(w, &refreshTokenCookie)

	w.WriteHeader(200)
	w.Write(bytes)
}
