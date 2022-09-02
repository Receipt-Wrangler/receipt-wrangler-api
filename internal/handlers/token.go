package handlers

import (
	"net/http"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"

	"github.com/auth0/go-jwt-middleware/v2/validator"
)

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	// if we get here then we have teh jwt in token
	// next we want to write a custom middle ware as  a wrapper for the jwt validator
	// and validate the refresh token, and custom claims
	// at this point we can generate a new set, and return them as cookies
	// et voiala
	oldRefreshToken := r.Context().Value("refreshToken").(*validator.ValidatedClaims)
	db := db.GetDB()
	var dbUser models.User

	err := db.Model(models.User{}).Where("id = ?", oldRefreshToken.RegisteredClaims.Subject).First(&dbUser).Error
	if err != nil {
		utils.WriteCustomErrorResponse(w, "Error refreshing token", 500)
		return
	}

	jwt, refreshToken, err := utils.GenerateJWT(dbUser.Username)
	if err != nil {
		utils.WriteErrorResponse(w, err, 500)
		return
	}

	accessTokenCookie := http.Cookie{Name: "jwt", Value: jwt, HttpOnly: false, Path: "/"}
	refreshTokenCookie := http.Cookie{Name: "refresh_token", Value: refreshToken, HttpOnly: true, Path: "/"}

	http.SetCookie(w, &accessTokenCookie)
	http.SetCookie(w, &refreshTokenCookie)

	w.WriteHeader(200)
}
