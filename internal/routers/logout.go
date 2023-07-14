package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildLogoutRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	logoutRouter := chi.NewRouter()

	// swagger:route POST /logout/ Auth auth
	//
	// Logout
	//
	// This will log a user out of the system and revoke their token [SYSTEM USER]
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
	logoutRouter.With(middleware.RevokeRefreshToken).Post("/", handlers.Logout)

	return logoutRouter
}
