package structs

import (
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

func GetJWT(r *http.Request) *Claims {
	return r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims).CustomClaims.(*Claims)
}
