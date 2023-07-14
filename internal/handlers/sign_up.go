package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/commands"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

func SignUp(w http.ResponseWriter, r *http.Request) {
	errMsg := "Error signing up"
	userData := r.Context().Value("user").(commands.SignUpCommand)
	_, err := repositories.CreateUser(userData)

	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
}

func validateSignUpData(userData models.User) structs.ValidatorError {
	db := db.GetDB()
	err := structs.ValidatorError{
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
