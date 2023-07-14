package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildTokenRefreshRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	tokenRefreshRouter := chi.NewRouter()

	tokenRefreshRouter.Use(middleware.ValidateRefreshToken, middleware.RevokeRefreshToken)

	// swagger:route POST /token/ Auth auth
	//
	// Get fresh tokens
	//
	// This will get a fresh token pair for the user
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
	tokenRefreshRouter.Post("/", handlers.RefreshToken)

	return tokenRefreshRouter
}
