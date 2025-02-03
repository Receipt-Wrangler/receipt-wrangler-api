package middleware

import (
	"context"
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/utils"
)

func ValidateRefreshToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenValidator, err := services.InitTokenValidator()
		errMessage := "Error refreshing token"

		if err != nil {
			logging.LogStd(logging.LOG_LEVEL_FATAL, err.Error())
			return
		}

		refreshTokenString, err := getRefreshTokenFromRequest(r, w)
		if err != nil {
			utils.WriteCustomErrorResponse(w, errMessage, http.StatusInternalServerError)
			logging.LogStd(logging.LOG_LEVEL_ERROR, "Refresh token not found")
			return
		}

		refreshToken, err := tokenValidator.ValidateToken(context.TODO(), refreshTokenString)
		if err != nil {
			utils.WriteCustomErrorResponse(w, errMessage, http.StatusInternalServerError)
			logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), "refreshToken", refreshToken)
		ctx = context.WithValue(ctx, "refreshTokenString", refreshTokenString)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RevokeRefreshToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := repositories.GetDB()
		dbToken := models.RefreshToken{}
		err := error(nil)
		errMessage := "Error refreshing token"

		refreshTokenString := r.Context().Value("refreshTokenString")
		if refreshTokenString == nil {
			refreshTokenString, err = getRefreshTokenFromRequest(r, w)
			if err != nil {
				utils.WriteCustomErrorResponse(w, errMessage, http.StatusInternalServerError)
				logging.LogStd(logging.LOG_LEVEL_ERROR, "Refresh token not found")
				return
			}
		}

		hashTokenString := utils.Sha256Hash([]byte(refreshTokenString.(string)))
		err = db.Model(&models.RefreshToken{}).Where("token = ?", hashTokenString).Find(&dbToken).Error
		if err != nil {
			utils.WriteCustomErrorResponse(w, errMessage, http.StatusInternalServerError)
			logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
			return
		}

		if dbToken.IsUsed {
			emptyAccessTokenCookie := services.GetEmptyAccessTokenCookie()
			emptyRefreshTokenCookie := services.GetEmptyRefreshTokenCookie()

			http.SetCookie(w, &emptyAccessTokenCookie)
			http.SetCookie(w, &emptyRefreshTokenCookie)

			utils.WriteCustomErrorResponse(w, errMessage, http.StatusInternalServerError)
			logging.LogStd(logging.LOG_LEVEL_ERROR, "Refresh token has been used already.")

			return
		} else {
			err = db.Model(&dbToken).Update("is_used", true).Error
			if err != nil {
				utils.WriteCustomErrorResponse(w, errMessage, http.StatusInternalServerError)
				logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func getRefreshTokenFromRequest(r *http.Request, w http.ResponseWriter) (string, error) {
	if utils.IsMobileApp(r) {
		var command commands.LogoutCommand
		err := command.LoadDataFromRequest(w, r)
		if err != nil {
			return "", err
		}

		return command.RefreshToken, nil
	} else {
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			return "", err
		}

		return cookie.Value, nil
	}
}
