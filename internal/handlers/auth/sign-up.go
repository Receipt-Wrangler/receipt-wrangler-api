package signUp

import (
	"fmt"
	"net/http"
	"receipt-wrangler/api/internal/models"
)

func SignUp(w http.ResponseWriter, r *http.Request) {
	validateData()
	userData := r.Context().Value("user").(models.User)
	fmt.Println(userData)
	w.Write([]byte("hello there"))
}

func validateData() {

}
