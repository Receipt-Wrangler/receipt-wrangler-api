package routers

import (
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
)

func BuildSystemSettingsRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	systemSettingsRouter := chi.NewRouter()

	systemSettingsRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)
	systemSettingsRouter.Get("/", handlers.GetSystemSettings)
	systemSettingsRouter.Put("/", handlers.UpdateSystemSettings)

	return systemSettingsRouter
}
