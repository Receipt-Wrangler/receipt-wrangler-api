package routers

import (
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
)

func BuildApiKeyRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	apiKeyRouter := chi.NewRouter()

	// Authenticated routes
	apiKeyRouter.With(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT).Post("/", handlers.CreateApiKey)

	return apiKeyRouter
}
