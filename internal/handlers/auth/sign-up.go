package signUp

import (
	"net/http"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	httpUtils "receipt-wrangler/api/internal/utils/http"

	"golang.org/x/crypto/bcrypt"
)

func SignUp(w http.ResponseWriter, r *http.Request) {
	validateData()

	db := db.GetDB()

	userData := r.Context().Value("user").(models.User)

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

}

func validateData() {
}
