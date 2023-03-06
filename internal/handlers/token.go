package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	"github.com/auth0/go-jwt-middleware/v2/validator"
)

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error refreshing token",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {

			oldRefreshToken := r.Context().Value("refreshToken").(*validator.ValidatedClaims).CustomClaims.(*utils.Claims)

			jwt, refreshToken, accessTokenClaims, err := utils.GenerateJWT(oldRefreshToken.UserId)
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
