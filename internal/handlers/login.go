package handlers

import (
	"encoding/json"
	"net/http"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

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
		handler_logger.Print(validatorErrors)
		utils.WriteValidatorErrorResponse(w, validatorErrors, 400)
		return
	}

	err := db.Model(models.User{}).Where("username = ?", userData.Username).First(&dbUser).Error
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(userData.Password))
	if err != nil {
		handler_logger.Print(err.Error(), r)
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	jwt, refreshToken, err := utils.GenerateJWT(dbUser.ID)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteErrorResponse(w, err, 500)
		return
	}

	responseBytes, err := json.Marshal(responseData)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteErrorResponse(w, err, 500)
		return
	}

	accessTokenCookie := http.Cookie{Name: "jwt", Value: jwt, HttpOnly: false, Path: "/"}
	refreshTokenCookie := http.Cookie{Name: "refresh_token", Value: refreshToken, HttpOnly: true, Path: "/"}

	http.SetCookie(w, &accessTokenCookie)
	http.SetCookie(w, &refreshTokenCookie)

	w.WriteHeader(200)
	w.Write(responseBytes)
}

func validateLoginData(userData models.User) structs.ValidatorError {
	err := structs.ValidatorError{
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
