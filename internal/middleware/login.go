package middleware

import (
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

func ValidateLoginData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userData := r.Context().Value("user").(models.User)
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
			middleware_logger.Print(err.Errors, r)
			utils.WriteValidatorErrorResponse(w, err, http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}
