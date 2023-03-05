package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/constants"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

func Login(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Invalid credentials",
		Writer:       w,
		Request:      r,
		ResponseType: "",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			db := db.GetDB()
			userData := r.Context().Value("user").(models.User)
			validatorErrors := validateLoginData(userData)
			var dbUser models.User

			if len(validatorErrors.Errors) > 0 {
				return http.StatusBadRequest, nil
			}

			err := db.Model(models.User{}).Where("username = ?", userData.Username).First(&dbUser).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(userData.Password))
			if err != nil {
				return http.StatusInternalServerError, err
			}

			jwt, refreshToken, err := utils.GenerateJWT(dbUser.ID)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			accessTokenCookie := http.Cookie{Name: constants.JWT_KEY, Value: jwt, HttpOnly: false, Path: "/"}
			refreshTokenCookie := http.Cookie{Name: constants.REFRESH_TOKEN_KEY, Value: refreshToken, HttpOnly: true, Path: "/"}

			http.SetCookie(w, &accessTokenCookie)
			http.SetCookie(w, &refreshTokenCookie)

			w.WriteHeader(200)

			return 0, nil
		},
	}

	HandleRequest(handler)
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
