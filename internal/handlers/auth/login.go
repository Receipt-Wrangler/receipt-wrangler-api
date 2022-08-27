package auth

import (
	"encoding/json"
	"net/http"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/models"
	httpUtils "receipt-wrangler/api/internal/utils/http"

	"golang.org/x/crypto/bcrypt"
)

func Login(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	userData := r.Context().Value("user").(models.User)
	validatorErrors := validateLoginData(userData)
	errMsg := "Either the user doesn't exist, or the password is incorrect"
	var responseData = make(map[string]string)
	var dbUser models.User

	if len(validatorErrors.Errors) > 0 {
		httpUtils.WriteValidatorErrorResponse(w, validatorErrors, 500)
		return
	}

	err := db.Model(models.User{}).Where("username = ?", userData.Username).First(&dbUser).Error
	if err != nil {
		httpUtils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(userData.Password))
	if err != nil {
		httpUtils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	jwt, err := generateJWT(userData.Username)
	if err != nil {
		httpUtils.WriteErrorResponse(w, err, 500)
		return
	}

	responseBytes, err := json.Marshal(responseData)
	if err != nil {
		httpUtils.WriteErrorResponse(w, err, 500)
		return
	}

	cookie := http.Cookie{Name: "jwt", Value: jwt, HttpOnly: false, Path: "/"}

	http.SetCookie(w, &cookie)
	w.WriteHeader(200)
	w.Write(responseBytes)
}

func validateLoginData(userData models.User) handlers.ValidatorError {
	err := handlers.ValidatorError{
		Errors: make(map[string]string),
	}

	if len(userData.Username) == 0 {
		err.Errors["username"] = "Username is required"
	}

	if len(userData.Password) == 0 {
		err.Errors["password"] = "Password is required"
	}

	return err
}
