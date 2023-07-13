package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildSignUpRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	signUpRouter := chi.NewRouter()
	signUpRouter.Use(middleware.SetBodyData, middleware.ValidateUserData(false))

	// swagger:route POST /signUp/ Auth auth
	//
	// Registers a user
	//
	// This will log a user into the system
	//
	//     Consumes:
	//     - application/json
	//
	//     Schemes: https
	//
	//     Deprecated: false
	//
	//     Security:
	//       api_key:
	//
	//     Responses:
	//       200: Ok
	//       500: Internal Server Error
	signUpRouter.Post("/", handlers.SignUp)

	return signUpRouter
}
