package auth

import (
	"net/http"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

func SignUp(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	userData := r.Context().Value("user").(models.User)
	validatorErrors := validateSignUpData(userData)

	if len(validatorErrors.Errors) > 0 {
		utils.WriteValidatorErrorResponse(w, validatorErrors, 500)
		return
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(userData.Password), 14)
	if err != nil {
		utils.WriteErrorResponse(w, err, 500)
	}

	userData.Password = string(bytes)
	result := db.Create(&userData)

	if result.Error != nil {
		utils.WriteErrorResponse(w, result.Error, 500)
		return
	}

	w.WriteHeader(200)
}

func validateSignUpData(userData models.User) handlers.ValidatorError {
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
