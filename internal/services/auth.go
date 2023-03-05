package services

import (
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"

	"golang.org/x/crypto/bcrypt"
)

func LoginUser(loginAttempt models.User) (models.User, error) {
	db := db.GetDB()
	var dbUser models.User

	err := db.Model(models.User{}).Where("username = ?", loginAttempt.Username).First(&dbUser).Error
	if err != nil {
		return models.User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(loginAttempt.Password))
	if err != nil {
		return models.User{}, err
	}

	return dbUser, nil
}
