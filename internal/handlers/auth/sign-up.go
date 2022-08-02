package signUp

import (
	"fmt"
	"net/http"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	httpUtils "receipt-wrangler/api/internal/utils/http"
)

func SignUp(w http.ResponseWriter, r *http.Request) {
	validateData()

	db := db.GetDB()

	userData := r.Context().Value("user").(models.User)
	fmt.Println(userData)

	// hash password

	result := db.Create(&userData)

	if result.Error != nil {
		httpUtils.WriteErrorResponse(w, result.Error, 500)
		return
	}

	w.Write([]byte("hello there"))
}

func validateData() {

}
