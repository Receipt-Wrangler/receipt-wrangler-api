package middleware

import (
	"context"
	"net/http"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
)

func ValidateRefreshToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenValidator, err := utils.InitTokenValidator()
		errMessage := "Error refreshing token"

		if err != nil {
			middleware_logger.Fatal(err.Error())
			return
		}

		refreshTokenCookie, err := r.Cookie("refresh_token")
		if err != nil {
			utils.WriteCustomErrorResponse(w, errMessage, 500)
			middleware_logger.Println("Refresh token cookie not found")
			return
		}

		refreshToken, err := tokenValidator.ValidateToken(context.TODO(), refreshTokenCookie.Value)
		if err != nil {
			utils.WriteCustomErrorResponse(w, errMessage, 500)
			middleware_logger.Println(err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), "refreshToken", refreshToken)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RevokeRefreshToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := db.GetDB()
		dbToken := models.RefreshToken{}
		errMessage := "Error refreshing token"

		refreshTokenCookie, err := r.Cookie("refresh_token")
		if err != nil {
			utils.WriteCustomErrorResponse(w, errMessage, 500)
			middleware_logger.Println("Refresh token cookie not found")
			return
		}

		err = db.Model(&models.RefreshToken{}).Where("token = ?", refreshTokenCookie.Value).Find(&dbToken).Error
		if err != nil {
			utils.WriteCustomErrorResponse(w, errMessage, 500)
			middleware_logger.Println(err.Error())
			return
		}

		// TODO: Fix
		if dbToken.IsUsed {
			utils.WriteCustomErrorResponse(w, errMessage, 500)
			middleware_logger.Println("Refresh token has been used already.", r, dbToken)
			return
		} else {
			err = db.Model(&dbToken).Update("is_used", true).Error
			if err != nil {
				utils.WriteCustomErrorResponse(w, errMessage, 500)
				middleware_logger.Println(err.Error())
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
