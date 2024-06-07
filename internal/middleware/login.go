package middleware

import (
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/structs"
)

func ValidateLoginData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userData := r.Context().Value("user").(commands.LoginCommand)
		err := structs.ValidatorError{
			Errors: make(map[string]string),
		}

		if len(userData.Username) == 0 {
			err.Errors["username"] = "Username is required"
		}

		if len(userData.Password) == 0 {
			err.Errors["password"] = "Password is required"
		}

		if len(err.Errors) > 0 {
			logging.LogStd(logging.LOG_LEVEL_ERROR, "Invalid login data")
			structs.WriteValidatorErrorResponse(w, err, http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}
