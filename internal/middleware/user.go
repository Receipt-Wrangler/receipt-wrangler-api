package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/commands"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

func SetUserData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		errMsg := "Error updating user."
		// TODO: Come up with a better way to handdle this
		var user commands.SignUpCommand
		bodyData, err := utils.GetBodyData(w, r)

		if err != nil {
			utils.WriteErrorResponse(w, err, 500)
			return
		}

		marshalErr := json.Unmarshal(bodyData, &user)
		if marshalErr != nil {
			middleware_logger.Print(marshalErr.Error())
			utils.WriteCustomErrorResponse(w, errMsg, 500)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
		return

	})
}

func SetResetPasswordData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errMsg := "Error updating user."
		// TODO: Come up with a better way to handdle this
		var resetPasswordData structs.ResetPasswordCommand
		bodyData, err := utils.GetBodyData(w, r)

		if err != nil {
			middleware_logger.Print(err.Error())
			utils.WriteErrorResponse(w, err, 500)
			return
		}

		marshalErr := json.Unmarshal(bodyData, &resetPasswordData)
		if marshalErr != nil {
			middleware_logger.Print(marshalErr.Error())
			utils.WriteCustomErrorResponse(w, errMsg, 500)
			return
		}

		ctx := context.WithValue(r.Context(), "reset_password", resetPasswordData)
		next.ServeHTTP(w, r.WithContext(ctx))
		return
	})
}

// TODO: REFACTOR
func ValidateUserData(roleRequired bool) (mw func(http.Handler) http.Handler) {

	mw = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			db := db.GetDB()
			userData := r.Context().Value("user").(commands.SignUpCommand)
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

			if len(userData.Password) == 0 && !userData.IsDummyUser {
				err.Errors["password"] = "Password is required"
			}

			if len(userData.Displayname) == 0 {
				err.Errors["displayName"] = "Displayname is required"
			}

			if roleRequired {
				if len(userData.UserRole) == 0 {
					err.Errors["userRole"] = "User Role is required"
				}
			}

			if len(err.Errors) > 0 {
				utils.WriteValidatorErrorResponse(w, err, http.StatusBadRequest)
				return
			}

			h.ServeHTTP(w, r)
			return
		})
	}
	return
}
