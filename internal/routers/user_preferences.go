package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildUserPreferencesRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	userPreferencesRouter := chi.NewRouter()
	userPreferencesRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)

	userPreferencesRouter.Get("/", handlers.GetUserPreferences)
	userPreferencesRouter.Put("/", handlers.UpdateUserPreferences)

	return userPreferencesRouter
}
