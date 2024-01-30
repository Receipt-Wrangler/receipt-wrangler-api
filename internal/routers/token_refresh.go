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
	tokenRefreshRouter.Post("/", handlers.RefreshToken)

	return tokenRefreshRouter
}
