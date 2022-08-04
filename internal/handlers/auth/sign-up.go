package signUp

import (
	"net/http"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	httpUtils "receipt-wrangler/api/internal/utils/http"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	jwt.RegisteredClaims
}

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

	tokenString, err := generateJWT()
	if err != nil {
		httpUtils.WriteErrorResponse(w, err, 500)
		return
	}

	w.Write([]byte(tokenString))
}

func generateJWT() (string, error) {
	claims := &Claims{}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signedString, err := token.SignedString([]byte("this is a placeholder for secret"))

	return signedString, err
}

func validateData() {
}
