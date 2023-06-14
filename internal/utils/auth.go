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
	"golang.org/x/crypto/bcrypt"
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

func GenerateJWT(userId uint) (string, string, Claims, error) {
	db := db.GetDB()
	config := config.GetConfig()
	var user models.User

	err := db.Model(models.User{}).Where("id = ?", userId).First(&user).Error
	if err != nil {
		return "", "", Claims{}, err
	}

	accessTokenClaims := Claims{
		DefaultAvatarColor: user.DefaultAvatarColor,
		Displayname:        user.DisplayName,
		UserId:             user.ID,
		Username:           user.Username,
		UserRole:           user.UserRole,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://receiptWrangler.io",
			Audience:  []string{"https://receiptWrangler.io"},
			ExpiresAt: GetAccessTokenExpiryDate(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, accessTokenClaims)
	signedString, err := accessToken.SignedString([]byte(config.SecretKey))

	if err != nil {
		return "", "", Claims{}, err
	}

	refreshTokenClaims := Claims{
		UserId: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://receiptWrangler.io",
			Audience:  []string{"https://receiptWrangler.io"},
			ExpiresAt: GetRefreshTokenExpiryDate(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(config.SecretKey))

	if err != nil {
		return "", "", Claims{}, err
	}

	token := models.RefreshToken{
		UserId: user.ID,
		Token:  refreshTokenString,
		IsUsed: false,
	}

	err = db.Model(&models.RefreshToken{}).Create(&token).Error

	if err != nil {
		return "", "", Claims{}, err
	}

	return signedString, refreshTokenString, accessTokenClaims, nil
}

func GetJWT(r *http.Request) *Claims {
	return r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims).CustomClaims.(*Claims)
}

func GetRefreshTokenExpiryDate() *jwt.NumericDate {
	return jwt.NewNumericDate(time.Now().Add(24 * time.Hour))
}

func GetAccessTokenExpiryDate() *jwt.NumericDate {
	return jwt.NewNumericDate(time.Now().Add(20 * time.Minute))
}

func HashPassword(password string) ([]byte, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
