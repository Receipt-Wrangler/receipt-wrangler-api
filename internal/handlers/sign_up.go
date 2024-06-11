package handlers

import (
	"errors"
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
)

func SignUp(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error signing up.",
		Writer:       w,
		Request:      r,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			systemSettingsService := services.NewSystemSettingsService(nil)
			featureConfig, err := systemSettingsService.GetFeatureConfig()
			if err != nil {
				return http.StatusInternalServerError, err
			}

			if !featureConfig.EnableLocalSignUp {
				return http.StatusNotFound, errors.New("Local sign up is disabled")
			}

			userData := r.Context().Value("signUpCommand").(commands.SignUpCommand)
			userRepository := repositories.NewUserRepository(nil)
			_, err = userRepository.CreateUser(userData)

			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
			return 0, nil
		},
	}

	HandleRequest(handler)

}

func validateSignUpData(userData models.User) structs.ValidatorError {
	db := repositories.GetDB()
	err := structs.ValidatorError{
		Errors: make(map[string]string),
	}

	if len(userData.Username) == 0 {
		err.Errors["username"] = "Username is required"
	} else {
		var count int64
		db.Model(&models.User{}).Where("username = ?", userData.Username).Count(&count)

		if count > 0 {
			err.Errors["username"] = "Username must be unique"
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
