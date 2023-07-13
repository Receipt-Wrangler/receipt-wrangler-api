package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildLoginRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	loginRouter := chi.NewRouter()

	// swagger:route POST /login/ Auth auth
	//
	// Login
	//
	// This will log a user into the system
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
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
	loginRouter.With(middleware.SetBodyData, middleware.ValidateLoginData).Post("/", handlers.Login)

	return loginRouter
}
