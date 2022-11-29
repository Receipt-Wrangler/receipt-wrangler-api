package utils

import (
	"context"
	"net/http"
	db "receipt-wrangler/api/internal/database"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/golang-jwt/jwt/v4"
)

func InitTokenValidator() (*validator.Validator, error) {
	keyFunc := func(ctx context.Context) (interface{}, error) {
		config := config.GetConfig()
		return []byte(config.SecretKey), nil
	}
	jwtValidator, err := validator.New(
		keyFunc,
		validator.HS512,
		"https://receiptWrangler.io",
		[]string{"https://receiptWrangler.io"},
		validator.WithCustomClaims(customClaims),
		validator.WithAllowedClockSkew(30*time.Second),
	)

	return jwtValidator, err
}

func customClaims() validator.CustomClaims {
	return &Claims{}
}

func GenerateJWT(userId uint) (string, string, error) {
	db := db.GetDB()
	config := config.GetConfig()
	var user models.User

	err := db.Model(models.User{}).Where("id = ?", userId).First(&user).Error
	if err != nil {
		return "", "", err
	}

	claims := &Claims{
		UserId:      user.ID,
		Displayname: user.DisplayName,
		Username:    user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://receiptWrangler.io",
			Audience:  []string{"https://receiptWrangler.io"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signedString, err := accessToken.SignedString([]byte(config.SecretKey))

	if err != nil {
		return "", "", err
	}

	refreshTokenClaims := Claims{
		UserId: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://receiptWrangler.io",
			Audience:  []string{"https://receiptWrangler.io"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(config.SecretKey))

	if err != nil {
		return "", "", err
	}

	token := models.RefreshToken{
		UserId: user.ID,
		Token:  refreshTokenString,
		IsUsed: false,
	}

	err = db.Model(&models.RefreshToken{}).Create(&token).Error

	if err != nil {
		return "", "", err
	}

	return signedString, refreshTokenString, nil
}

func GetJWT(r *http.Request) *Claims {
	return r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims).CustomClaims.(*Claims)
}
