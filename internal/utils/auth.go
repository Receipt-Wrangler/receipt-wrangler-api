package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) ([]byte, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func GetRefreshTokenExpiryDate() *jwt.NumericDate {
	return jwt.NewNumericDate(time.Now().Add(24 * time.Hour))
}

func GetAccessTokenExpiryDate() *jwt.NumericDate {
	return jwt.NewNumericDate(time.Now().Add(20 * time.Minute))
}
