package routers

import (
	"github.com/go-chi/chi/v5"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
)

func BuildSystemSettingsRouter() *chi.Mux {
	systemSettingsRouter := chi.NewRouter()

	systemSettingsRouter.Use(middleware.UnifiedAuthMiddleware)
	systemSettingsRouter.Get("/", handlers.GetSystemSettings)
	systemSettingsRouter.Put("/", handlers.UpdateSystemSettings)
	systemSettingsRouter.Post("/restartTaskServer", handlers.RestartTaskServer)

	return systemSettingsRouter
}
