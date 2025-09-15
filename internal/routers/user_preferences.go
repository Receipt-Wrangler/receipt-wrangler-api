package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func BuildUserPreferencesRouter() *chi.Mux {
	userPreferencesRouter := chi.NewRouter()
	userPreferencesRouter.Use(middleware.UnifiedAuthMiddleware)

	userPreferencesRouter.Get("/", handlers.GetUserPreferences)
	userPreferencesRouter.Put("/", handlers.UpdateUserPreferences)

	return userPreferencesRouter
}
