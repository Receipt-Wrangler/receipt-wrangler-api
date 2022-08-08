package signUp

import (
	"net/http"
	db "receipt-wrangler/api/internal/database"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/models"
	httpUtils "receipt-wrangler/api/internal/utils/http"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	jwt.RegisteredClaims
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	userData := r.Context().Value("user").(models.User)
	validatorErrors := validateData(userData)

	if len(validatorErrors.Errors) > 0 {
		httpUtils.WriteValidatorErrorResponse(w, validatorErrors, 500)
		return
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(userData.Password), 14)
	if err != nil {
		httpUtils.WriteErrorResponse(w, err, 500)
	}

	userData.Password = string(bytes)
	result := db.Create(&userData)

	if result.Error != nil {
		httpUtils.WriteErrorResponse(w, result.Error, 500)
		return
	}

	tokenString, err := generateJWT()
	if err != nil {
		httpUtils.WriteErrorResponse(w, err, 500)
		return
	}

	w.Write([]byte(tokenString))
}

func generateJWT() (string, error) {
	config := config.GetConfig()
	claims := &Claims{}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signedString, err := token.SignedString([]byte(config.SecretKey))

	return signedString, err
}

func validateData(userData models.User) handlers.ValidatorError {
	db := db.GetDB()
	err := handlers.ValidatorError{
		Errors: make(map[string]string),
	}

	if len(userData.Username) == 0 {
		err.Errors["username"] = "Username is required"
	} else {
		var count int64
		db.Model(&models.User{}).Where("username = ?", userData.Username).Count(&count)

		if count > 0 {
			err.Errors["username"] = "Username already exists"
		}
	}

	if len(userData.Password) == 0 {
		err.Errors["password"] = "Password is required"
	}

	if len(userData.DisplayName) == 0 {
		err.Errors["displayName"] = "Displayname is required"
	}

	return err
}
